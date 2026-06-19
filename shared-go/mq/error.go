package mq

import (
	stderrors "errors"

	kratoserrors "github.com/go-kratos/kratos/v3/errors"
)

// MQDefaultError holds the application error instances a service maps the MQ
// sentinel errors to. The mapping MECHANISM lives here in shared-go (reusable,
// backend-agnostic); the concrete *errors.Error instances are supplied by each
// service's platform layer — the same mechanism/instance split cache/storage use.
type MQDefaultError struct {
	InvalidArgument *kratoserrors.Error
	Unavailable     *kratoserrors.Error
	Timeout         *kratoserrors.Error // optional: nil → unknown errors map to Unavailable
}

// ValidateMQDefaultError fails fast at construction if a required sentinel
// slot is nil. Timeout is optional (nil-safe in HandleMQError).
func ValidateMQDefaultError(errs *MQDefaultError) error {
	for _, e := range []*kratoserrors.Error{errs.InvalidArgument, errs.Unavailable} {
		if e == nil {
			return stderrors.New("mq: MQDefaultError has a nil sentinel slot")
		}
	}
	return nil
}

// HandleMQError maps an MQ-infra error to the supplied application error,
// attaching the original as the cause. Used only for infra errors (publish/
// connect); consumer handler errors are retried/DLQ'd, not mapped here.
func HandleMQError(err error, errs *MQDefaultError) error {
	if e := ValidateMQDefaultError(errs); e != nil {
		return e
	}
	switch {
	case stderrors.Is(err, ErrInvalidArgument):
		return errs.InvalidArgument.WithCause(err)
	case stderrors.Is(err, ErrUnavailable):
		return errs.Unavailable.WithCause(err)
	case stderrors.Is(err, ErrTimeout) && errs.Timeout != nil:
		return errs.Timeout.WithCause(err)
	default:
		// Unavailable is guaranteed non-nil by ValidateMQDefaultError above.
		return errs.Unavailable.WithCause(err)
	}
}
