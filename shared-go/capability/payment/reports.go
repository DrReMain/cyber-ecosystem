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
// (CSV, possibly gzipped — provider-specific formats). Neutralizing them into
// contract rows is provider wire knowledge and lives one call away on the
// same surface; the consumer's reconciliation job starts at BillLine.
type BillResult struct {
	Raw []byte
}

// BillLineState names a statement row's fact in the contract's vocabulary;
// the provider maps its own wire words onto this set (WeChat's
// SUCCESS/REFUND/REVOKED/USERPAYING, the mock's state strings).
type BillLineState string

const (
	BillLinePaid    BillLineState = "paid"
	BillLineRefund  BillLineState = "refund"
	BillLineClosed  BillLineState = "closed"
	BillLinePending BillLineState = "pending"
)

// BillLine is one neutralized statement row. A statement is row-per-fact,
// never row-per-order — the same out_trade_no appears once per channel-side
// transition (a refund lands beside its payment), and the consumer's
// reconciliation aggregates. AmountFen is the order amount every row of the
// trade bill carries; RefundFen is this row's refund (zero unless the state
// is refund); IncomeFen is the settled amount after channel deductions.
type BillLine struct {
	OutTradeNo    string
	TransactionID string
	State         BillLineState
	AmountFen     AmountFen
	RefundFen     AmountFen
	FeeFen        AmountFen
	IncomeFen     AmountFen
	At            time.Time
}

// Reports is the reconciliation surface: the daily cross-check that catches
// what notifications and queries alone cannot (missed callbacks, channel-side
// discrepancies). Money systems settle on the bill, not on the event stream.
type Reports interface {
	DownloadBill(ctx context.Context, in *BillInput) (*BillResult, error)
	ParseTradeBill(ctx context.Context, raw []byte) ([]BillLine, error)
}
