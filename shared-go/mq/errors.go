package mq

import "errors"

// Sentinel errors are the MQ contract (backend-agnostic). They map to the
// InfraError 34xx MQ codes (infra.proto); the service platform layer adapts
// them to application errors via HandleMQError. Consumer handler BUSINESS
// errors are not MQ errors — they are retried/DLQ'd, never passed here.
var (
	ErrInvalidArgument = errors.New("mq: invalid argument")
	ErrUnavailable     = errors.New("mq: unavailable")
	ErrTimeout         = errors.New("mq: timeout")
)
