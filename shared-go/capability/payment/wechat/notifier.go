package wechat

// The notification face: a channel push is a signed envelope wrapping an
// AES-GCM-sealed domain event. Verification reuses the client's serial-routed
// ladder (the same key table, the same replay window), so a notification and
// a response cannot drift apart in how they are trusted. Event types beyond
// the contract's three parse successfully with a zero Type and the raw
// channel type riding ChannelType — the consumer audits, the parse never
// drops a notification for being unrecognized.

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cyber-ecosystem/shared-go/capability/payment"
)

type notifier struct {
	c *client
}

func (n *notifier) Parse(ctx context.Context, in *payment.NotificationInput) (*payment.NotificationEvent, error) {
	if err := payment.ValidateNotificationInput(in); err != nil {
		return nil, err
	}
	if err := n.c.verifySigned(ctx, in.Headers, in.Body); err != nil {
		return nil, err
	}
	var wire notificationWire
	if err := json.Unmarshal(in.Body, &wire); err != nil {
		return nil, fmt.Errorf("%w: malformed notification: %w", payment.ErrInvalidArgument, err)
	}
	if wire.Resource.Algorithm != "AEAD_AES_256_GCM" {
		return nil, fmt.Errorf("%w: unsupported algorithm %q", payment.ErrInvalidArgument, wire.Resource.Algorithm)
	}
	plaintext, err := DecryptAES256GCM(n.c.cfg.APIv3Key,
		wire.Resource.Nonce, wire.Resource.AssociatedData, wire.Resource.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", payment.ErrDecryptFailed, err)
	}
	event := &payment.NotificationEvent{ChannelType: wire.EventType}
	switch wire.Resource.OriginalType {
	case "transaction":
		var tx transactionEventWire
		if err := json.Unmarshal(plaintext, &tx); err != nil {
			return nil, fmt.Errorf("%w: malformed transaction: %w", payment.ErrInvalidArgument, err)
		}
		state, err := mapTradeState(tx.TradeState)
		if err != nil {
			return nil, err
		}
		if wire.EventType == "TRANSACTION.SUCCESS" {
			event.Type = payment.EventPaymentSucceeded
		}
		event.OutTradeNo = tx.OutTradeNo
		event.TransactionID = tx.TransactionID
		event.State = state
		event.AmountFen = payment.Fen(tx.Amount.Total)
		event.SuccessAt = parseChannelTime(tx.SuccessTime)
	case "refund":
		var rf refundEventWire
		if err := json.Unmarshal(plaintext, &rf); err != nil {
			return nil, fmt.Errorf("%w: malformed refund: %w", payment.ErrInvalidArgument, err)
		}
		state, err := mapRefundStatus(rf.RefundStatus)
		if err != nil {
			return nil, err
		}
		switch wire.EventType {
		case "REFUND.SUCCESS":
			event.Type = payment.EventRefundSucceeded
		case "REFUND.ABNORMAL":
			event.Type = payment.EventRefundAbnormal
		}
		event.OutTradeNo = rf.OutTradeNo
		event.OutRefundNo = rf.OutRefundNo
		event.RefundID = rf.RefundID
		event.RefundState = state
		// The refund notification carries this refund's amount only; the
		// cumulative total is a query fact (QueryPayment), not a push fact.
		event.AmountFen = payment.Fen(rf.Amount.Refund)
		event.SuccessAt = parseChannelTime(rf.SuccessTime)
	default:
		// An original_type this build has never met still parses: the
		// consumer reads ChannelType and decides.
	}
	return event, nil
}

// parseChannelTime tolerates a missing or unparseable stamp: a push is an
// accelerator whose value is the state, while the query paths refuse a bad
// success_time outright — their stamps are facts being settled on.
func parseChannelTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

type notificationWire struct {
	EventType string `json:"event_type"`
	Resource  struct {
		OriginalType   string `json:"original_type"`
		Algorithm      string `json:"algorithm"`
		Nonce          string `json:"nonce"`
		AssociatedData string `json:"associated_data"`
		Ciphertext     string `json:"ciphertext"`
	} `json:"resource"`
}

type transactionEventWire struct {
	OutTradeNo    string `json:"out_trade_no"`
	TransactionID string `json:"transaction_id"`
	TradeState    string `json:"trade_state"`
	SuccessTime   string `json:"success_time"`
	Amount        struct {
		Total int64 `json:"total"`
	} `json:"amount"`
}

type refundEventWire struct {
	OutTradeNo   string `json:"out_trade_no"`
	OutRefundNo  string `json:"out_refund_no"`
	RefundID     string `json:"refund_id"`
	RefundStatus string `json:"refund_status"`
	SuccessTime  string `json:"success_time"`
	Amount       struct {
		Refund int64 `json:"refund"`
	} `json:"amount"`
}
