package platform

import (
	"cyber-ecosystem/shared-go/capability/mq"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
)

var defaultMQError = &mq.MQDefaultError{
	InvalidArgument: errorspb.ErrorInfraErrorMqInvalidArgument(""),
	Unavailable:     errorspb.ErrorInfraErrorMqUnavailable(""),
	Timeout:         errorspb.ErrorInfraErrorMqTimeout(""),
}

func NewMQErrorHandler() (MQErrorHandler, error) {
	if err := mq.ValidateMQDefaultError(defaultMQError); err != nil {
		return nil, err
	}
	return func(err error) error {
		return mq.HandleMQError(err, defaultMQError)
	}, nil
}
