package payment

import (
	stderrors "errors"

	kratoserrors "github.com/go-kratos/kratos/v3/errors"
)

// Sentinel errors are the backend-agnostic payment contract, mapped 1:1 to
// the InfraError 35xx Payment codes (infra.proto); the service layer adapts
// them to application errors. The retry posture rides on the sentinel:
// InvalidArgument/Duplicate/NotFound/ChannelRejected never retry, Timeout
// and Unavailable may, and Timeout's outcome is unknown — converge through
// QueryPayment/QueryRefund, never a blind retry of the mutating call.
var (
	ErrInvalidArgument  = stderrors.New("payment: invalid argument")
	ErrNotFound         = stderrors.New("payment: not found")
	ErrDuplicate        = stderrors.New("payment: duplicate trade number with different parameters")
	ErrChannelRejected  = stderrors.New("payment: channel rejected")
	ErrUnavailable      = stderrors.New("payment: unavailable")
	ErrTimeout          = stderrors.New("payment: timeout")
	ErrInvalidSignature = stderrors.New("payment: invalid notification signature")
	ErrDecryptFailed    = stderrors.New("payment: notification decryption failed")
)

// ChannelError is the typed detail a channel business rejection carries; it
// unwraps to ErrChannelRejected, so errors.Is classifies while errors.As
// exposes the channel's own code/message (NOTENOUGH, TRADE_ERROR, ...) for
// consumer-side handling and audit. Never re-prose these — the channel's
// vocabulary is the audit trail.
type ChannelError struct {
	Code    string
	Message string
}

func (e *ChannelError) Error() string {
	return "payment: channel rejected: " + e.Code + ": " + e.Message
}

func (e *ChannelError) Unwrap() error { return ErrChannelRejected }

// PaymentDefaultError holds the app error instances a service maps the
// payment sentinels to: the mapping mechanism lives here, the concrete
// instances are supplied by each service's platform layer (same
// mechanism/instance split as cache.HandleCacheError), so this package never
// imports a service error proto.
type PaymentDefaultError struct {
	InvalidArgument  *kratoserrors.Error
	NotFound         *kratoserrors.Error
	Duplicate        *kratoserrors.Error
	ChannelRejected  *kratoserrors.Error
	Unavailable      *kratoserrors.Error
	Timeout          *kratoserrors.Error // optional: nil → timeouts map to Unavailable
	InvalidSignature *kratoserrors.Error // optional: nil → signature failures map to InvalidArgument
	DecryptFailed    *kratoserrors.Error // optional: nil → decryption failures map to InvalidArgument
}

// ValidatePaymentDefaultError fails fast if a required slot is nil — a nil
// slot nil-derefs inside WithCause. Timeout/InvalidSignature/DecryptFailed
// are optional.
func ValidatePaymentDefaultError(errs *PaymentDefaultError) error {
	for _, e := range []*kratoserrors.Error{
		errs.InvalidArgument, errs.NotFound, errs.Duplicate,
		errs.ChannelRejected, errs.Unavailable,
	} {
		if e == nil {
			return stderrors.New("payment: PaymentDefaultError has a nil sentinel slot")
		}
	}
	return nil
}

// HandlePaymentError maps a payment error to the supplied app error, keeping
// the original as cause. Unknown errors land on Unavailable — payment errors
// never leak past the service error domain untranslated.
func HandlePaymentError(err error, errs *PaymentDefaultError) error {
	if e := ValidatePaymentDefaultError(errs); e != nil {
		return e
	}
	switch {
	case stderrors.Is(err, ErrInvalidArgument):
		return errs.InvalidArgument.WithCause(err)
	case stderrors.Is(err, ErrNotFound):
		return errs.NotFound.WithCause(err)
	case stderrors.Is(err, ErrDuplicate):
		return errs.Duplicate.WithCause(err)
	case stderrors.Is(err, ErrChannelRejected):
		return errs.ChannelRejected.WithCause(err)
	case stderrors.Is(err, ErrTimeout):
		// Optional slot; a nil folds onto Unavailable — outcome unknown,
		// converge through the query surface.
		if errs.Timeout != nil {
			return errs.Timeout.WithCause(err)
		}
		return errs.Unavailable.WithCause(err)
	case stderrors.Is(err, ErrInvalidSignature):
		if errs.InvalidSignature != nil {
			return errs.InvalidSignature.WithCause(err)
		}
		return errs.InvalidArgument.WithCause(err)
	case stderrors.Is(err, ErrDecryptFailed):
		if errs.DecryptFailed != nil {
			return errs.DecryptFailed.WithCause(err)
		}
		return errs.InvalidArgument.WithCause(err)
	default:
		return errs.Unavailable.WithCause(err)
	}
}
