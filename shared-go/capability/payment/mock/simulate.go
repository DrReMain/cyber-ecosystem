package mock

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"cyber-ecosystem/shared-go/capability/payment"
)

// Controller is the test/dev face over the backend the container shares:
// it moves the state machine the way the real channel would (a payment
// happens, a refund lands), mints the notifications those transitions
// produce, and delivers them over real HTTP. It is never reachable through
// the payment container — production code sees only the contract.
type Controller struct {
	b *backend
}

// RejectWith forces the next CreatePayment for outTradeNo to fail as a
// channel business rejection — the ErrChannelRejected conformance probe.
func (c *Controller) RejectWith(outTradeNo string, code, message string) {
	c.b.mu.Lock()
	defer c.b.mu.Unlock()
	c.b.rejects[outTradeNo] = &payment.ChannelError{Code: code, Message: message}
}

// MarkPaid settles the order the way the channel would and returns the
// minted payment-succeeded notification (cached: repeated calls return the
// same bytes, so duplicate delivery is a redelivery, not a remint).
func (c *Controller) MarkPaid(outTradeNo string) (*payment.NotificationInput, error) {
	c.b.mu.Lock()
	o, ok := c.b.orders[outTradeNo]
	if !ok {
		c.b.mu.Unlock()
		return nil, fmt.Errorf("%w: order %s", payment.ErrNotFound, outTradeNo)
	}
	if o.state == payment.PaymentStatePaid {
		cached := o.notification
		c.b.mu.Unlock()
		return cached, nil
	}
	if o.state != payment.PaymentStatePending {
		c.b.mu.Unlock()
		return nil, &payment.ChannelError{Code: "TRADE_STATE_ERROR", Message: "cannot pay a closed order"}
	}
	o.state = payment.PaymentStatePaid
	o.paidAt = time.Now()
	o.txID = txIDFor(outTradeNo)
	event := &payment.NotificationEvent{
		Type:          payment.EventPaymentSucceeded,
		ChannelType:   "MOCK.TRANSACTION.SUCCESS",
		OutTradeNo:    outTradeNo,
		TransactionID: o.txID,
		State:         payment.PaymentStatePaid,
		AmountFen:     o.in.AmountFen,
		SuccessAt:     o.paidAt,
	}
	ni, err := c.b.seal(event)
	if err != nil {
		c.b.mu.Unlock()
		return nil, err
	}
	o.notification = ni
	c.b.mu.Unlock()
	return ni, nil
}

// RefundNotification returns the minted refund-succeeded notification for a
// refund the Gateway already settled — the minting side only; state moves
// with Refund itself, mirroring a synchronous refund channel.
func (c *Controller) RefundNotification(outRefundNo string) (*payment.NotificationInput, error) {
	c.b.mu.Lock()
	defer c.b.mu.Unlock()
	for _, o := range c.b.orders {
		e, ok := o.refunds[outRefundNo]
		if !ok {
			continue
		}
		event := &payment.NotificationEvent{
			Type:        payment.EventRefundSucceeded,
			ChannelType: "MOCK.REFUND.SUCCESS",
			OutTradeNo:  e.in.OutTradeNo,
			OutRefundNo: outRefundNo,
			RefundID:    e.refundID,
			RefundState: payment.RefundStateSucceeded,
			AmountFen:   e.in.AmountFen,
			RefundedFen: o.refunded,
			SuccessAt:   e.doneAt,
		}
		return c.b.seal(event)
	}
	return nil, fmt.Errorf("%w: refund %s", payment.ErrNotFound, outRefundNo)
}

// Deliver POSTs a notification to the consumer's notify endpoint repeat
// times at the given interval (channels redeliver until acked, so callback
// paths must survive duplicates). It returns the first delivery error.
func (c *Controller) Deliver(ni *payment.NotificationInput, url string, repeat int, interval time.Duration) error {
	for i := range repeat {
		if i > 0 {
			time.Sleep(interval)
		}
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(ni.Body))
		if err != nil {
			return fmt.Errorf("mock deliver: build request: %w", err)
		}
		req.Header = ni.Headers.Clone()
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("mock deliver: %w", err)
		}
		_ = resp.Body.Close()
	}
	return nil
}
