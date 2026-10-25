package payment

import "time"

const (
	// maxTradeNoLen is the trade/refund number cap shared by every known
	// channel wire (WeChat caps at 32; AliPay tolerates 64 but 32 keeps the
	// contract one rule for all backends).
	maxTradeNoLen = 32
)

// isTradeNo reports whether s contains only bytes safe as a merchant-side
// trade/refund number on every known channel: ASCII alphanumerics, dash and
// underscore. Everything else (separators, whitespace, non-ASCII, control)
// is rejected before it can reach a channel that would reject it less
// legibly.
func isTradeNo(s string) bool {
	if s == "" || len(s) > maxTradeNoLen {
		return false
	}
	for _, r := range s {
		switch {
		case 'a' <= r && r <= 'z',
			'A' <= r && r <= 'Z',
			'0' <= r && r <= '9',
			r == '-', r == '_':
		default:
			return false
		}
	}
	return true
}

// ValidateCreatePaymentInput checks the channel-neutral single-request
// shape: trade number well-formed, amount positive, notification URL
// present, deadline sane. Which channels exist and which Payer hints each
// requires are provider knowledge — the backend rejects an unsupported
// channel or a missing hint as ErrInvalidArgument. Stateful invariants
// (duplicate trade numbers, refund sums) belong to the provider's state
// machine; input validation never sees state.
func ValidateCreatePaymentInput(in *CreatePaymentInput) error {
	if in == nil {
		return ErrInvalidArgument
	}
	if !isTradeNo(in.OutTradeNo) {
		return ErrInvalidArgument
	}
	if in.AmountFen <= 0 {
		return ErrInvalidArgument
	}
	if in.NotifyURL == "" {
		return ErrInvalidArgument
	}
	if !in.ExpireAt.IsZero() && !in.ExpireAt.After(time.Now()) {
		return ErrInvalidArgument
	}
	return nil
}

// ValidateRefundInput checks the single-request shape: refund numbers
// well-formed, the refund itself and the declared original amount positive,
// the refund within the declared original. The stateful invariant (this
// refund plus prior refunds within the real order amount) stays with the
// provider, which also refuses a declared total that contradicts the order.
func ValidateRefundInput(in *RefundInput) error {
	if in == nil {
		return ErrInvalidArgument
	}
	if !isTradeNo(in.OutRefundNo) || !isTradeNo(in.OutTradeNo) {
		return ErrInvalidArgument
	}
	if in.AmountFen <= 0 || in.TotalAmountFen <= 0 {
		return ErrInvalidArgument
	}
	if in.AmountFen > in.TotalAmountFen {
		return ErrInvalidArgument
	}
	return nil
}

// ValidateNotificationInput checks the raw envelope is present; signature
// and decryption validity are the Notifier's verdict, never a shape check.
func ValidateNotificationInput(in *NotificationInput) error {
	if in == nil || in.Body == nil || len(in.Body) == 0 {
		return ErrInvalidArgument
	}
	return nil
}

// ValidateBillInput checks the statement request: a real statement day and a
// declared bill kind. Trade and fund-flow are payment-domain universals, so
// unlike channels this value set is contract knowledge.
func ValidateBillInput(in *BillInput) error {
	if in == nil || in.Date.IsZero() {
		return ErrInvalidArgument
	}
	switch in.Kind {
	case BillKindTrade, BillKindFundFlow:
		return nil
	default:
		return ErrInvalidArgument
	}
}
