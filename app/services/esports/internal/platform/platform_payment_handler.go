package platform

import (
	"cyber-ecosystem/shared-go/capability/payment"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
)

var defaultPaymentError = &payment.PaymentDefaultError{
	InvalidArgument:  errorspb.ErrorInfraErrorPaymentInvalidArgument(""),
	NotFound:         errorspb.ErrorInfraErrorPaymentNotFound(""),
	Duplicate:        errorspb.ErrorInfraErrorPaymentDuplicate(""),
	ChannelRejected:  errorspb.ErrorInfraErrorPaymentChannelRejected(""),
	Unavailable:      errorspb.ErrorInfraErrorPaymentUnavailable(""),
	Timeout:          errorspb.ErrorInfraErrorPaymentTimeout(""),
	InvalidSignature: errorspb.ErrorInfraErrorPaymentInvalidSignature(""),
	DecryptFailed:    errorspb.ErrorInfraErrorPaymentDecryptFailed(""),
}

func NewPaymentErrorHandler() (PaymentErrorHandler, error) {
	if err := payment.ValidatePaymentDefaultError(defaultPaymentError); err != nil {
		return nil, err
	}
	return func(err error) error {
		return payment.HandlePaymentError(err, defaultPaymentError)
	}, nil
}
