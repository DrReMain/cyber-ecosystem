package mock

import (
	"context"
	"fmt"
	"time"

	"cyber-ecosystem/shared-go/capability/payment"
)

// enforceChannel is the provider-knowledge gate the contract does not
// carry: which channels exist and which Payer hints each requires.
func enforceChannel(in *payment.CreatePaymentInput) error {
	switch in.Channel {
	case ChannelJSAPI:
		if in.Payer == nil || in.Payer.OpenID == "" {
			return fmt.Errorf("%w: %s requires payer.openid", payment.ErrInvalidArgument, in.Channel)
		}
	case ChannelH5:
		if in.Payer == nil || in.Payer.ClientIP == "" {
			return fmt.Errorf("%w: %s requires payer.client_ip", payment.ErrInvalidArgument, in.Channel)
		}
	case ChannelNative:
		// A QR rendered on the page — no payer hint applies.
	default:
		return fmt.Errorf("%w: unsupported channel %q", payment.ErrInvalidArgument, in.Channel)
	}
	return nil
}

func credentialFor(in *payment.CreatePaymentInput) payment.Credential {
	switch in.Channel {
	case ChannelJSAPI:
		return payment.Credential{Kind: payment.CredentialInvoke, Value: "mock-prepay-" + in.OutTradeNo}
	case ChannelH5:
		return payment.Credential{Kind: payment.CredentialJumpURL, Value: "https://mock.pay/jump?out_trade_no=" + in.OutTradeNo}
	default:
		return payment.Credential{Kind: payment.CredentialQRCode, Value: "mockqr://" + in.OutTradeNo}
	}
}

// sameOrderInput decides channel-idempotent reuse: the same trade number
// with identical parameters returns the same credential; different
// parameters is a duplicate refusal.
func sameOrderInput(a, b *payment.CreatePaymentInput) bool {
	payerEq := func() bool {
		if (a.Payer == nil) != (b.Payer == nil) {
			return false
		}
		if a.Payer != nil {
			return *a.Payer == *b.Payer
		}
		return true
	}
	return a.Channel == b.Channel &&
		a.AmountFen == b.AmountFen &&
		a.Description == b.Description &&
		a.NotifyURL == b.NotifyURL &&
		a.ExpireAt.Equal(b.ExpireAt) &&
		payerEq()
}

func sameRefundInput(a, b *payment.RefundInput) bool {
	return a.OutTradeNo == b.OutTradeNo &&
		a.AmountFen == b.AmountFen &&
		a.TotalAmountFen == b.TotalAmountFen &&
		a.Reason == b.Reason &&
		a.NotifyURL == b.NotifyURL
}

func (b *backend) CreatePayment(_ context.Context, in *payment.CreatePaymentInput) (*payment.PrepayResult, error) {
	if err := payment.ValidateCreatePaymentInput(in); err != nil {
		return nil, err
	}
	if err := enforceChannel(in); err != nil {
		return nil, err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if existing, ok := b.orders[in.OutTradeNo]; ok {
		if sameOrderInput(&existing.in, in) {
			return &payment.PrepayResult{OutTradeNo: in.OutTradeNo, Credential: existing.credential}, nil
		}
		return nil, fmt.Errorf("%w: %s", payment.ErrDuplicate, in.OutTradeNo)
	}
	if forced, ok := b.rejects[in.OutTradeNo]; ok {
		return nil, forced
	}
	o := &order{
		in:         *in,
		credential: credentialFor(in),
		state:      payment.PaymentStatePending,
		refunds:    make(map[string]*refundEntry),
	}
	b.orders[in.OutTradeNo] = o
	return &payment.PrepayResult{OutTradeNo: in.OutTradeNo, Credential: o.credential}, nil
}

func (b *backend) QueryPayment(_ context.Context, outTradeNo string) (*payment.PaymentStatusResult, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	o, ok := b.orders[outTradeNo]
	if !ok {
		return nil, fmt.Errorf("%w: order %s", payment.ErrNotFound, outTradeNo)
	}
	return &payment.PaymentStatusResult{
		OutTradeNo:    outTradeNo,
		TransactionID: o.txID,
		State:         o.state,
		AmountFen:     o.in.AmountFen,
		RefundedFen:   o.refunded,
		PaidAt:        o.paidAt,
	}, nil
}

func (b *backend) ClosePayment(_ context.Context, outTradeNo string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	o, ok := b.orders[outTradeNo]
	if !ok {
		return fmt.Errorf("%w: order %s", payment.ErrNotFound, outTradeNo)
	}
	switch o.state {
	case payment.PaymentStatePending:
		o.state = payment.PaymentStateClosed
		return nil
	case payment.PaymentStateClosed:
		// Idempotent re-close.
		return nil
	default:
		return &payment.ChannelError{Code: "TRADE_STATE_ERROR", Message: "a paid order cannot be closed"}
	}
}

func (b *backend) Refund(_ context.Context, in *payment.RefundInput) (*payment.RefundResult, error) {
	if err := payment.ValidateRefundInput(in); err != nil {
		return nil, err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	o, ok := b.orders[in.OutTradeNo]
	if !ok {
		return nil, fmt.Errorf("%w: order %s", payment.ErrNotFound, in.OutTradeNo)
	}
	if existing, ok := o.refunds[in.OutRefundNo]; ok {
		if sameRefundInput(&existing.in, in) {
			return &payment.RefundResult{OutRefundNo: in.OutRefundNo, RefundID: existing.refundID, State: existing.state}, nil
		}
		return nil, fmt.Errorf("%w: %s", payment.ErrDuplicate, in.OutRefundNo)
	}
	if in.TotalAmountFen != o.in.AmountFen {
		return nil, fmt.Errorf("%w: declared total %d contradicts order amount %d",
			payment.ErrInvalidArgument, int64(in.TotalAmountFen), int64(o.in.AmountFen))
	}
	if o.state != payment.PaymentStatePaid {
		return nil, &payment.ChannelError{Code: "TRADE_STATE_ERROR", Message: "only a paid order can be refunded"}
	}
	if o.refunded+in.AmountFen > o.in.AmountFen {
		return nil, &payment.ChannelError{
			Code:    "REFUND_EXCEED_BALANCE",
			Message: fmt.Sprintf("refund %d exceeds remaining %d", int64(in.AmountFen), int64(o.in.AmountFen-o.refunded)),
		}
	}
	e := &refundEntry{
		in:       *in,
		state:    payment.RefundStateSucceeded,
		refundID: refundIDFor(in.OutRefundNo),
		doneAt:   time.Now(),
	}
	o.refunds[in.OutRefundNo] = e
	o.refunded += in.AmountFen
	return &payment.RefundResult{OutRefundNo: in.OutRefundNo, RefundID: e.refundID, State: e.state}, nil
}

func (b *backend) QueryRefund(_ context.Context, outRefundNo string) (*payment.RefundStatusResult, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, o := range b.orders {
		if e, ok := o.refunds[outRefundNo]; ok {
			return &payment.RefundStatusResult{
				OutRefundNo: outRefundNo,
				RefundID:    e.refundID,
				State:       e.state,
				AmountFen:   e.in.AmountFen,
				SuccessAt:   e.doneAt,
			}, nil
		}
	}
	return nil, fmt.Errorf("%w: refund %s", payment.ErrNotFound, outRefundNo)
}
