package payment

import (
	"context"
	"strconv"
	"time"
)

// Channel names the access context a payment runs in; the access
// environment decides it, not user preference: in-app browsers pay by
// invoke, phone browsers pay by jump (the payment app is pulled up; without
// it installed a jump cannot complete), desktop pays by QR (scanned by a
// phone; completion is learned by polling QueryPayment — no browser
// callback). The value set is provider knowledge: each backend declares
// its own values ("<provider>_<mode>") and rejects what it does not know
// as ErrInvalidArgument; the contract fixes only the type.
type Channel string

// AmountFen is a money amount in fen (1/100 yuan) — the smallest CNY unit.
// The whole chain (channel wire, order domain, storage) stays in fen so no
// conversion ever happens; the channel wire format is itself int64 fen, so a
// bare int64 is only ever a unit-confusion accident waiting to happen. Money
// never rides float types anywhere.
type AmountFen int64

// Fen constructs an AmountFen from a fen count.
func Fen(n int64) AmountFen { return AmountFen(n) }

// String renders the amount as yuan with a two-digit fen part ("12.34",
// "0.05"), display-only: arithmetic stays on the integer value.
func (a AmountFen) String() string {
	neg := a < 0
	if neg {
		a = -a
	}
	yuan := int64(a) / 100
	fen := int64(a) % 100
	sign := ""
	if neg {
		sign = "-"
	}
	if fen < 10 {
		return sign + strconv.FormatInt(yuan, 10) + ".0" + strconv.FormatInt(fen, 10)
	}
	return sign + strconv.FormatInt(yuan, 10) + "." + strconv.FormatInt(fen, 10)
}

// Payer carries the channel hints a Channel requires. OpenID is the merchant
// appid-scoped id (required by JSAPI, supplied by the login side — for an
// SMS-registered account that id only exists once the WeChat binding lands);
// ClientIP is the payer's real IP (required by H5 for risk control);
// ReturnURL is the optional post-payment landing page.
type Payer struct {
	OpenID    string
	ClientIP  string
	ReturnURL string
}

// CreatePaymentInput is the caller-minted order: OutTradeNo is the caller's
// idempotency key — creating again with the same key and identical parameters
// returns the same credential (channel idempotency); the same key with
// different parameters is ErrDuplicate.
type CreatePaymentInput struct {
	Channel     Channel
	OutTradeNo  string
	AmountFen   AmountFen
	Description string
	NotifyURL   string
	// ExpireAt is the payment deadline; the zero value means the channel
	// default (WeChat: two hours).
	ExpireAt time.Time
	Payer    *Payer
}

// CredentialKind names what the caller does with the credential to continue
// the payment; neutral across channels (Alipay's QR/WAP values land here too).
type CredentialKind string

const (
	// CredentialInvoke carries what an in-app payment invoke needs; the
	// provider's docs own the value's exact shape for its channels.
	CredentialInvoke CredentialKind = "invoke"
	// CredentialJumpURL is a URL the page navigates to; the target page pulls
	// up the payment app.
	CredentialJumpURL CredentialKind = "jump_url"
	// CredentialQRCode is a QR payload the page renders for scanning;
	// completion is learned by polling QueryPayment.
	CredentialQRCode CredentialKind = "qr_code"
)

// Credential is the channel's answer to a created payment: what it takes to
// continue, never payment state.
type Credential struct {
	Kind  CredentialKind
	Value string
}

// PrepayResult is the idempotent product of CreatePayment.
type PrepayResult struct {
	OutTradeNo string
	Credential Credential
}

// PaymentState is the neutral channel payment state. Partial refunds are a
// derived fact (RefundedFen on the status), not a state: a paid order with
// refunds stays Paid.
type PaymentState string

const (
	PaymentStatePending PaymentState = "pending"
	PaymentStatePaid    PaymentState = "paid"
	PaymentStateClosed  PaymentState = "closed"
)

// PaymentStatusResult is the truth read. In the settle architecture the
// notification is only an accelerator — the authoritative state always
// converges here.
type PaymentStatusResult struct {
	OutTradeNo    string
	TransactionID string
	State         PaymentState
	AmountFen     AmountFen
	// RefundedFen is the cumulative refunded amount on this order (stateful
	// invariant: the provider enforces the sum never exceeding AmountFen).
	RefundedFen AmountFen
	// PaidAt is zero unless State is Paid.
	PaidAt time.Time
}

// Gateway is the trading surface. Unknown outcomes (ErrTimeout, network
// breaks) converge through QueryPayment — never through blind retry of the
// mutating calls, whose channel idempotency is keyed on the caller's
// OutTradeNo / OutRefundNo.
type Gateway interface {
	CreatePayment(ctx context.Context, in *CreatePaymentInput) (*PrepayResult, error)
	QueryPayment(ctx context.Context, outTradeNo string) (*PaymentStatusResult, error)
	ClosePayment(ctx context.Context, outTradeNo string) error
	Refund(ctx context.Context, in *RefundInput) (*RefundResult, error)
	QueryRefund(ctx context.Context, outRefundNo string) (*RefundStatusResult, error)
}
