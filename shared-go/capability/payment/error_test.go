package payment

import (
	"errors"
	"testing"

	kratoserrors "github.com/go-kratos/kratos/v3/errors"
)

func fullDefaultError() *PaymentDefaultError {
	return &PaymentDefaultError{
		InvalidArgument:  kratoserrors.New(400, "PAYMENT_INVALID_ARGUMENT", "x"),
		NotFound:         kratoserrors.New(404, "PAYMENT_NOT_FOUND", "x"),
		Duplicate:        kratoserrors.New(409, "PAYMENT_DUPLICATE", "x"),
		ChannelRejected:  kratoserrors.New(422, "PAYMENT_CHANNEL_REJECTED", "x"),
		Unavailable:      kratoserrors.New(503, "PAYMENT_UNAVAILABLE", "x"),
		Timeout:          kratoserrors.New(504, "PAYMENT_TIMEOUT", "x"),
		InvalidSignature: kratoserrors.New(400, "PAYMENT_INVALID_SIGNATURE", "x"),
		DecryptFailed:    kratoserrors.New(400, "PAYMENT_DECRYPT_FAILED", "x"),
	}
}

// TestHandlePaymentError walks every sentinel, the typed channel rejection,
// the unknown-error fallback, and the optional-slot fallbacks (a nil Timeout
// slot folds timeouts onto Unavailable — outcome unknown, converge via query;
// nil signature/decryption slots fold onto InvalidArgument).
func TestHandlePaymentError(t *testing.T) {
	errs := fullDefaultError()
	channel := &ChannelError{Code: "NOTENOUGH", Message: "not enough balance"}
	cases := []struct {
		name string
		in   error
		want int32
	}{
		{"invalid_argument", ErrInvalidArgument, 400},
		{"not_found", ErrNotFound, 404},
		{"duplicate", ErrDuplicate, 409},
		{"channel_rejected", ErrChannelRejected, 422},
		{"channel_error_typed", channel, 422},
		{"unavailable", ErrUnavailable, 503},
		{"timeout", ErrTimeout, 504},
		{"invalid_signature", ErrInvalidSignature, 400},
		{"decrypt_failed", ErrDecryptFailed, 400},
		{"unknown", errors.New("boom"), 503},
	}
	for _, c := range cases {
		got := HandlePaymentError(c.in, errs)
		var ke *kratoserrors.Error
		if !errors.As(got, &ke) || ke.Code != c.want {
			t.Errorf("%s: code=%v, want %d", c.name, ke, c.want)
			continue
		}
		// The mapped error keeps the original as cause.
		if errors.Unwrap(got) == nil {
			t.Errorf("%s: mapped error lost the cause", c.name)
		}
	}
}

// TestHandlePaymentErrorOptionalSlots pins the fold behavior: without a
// Timeout slot timeouts map to Unavailable; without signature/decryption
// slots both map to InvalidArgument.
func TestHandlePaymentErrorOptionalSlots(t *testing.T) {
	errs := fullDefaultError()
	errs.Timeout = nil
	errs.InvalidSignature = nil
	errs.DecryptFailed = nil
	for in, want := range map[error]int32{
		ErrTimeout:          503,
		ErrInvalidSignature: 400,
		ErrDecryptFailed:    400,
	} {
		got := HandlePaymentError(in, errs)
		var ke *kratoserrors.Error
		if !errors.As(got, &ke) || ke.Code != want {
			t.Errorf("optional fold %v: code=%v, want %d", in, ke, want)
		}
	}
}

func TestValidatePaymentDefaultError(t *testing.T) {
	if err := ValidatePaymentDefaultError(&PaymentDefaultError{}); err == nil {
		t.Fatal("want error for all-nil slots")
	}
	// Every optional slot nil still validates: the five required slots are
	// the classification floor.
	errs := fullDefaultError()
	errs.Timeout = nil
	errs.InvalidSignature = nil
	errs.DecryptFailed = nil
	if err := ValidatePaymentDefaultError(errs); err != nil {
		t.Fatalf("optional slots nil should validate: %v", err)
	}
	errs.NotFound = nil
	if err := ValidatePaymentDefaultError(errs); err == nil {
		t.Fatal("want error for a nil required slot")
	}
}

// TestChannelErrorClassification pins the dual read: errors.Is classifies
// onto the sentinel, errors.As exposes the channel's own vocabulary — the
// pair a consumer needs to route the rejection and audit the channel code.
func TestChannelErrorClassification(t *testing.T) {
	channel := &ChannelError{Code: "TRADE_ERROR", Message: "trade closed"}
	if !errors.Is(channel, ErrChannelRejected) {
		t.Error("ChannelError must unwrap onto ErrChannelRejected")
	}
	got := HandlePaymentError(channel, fullDefaultError())
	if !errors.Is(got, ErrChannelRejected) {
		t.Error("mapped error must still classify as channel-rejected")
	}
	var detail *ChannelError
	if !errors.As(got, &detail) || detail.Code != "TRADE_ERROR" {
		t.Error("mapped error must expose the channel code via errors.As")
	}
}

// TestAmountFenString pins the display format: yuan, two-digit fen, sign
// carried — display-only, arithmetic stays on the integer.
func TestAmountFenString(t *testing.T) {
	cases := []struct {
		in   AmountFen
		want string
	}{
		{Fen(0), "0.00"},
		{Fen(5), "0.05"},
		{Fen(100), "1.00"},
		{Fen(1234), "12.34"},
		{Fen(-101), "-1.01"},
		{Fen(999999999), "9999999.99"},
	}
	for _, c := range cases {
		if got := c.in.String(); got != c.want {
			t.Errorf("AmountFen(%d).String(): want %q, got %q", int64(c.in), c.want, got)
		}
	}
}
