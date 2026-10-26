package wechat

// Notification probes with channel-shaped envelopes: the wire the real edge
// pushes (signed headers, AES-GCM-sealed resource), the event folding onto
// the contract's vocabulary, and the refusal ladder for everything that is
// not provably from the channel.

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/capability/payment"
)

func newTestNotifier(t *testing.T) (*fakeChannel, *notifier) {
	t.Helper()
	f := newFakeChannel(t)
	c, err := newClient(f.publicKeyConfig(t))
	if err != nil {
		t.Fatalf("newClient: %v", err)
	}
	return f, &notifier{c: c}
}

// mintNotification builds a channel-shaped push: the domain event sealed
// under the APIv3 key (associated data = original type, the protocol's
// convention) inside the envelope, signed over the whole body.
func mintNotification(t *testing.T, f *fakeChannel, eventType, originalType string, payload any) *payment.NotificationInput {
	t.Helper()
	plaintext, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	nonce := freshNonce()
	ciphertext, err := EncryptAES256GCM(f.apiv3Key, []byte(nonce), originalType, plaintext)
	if err != nil {
		t.Fatalf("seal: %v", err)
	}
	body := mustMarshal(t, map[string]any{
		"id":            "NOTIF-1",
		"event_type":    eventType,
		"resource_type": "encrypt-resource",
		"resource": map[string]string{
			"original_type":   originalType,
			"algorithm":       "AEAD_AES_256_GCM",
			"nonce":           nonce,
			"associated_data": originalType,
			"ciphertext":      ciphertext,
		},
	})
	h := f.pushHeaders([]byte(body), time.Now().Unix())
	return &payment.NotificationInput{Headers: h, Body: []byte(body)}
}

func TestNotifierPaymentEvent(t *testing.T) {
	f, n := newTestNotifier(t)
	ni := mintNotification(t, f, "TRANSACTION.SUCCESS", "transaction", map[string]any{
		"mchid":          testMchID,
		"out_trade_no":   "ord-1",
		"transaction_id": "tx-9",
		"trade_state":    "SUCCESS",
		"success_time":   "2026-10-25T12:00:00+08:00",
		"amount":         map[string]any{"total": 100, "payer_total": 100},
	})
	event, err := n.Parse(context.Background(), ni)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	want := payment.NotificationEvent{
		Type:          payment.EventPaymentSucceeded,
		ChannelType:   "TRANSACTION.SUCCESS",
		OutTradeNo:    "ord-1",
		TransactionID: "tx-9",
		State:         payment.PaymentStatePaid,
		AmountFen:     payment.Fen(100),
		SuccessAt:     time.Date(2026, 10, 25, 12, 0, 0, 0, time.FixedZone("", 8*3600)),
	}
	// Wall-clock compare: the parsed zone keeps its label ("CST"), the
	// fixture's synthetic zone does not — same instant either way.
	if !event.SuccessAt.Equal(want.SuccessAt) {
		t.Errorf("success at: want %v, got %v", want.SuccessAt, event.SuccessAt)
	}
	event.SuccessAt, want.SuccessAt = time.Time{}, time.Time{}
	if *event != want {
		t.Errorf("event: want %+v, got %+v", want, event)
	}
}

func TestNotifierRefundEvent(t *testing.T) {
	f, n := newTestNotifier(t)
	ni := mintNotification(t, f, "REFUND.SUCCESS", "refund", map[string]any{
		"out_trade_no":  "ord-1",
		"out_refund_no": "ref-1",
		"refund_id":     "rf-9",
		"refund_status": "SUCCESS",
		"success_time":  "2026-10-25T12:05:00+08:00",
		"amount":        map[string]any{"refund": 40, "payer_refund": 40, "total": 100},
	})
	event, err := n.Parse(context.Background(), ni)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if event.Type != payment.EventRefundSucceeded || event.RefundState != payment.RefundStateSucceeded {
		t.Errorf("type/state: %+v", event)
	}
	if event.OutRefundNo != "ref-1" || event.RefundID != "rf-9" || event.AmountFen != payment.Fen(40) {
		t.Errorf("refund identity: %+v", event)
	}
	// The cumulative total is a query fact; the push carries this refund only.
	if event.RefundedFen != 0 {
		t.Errorf("refunded total must stay zero, got %s", event.RefundedFen)
	}
	if event.SuccessAt.IsZero() {
		t.Error("success time must parse")
	}
}

func TestNotifierAbnormalRefund(t *testing.T) {
	f, n := newTestNotifier(t)
	ni := mintNotification(t, f, "REFUND.ABNORMAL", "refund", map[string]any{
		"out_trade_no": "ord-1", "out_refund_no": "ref-2",
		"refund_id": "rf-8", "refund_status": "ABNORMAL", "amount": map[string]any{"refund": 60},
	})
	event, err := n.Parse(context.Background(), ni)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if event.Type != payment.EventRefundAbnormal || event.RefundState != payment.RefundStateAbnormal {
		t.Errorf("type/state: %+v", event)
	}
}

// A push with an unparseable success time still settles the state — the
// stamp degrades to zero instead of discarding a verified notification.
func TestNotifierBadSuccessTimeTolerated(t *testing.T) {
	f, n := newTestNotifier(t)
	ni := mintNotification(t, f, "TRANSACTION.SUCCESS", "transaction", map[string]any{
		"out_trade_no": "ord-1", "trade_state": "SUCCESS", "success_time": "not-a-time",
		"amount": map[string]any{"total": 100},
	})
	event, err := n.Parse(context.Background(), ni)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if event.State != payment.PaymentStatePaid || event.AmountFen != payment.Fen(100) {
		t.Errorf("event: %+v", event)
	}
	if !event.SuccessAt.IsZero() {
		t.Errorf("success time must degrade to zero, got %v", event.SuccessAt)
	}
}

// An event type this build has never met parses with a zero Type and the
// raw channel type riding ChannelType — the consumer audits.
func TestNotifierUnknownEventType(t *testing.T) {
	f, n := newTestNotifier(t)
	ni := mintNotification(t, f, "TRANSACTION.CLOSED", "transaction", map[string]any{
		"out_trade_no": "ord-1", "trade_state": "CLOSED", "amount": map[string]any{"total": 100},
	})
	event, err := n.Parse(context.Background(), ni)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if event.Type != "" || event.ChannelType != "TRANSACTION.CLOSED" || event.State != payment.PaymentStateClosed {
		t.Errorf("unknown event: %+v", event)
	}
}

// The refusal ladder keeps each probe single-failure: mutations that break
// the body get it re-signed first, so the signature gate passes and the seal
// gate is what fails.
func TestNotifierFailureLadder(t *testing.T) {
	paymentPayload := map[string]any{
		"out_trade_no": "ord-1", "trade_state": "SUCCESS", "amount": map[string]any{"total": 100},
	}
	type fix struct {
		mut  func(t *testing.T, f *fakeChannel, ni *payment.NotificationInput)
		want error
	}
	cases := map[string]fix{
		"tampered_body": {func(_ *testing.T, _ *fakeChannel, ni *payment.NotificationInput) {
			ni.Body = append(ni.Body, 'x')
		}, payment.ErrInvalidSignature},
		"missing_signature_header": {func(_ *testing.T, _ *fakeChannel, ni *payment.NotificationInput) {
			ni.Headers.Del("Wechatpay-Signature")
		}, payment.ErrInvalidSignature},
		"stale_timestamp": {func(_ *testing.T, _ *fakeChannel, ni *payment.NotificationInput) {
			ni.Headers.Set("Wechatpay-Timestamp", "1")
		}, payment.ErrInvalidSignature},
		"tampered_ciphertext": {func(t *testing.T, f *fakeChannel, ni *payment.NotificationInput) {
			var w map[string]any
			if err := json.Unmarshal(ni.Body, &w); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			res := w["resource"].(map[string]any)
			res["ciphertext"] = "AAAA" + res["ciphertext"].(string)
			body := []byte(mustMarshal(t, w))
			ni.Body, ni.Headers = body, f.pushHeaders(body, time.Now().Unix())
		}, payment.ErrDecryptFailed},
		"bad_nonce_length": {func(t *testing.T, f *fakeChannel, ni *payment.NotificationInput) {
			var w map[string]any
			if err := json.Unmarshal(ni.Body, &w); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			// 11 chars: the sealed envelope carries a wrong-length nonce, and
			// the decrypt step must refuse it rather than panic
			w["resource"].(map[string]any)["nonce"] = "nonce123456"
			body := []byte(mustMarshal(t, w))
			ni.Body, ni.Headers = body, f.pushHeaders(body, time.Now().Unix())
		}, payment.ErrDecryptFailed},
		"unknown_algorithm": {func(_ *testing.T, f *fakeChannel, ni *payment.NotificationInput) {
			body := []byte(`{"event_type":"TRANSACTION.SUCCESS","resource":{"original_type":"transaction","algorithm":"RSA-OAEP","nonce":"n","ciphertext":"c","associated_data":"a"}}`)
			ni.Body, ni.Headers = body, f.pushHeaders(body, time.Now().Unix())
		}, payment.ErrInvalidArgument},
		"malformed_envelope": {func(_ *testing.T, f *fakeChannel, ni *payment.NotificationInput) {
			body := []byte("not json")
			ni.Body, ni.Headers = body, f.pushHeaders(body, time.Now().Unix())
		}, payment.ErrInvalidArgument},
		"sealed_non_event": {func(t *testing.T, f *fakeChannel, ni *payment.NotificationInput) {
			*ni = *mintNotification(t, f, "TRANSACTION.SUCCESS", "transaction", "garbage")
		}, payment.ErrInvalidArgument},
		"sealed_non_refund": {func(t *testing.T, f *fakeChannel, ni *payment.NotificationInput) {
			*ni = *mintNotification(t, f, "REFUND.SUCCESS", "refund", 42)
		}, payment.ErrInvalidArgument},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			f, n := newTestNotifier(t)
			ni := mintNotification(t, f, "TRANSACTION.SUCCESS", "transaction", paymentPayload)
			// re-sign the pristine envelope the way the channel would, then
			// apply the single mutation
			c.mut(t, f, ni)
			event, err := n.Parse(context.Background(), ni)
			if !errors.Is(err, c.want) {
				t.Fatalf("want %v, got %v (event %+v)", c.want, err, event)
			}
		})
	}
}
