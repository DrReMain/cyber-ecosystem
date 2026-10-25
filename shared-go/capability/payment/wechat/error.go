package wechat

import (
	"fmt"
	"net/http"

	"cyber-ecosystem/shared-go/capability/payment"
)

// APIError is a non-2xx channel response: the status, the channel's machine
// code and human message, and the Request-ID of the failed call for support
// escalation.
type APIError struct {
	Status    int
	Code      string
	Message   string
	RequestID string
}

func (e *APIError) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("wechat pay: http %d %s: %s (request %s)", e.Status, e.Code, e.Message, e.RequestID)
	}
	return fmt.Sprintf("wechat pay: http %d %s: %s", e.Status, e.Code, e.Message)
}

// classifyAPIError folds a channel rejection onto the contract's error
// domain. Input and credential faults are the caller's to fix; congestion
// and channel-state faults are retriable unavailability; every other
// business rejection stays a ChannelError the consumer reads through
// ErrChannelRejected.
func classifyAPIError(a *APIError) error {
	switch {
	case a.Code == "NO_BILL_EXIST":
		// A statement day with no bill is a query miss, not a bad request —
		// the channel reports it as 400 regardless.
		return fmt.Errorf("%w: %s", payment.ErrNotFound, a.Code)
	case a.Status == http.StatusBadRequest, a.Status == http.StatusUnauthorized:
		// 401 is the channel refusing our signature — a credential or
		// signing fault on our side, the same caller-fixes-it bucket as a
		// malformed body.
		return fmt.Errorf("%w: %s: %s", payment.ErrInvalidArgument, a.Code, a.Message)
	case a.Status == http.StatusNotFound:
		return fmt.Errorf("%w: %s: %s", payment.ErrNotFound, a.Code, a.Message)
	case a.Status == http.StatusForbidden && a.Code == "OUT_TRADE_NO_USED":
		// Same trade number reused with different parameters; an identical
		// retry is idempotent and never reaches this branch.
		return fmt.Errorf("%w: %s", payment.ErrDuplicate, a.Code)
	case a.Status == http.StatusTooManyRequests, a.Status >= 500:
		return fmt.Errorf("%w: %s", payment.ErrUnavailable, a.Code)
	default:
		return &payment.ChannelError{Code: a.Code, Message: a.Message}
	}
}
