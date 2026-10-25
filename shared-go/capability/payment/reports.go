package payment

import (
	"context"
	"time"
)

// BillKind names the channel's daily statement kinds. TradeBill is the
// transaction statement (every order/refund with its fees); FundFlowBill is
// the money-movement statement (actual account balance changes).
type BillKind string

const (
	BillKindTrade    BillKind = "trade"
	BillKindFundFlow BillKind = "fund_flow"
)

// BillInput requests one statement day. Date is the calendar day in the
// channel's statement timezone (WeChat: UTC+8); the provider formats what the
// channel wire needs. Channels publish a day's statement the following day —
// a missing recent statement is the channel being the channel, not an error
// to retry against.
type BillInput struct {
	Date time.Time
	Kind BillKind
}

// BillResult is the raw statement bytes exactly as the channel serves them
// (CSV, possibly gzipped — provider-specific formats). Parsing is the
// consumer's reconciliation job; the capability transports, never interprets.
type BillResult struct {
	Raw []byte
}

// Reports is the reconciliation surface: the daily cross-check that catches
// what notifications and queries alone cannot (missed callbacks, channel-side
// discrepancies). Money systems settle on the bill, not on the event stream.
type Reports interface {
	DownloadBill(ctx context.Context, in *BillInput) (*BillResult, error)
}
