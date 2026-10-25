package payment

import "time"

// RefundInput refunds part or all of a paid order. AmountFen is explicit —
// full refund passes the original amount, there is no magic zero for "all":
// a zero/negative amount is ErrInvalidArgument before it ever reaches the
// channel. Whether this refund (plus prior refunds) fits within the order
// amount is a stateful invariant enforced by the provider, not by input
// validation.
type RefundInput struct {
	OutRefundNo string
	OutTradeNo  string
	AmountFen   AmountFen
	// TotalAmountFen is the original order amount — refund endpoints require
	// it (WeChat's amount.total is mandatory), and the caller carries it from
	// the order row, not from memory.
	TotalAmountFen AmountFen
	Reason         string
	// NotifyURL is optional; without it no refund notification is pushed and
	// settlement converges through QueryRefund.
	NotifyURL string
}

// RefundState is the neutral refund state. Abnormal keeps the channel's
// manual-intervention bucket (frozen banks, channel anomalies) — it is a
// terminal-for-now state that human ops resolve, not one the caller retries.
type RefundState string

const (
	RefundStateProcessing RefundState = "processing"
	RefundStateSucceeded  RefundState = "succeeded"
	RefundStateAbnormal   RefundState = "abnormal"
)

// RefundResult is the synchronous answer of a refund request; channels work
// asynchronously, so Succeeded here means accepted-and-done only when State
// says so — Processing resolves through QueryRefund or the refund
// notification.
type RefundResult struct {
	OutRefundNo string
	RefundID    string
	State       RefundState
}

// RefundStatusResult is the truth read for one refund, keyed by the caller's
// refund number.
type RefundStatusResult struct {
	OutRefundNo string
	RefundID    string
	State       RefundState
	AmountFen   AmountFen
	// SuccessAt is zero unless State is Succeeded.
	SuccessAt time.Time
}
