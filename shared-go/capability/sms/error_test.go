package sms

import (
	"errors"
	"testing"

	kratoserrors "github.com/go-kratos/kratos/v3/errors"
)

func fullDefaultError() *SmsDefaultError {
	return &SmsDefaultError{
		InvalidArgument: kratoserrors.New(400, "SMS_INVALID_ARGUMENT", "x"),
		ChannelRejected: kratoserrors.New(422, "SMS_CHANNEL_REJECTED", "x"),
		Unavailable:     kratoserrors.New(503, "SMS_UNAVAILABLE", "x"),
		Timeout:         kratoserrors.New(504, "SMS_TIMEOUT", "x"),
	}
}

// TestHandleSmsError walks every sentinel, the typed channel rejection, the
// unknown-error fallback, and the optional-slot fallback (a nil Timeout slot
// folds timeouts onto Unavailable — the message's fate is unknown and the
// caller's limiter decides any resend).
func TestHandleSmsError(t *testing.T) {
	errs := fullDefaultError()
	channel := &ChannelError{Code: "isv.BUSINESS_LIMIT_CONTROL", Message: "triggered flow control"}
	cases := []struct {
		name string
		in   error
		want int32
	}{
		{"invalid_argument", ErrInvalidArgument, 400},
		{"channel_rejected", ErrChannelRejected, 422},
		{"channel_error_typed", channel, 422},
		{"unavailable", ErrUnavailable, 503},
		{"timeout", ErrTimeout, 504},
		{"unknown_falls_unavailable", errors.New("socket exploded"), 503},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := HandleSmsError(tc.in, errs)
			if code := func() *kratoserrors.Error {
				target := &kratoserrors.Error{}
				_ = errors.As(got, &target)
				return target
			}().Code; code != tc.want {
				t.Fatalf("code = %d, want %d", code, tc.want)
			}
			if !errors.Is(got, tc.in) && tc.in != nil {
				t.Fatalf("original lost as cause: %v", got)
			}
		})
	}
}

// TestHandleSmsErrorNilTimeoutFoldsUnavailable pins the optional-slot
// contract: without a Timeout instance a timeout stays presentable, mapped
// onto Unavailable.
func TestHandleSmsErrorNilTimeoutFoldsUnavailable(t *testing.T) {
	errs := fullDefaultError()
	errs.Timeout = nil
	if err := ValidateSmsDefaultError(errs); err != nil {
		t.Fatalf("nil Timeout is legal: %v", err)
	}
	got := HandleSmsError(ErrTimeout, errs)
	if code := func() *kratoserrors.Error {
		target := &kratoserrors.Error{}
		_ = errors.As(got, &target)
		return target
	}().Code; code != 503 {
		t.Fatalf("code = %d, want 503", code)
	}
}

// TestValidateSmsDefaultError requires every non-optional slot — a nil slot
// would nil-deref inside WithCause, so it fails the boot instead.
func TestValidateSmsDefaultError(t *testing.T) {
	if err := ValidateSmsDefaultError(fullDefaultError()); err != nil {
		t.Fatalf("full error set must validate: %v", err)
	}
	for _, slot := range []struct {
		name string
		mut  func(*SmsDefaultError)
	}{
		{"invalid_argument", func(e *SmsDefaultError) { e.InvalidArgument = nil }},
		{"channel_rejected", func(e *SmsDefaultError) { e.ChannelRejected = nil }},
		{"unavailable", func(e *SmsDefaultError) { e.Unavailable = nil }},
	} {
		t.Run("nil_"+slot.name, func(t *testing.T) {
			errs := fullDefaultError()
			slot.mut(errs)
			if err := ValidateSmsDefaultError(errs); err == nil {
				t.Fatalf("nil %s slot must fail validation", slot.name)
			}
		})
	}
}

// TestChannelErrorUnwrap pins the two consumption modes: errors.Is on the
// sentinel classifies, errors.As exposes the channel's own vocabulary.
func TestChannelErrorUnwrap(t *testing.T) {
	e := &ChannelError{Code: "isv.MOBILE_NUMBER_ILLEGAL", Message: "bad phone"}
	if !errors.Is(e, ErrChannelRejected) {
		t.Fatal("channel error must unwrap to ErrChannelRejected")
	}
	var ce *ChannelError
	if !errors.As(error(e), &ce) || ce.Code != "isv.MOBILE_NUMBER_ILLEGAL" {
		t.Fatal("errors.As must expose the channel code verbatim")
	}
}
