package payment

import (
	"context"
	"net/http"
	"time"
)

// EventType is the neutral notification vocabulary. Channels push far more
// event kinds than these; unknown kinds parse as themselves via ChannelType —
// the consumer audits, the contract never drops a notification on the floor
// for being unrecognized.
type EventType string

const (
	EventPaymentSucceeded EventType = "payment_succeeded"
	EventRefundSucceeded  EventType = "refund_succeeded"
	EventRefundAbnormal   EventType = "refund_abnormal"
)

// NotificationInput is the raw HTTP notification exactly as it arrived:
// headers and body untouched. The signature/decryption material lives in the
// provider (platform certs, APIv3 key), so parsing is a provider operation —
// a bad signature or an undecryptable payload is an error, never an event.
type NotificationInput struct {
	Headers http.Header
	Body    []byte
}

// NotificationEvent is the decoded notification. Consumers must treat it as
// at-least-once: channels redeliver (WeChat retries until acked), so every
// consumer path must be idempotent — the conformance suite pins duplicate
// notification ⇒ no double settlement.
type NotificationEvent struct {
	Type        EventType
	ChannelType string
	OutTradeNo  string
	OutRefundNo string
	// TransactionID / RefundID are the channel-side identifiers for audit
	// joins against the channel's bill.
	TransactionID string
	RefundID      string
	// State / RefundState / AmountFen / RefundedFen: for payment events,
	// State is PaymentStatePaid and AmountFen is the order amount; for
	// refund events, RefundState carries the refund state, AmountFen is the
	// refunded amount of this notification and RefundedFen the cumulative
	// refunded total on the order after it.
	State       PaymentState
	RefundState RefundState
	AmountFen   AmountFen
	RefundedFen AmountFen
	SuccessAt   time.Time
}

// Notifier parses channel notifications. Duplicate delivery is normal; a
// notification is an accelerator — the settled state always belongs to
// QueryPayment / QueryRefund.
type Notifier interface {
	Parse(ctx context.Context, in *NotificationInput) (*NotificationEvent, error)
}
