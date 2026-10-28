package sms

import (
	stderrors "errors"

	kratoserrors "github.com/go-kratos/kratos/v3/errors"
)

// Sentinel errors are the backend-agnostic sms contract, mapped 1:1 to the
// InfraError 36xx Sms codes; the service layer adapts them to application
// errors. The retry posture rides on the sentinel:
// InvalidArgument and ChannelRejected never retry; Unavailable may; Timeout
// may but the message's fate is unknown — the caller's rate limiter is the
// throttle, never the backend's silent retry.
var (
	ErrInvalidArgument = stderrors.New("sms: invalid argument")
	ErrChannelRejected = stderrors.New("sms: channel rejected")
	ErrUnavailable     = stderrors.New("sms: unavailable")
	ErrTimeout         = stderrors.New("sms: timeout")
)

// ChannelError is the typed detail a channel business rejection carries; it
// unwraps to ErrChannelRejected, so errors.Is classifies while errors.As
// exposes the channel's own code/message (isv.BUSINESS_LIMIT_CONTROL,
// isv.MOBILE_NUMBER_ILLEGAL, ...) for consumer-side handling and audit.
// Never re-prose these — the channel's vocabulary is the audit trail.
type ChannelError struct {
	Code    string
	Message string
}

func (e *ChannelError) Error() string {
	return "sms: channel rejected: " + e.Code + ": " + e.Message
}

func (e *ChannelError) Unwrap() error { return ErrChannelRejected }

// SmsDefaultError holds the app error instances a service maps the sms
// sentinels to: the mapping mechanism lives here, the concrete instances
// are supplied by each service's platform layer (same mechanism/instance
// split as cache.HandleCacheError), so this package never imports a
// service error proto.
type SmsDefaultError struct {
	InvalidArgument *kratoserrors.Error
	ChannelRejected *kratoserrors.Error
	Unavailable     *kratoserrors.Error
	Timeout         *kratoserrors.Error // optional: nil → timeouts map to Unavailable
}

// ValidateSmsDefaultError fails fast if a required slot is nil — a nil slot
// nil-derefs inside WithCause. Timeout is optional.
func ValidateSmsDefaultError(errs *SmsDefaultError) error {
	for _, e := range []*kratoserrors.Error{
		errs.InvalidArgument, errs.ChannelRejected, errs.Unavailable,
	} {
		if e == nil {
			return stderrors.New("sms: SmsDefaultError has a nil sentinel slot")
		}
	}
	return nil
}

// HandleSmsError maps an sms error to the supplied app error, keeping the
// original as cause. Unknown errors land on Unavailable — sms errors never
// leak past the service error domain untranslated.
func HandleSmsError(err error, errs *SmsDefaultError) error {
	if e := ValidateSmsDefaultError(errs); e != nil {
		return e
	}
	switch {
	case stderrors.Is(err, ErrInvalidArgument):
		return errs.InvalidArgument.WithCause(err)
	case stderrors.Is(err, ErrChannelRejected):
		return errs.ChannelRejected.WithCause(err)
	case stderrors.Is(err, ErrTimeout):
		// Optional slot; a nil folds onto Unavailable — outcome unknown,
		// the caller's limiter decides whether a resend ever happens.
		if errs.Timeout != nil {
			return errs.Timeout.WithCause(err)
		}
		return errs.Unavailable.WithCause(err)
	default:
		return errs.Unavailable.WithCause(err)
	}
}
