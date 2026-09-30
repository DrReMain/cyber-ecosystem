package mq

import (
	stderrors "errors"

	kratoserrors "github.com/go-kratos/kratos/v3/errors"
)

// Sentinel errors are the backend-agnostic MQ contract, mapped 1:1 to the
// InfraError 34xx MQ codes (infra.proto). Consumer handler BUSINESS errors are
// not MQ errors — they are retried/DLQ'd, never passed here.
var (
	ErrInvalidArgument = stderrors.New("mq: invalid argument")
	ErrUnavailable     = stderrors.New("mq: unavailable")
	ErrTimeout         = stderrors.New("mq: timeout")
)

// MQDefaultError holds the app error instances a service maps the MQ sentinels
// to: the mapping mechanism lives here, the concrete instances are supplied by
// each service's platform layer (same mechanism/instance split as
// cache.HandleCacheError), so this package never imports a service error proto.
type MQDefaultError struct {
	InvalidArgument *kratoserrors.Error
	Unavailable     *kratoserrors.Error
	Timeout         *kratoserrors.Error // optional: nil → timeouts map to Unavailable
}

// ValidateMQDefaultError fails fast if a required slot is nil — a nil slot
// nil-derefs inside WithCause. Timeout is optional.
func ValidateMQDefaultError(errs *MQDefaultError) error {
	for _, e := range []*kratoserrors.Error{errs.InvalidArgument, errs.Unavailable} {
		if e == nil {
			return stderrors.New("mq: MQDefaultError has a nil sentinel slot")
		}
	}
	return nil
}

// HandleMQError maps an MQ-infra error (publish/connect) to the supplied app
// error, keeping the original as cause. Consumer handler errors are
// retried/DLQ'd, never mapped here.
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
		return errs.Unavailable.WithCause(err)
	}
}
