package platform

import (
	"cyber-ecosystem/shared-go/capability/sms"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
)

var defaultSmsError = &sms.SmsDefaultError{
	InvalidArgument: errorspb.ErrorInfraErrorSmsInvalidArgument(""),
	ChannelRejected: errorspb.ErrorInfraErrorSmsChannelRejected(""),
	Unavailable:     errorspb.ErrorInfraErrorSmsUnavailable(""),
	Timeout:         errorspb.ErrorInfraErrorSmsTimeout(""),
}

func NewSmsErrorHandler() (SmsErrorHandler, error) {
	if err := sms.ValidateSmsDefaultError(defaultSmsError); err != nil {
		return nil, err
	}
	return func(err error) error {
		return sms.HandleSmsError(err, defaultSmsError)
	}, nil
}
