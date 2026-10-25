package mock

// Conformance probes over the mock backend: creation/query round-trips,
// channel-idempotent reuse, provider-knowledge enforcement, refund
// invariants, the notification parse ladder, duplicate-delivery stability,
// and the concurrency behaviors. The memory is the environment — the full
// list runs with no external setup.

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/capability/payment"
)

func input(channel payment.Channel, payer *payment.Payer, no string) *payment.CreatePaymentInput {
	return &payment.CreatePaymentInput{
		Channel:     channel,
		OutTradeNo:  no,
		AmountFen:   payment.Fen(30000),
		Description: "membership",
		NotifyURL:   "https://xgwl-edu.com/pay/notify",
		Payer:       payer,
	}
}

func jsapi(no string) *payment.CreatePaymentInput {
	return input(ChannelJSAPI, &payment.Payer{OpenID: "oX-1"}, no)
}

func TestCreateQueryRoundTrip(t *testing.T) {
	p, _ := Open()
	cases := []struct {
		channel payment.Channel
		payer   *payment.Payer
		kind    payment.CredentialKind
	}{
		{ChannelJSAPI, &payment.Payer{OpenID: "oX-1"}, payment.CredentialInvoke},
		{ChannelH5, &payment.Payer{ClientIP: "203.0.113.7"}, payment.CredentialJumpURL},
		{ChannelNative, nil, payment.CredentialQRCode},
	}
	for i, c := range cases {
		no := fmt.Sprintf("ord-round-%d", i)
		res, err := p.Gateway.CreatePayment(context.Background(), input(c.channel, c.payer, no))
		if err != nil {
			t.Fatalf("create %s: %v", c.channel, err)
		}
		if res.Credential.Kind != c.kind {
			t.Errorf("%s: credential kind %q, want %q", c.channel, res.Credential.Kind, c.kind)
		}
		if res.Credential.Value == "" {
			t.Errorf("%s: empty credential value", c.channel)
		}
		status, err := p.Gateway.QueryPayment(context.Background(), no)
		if err != nil || status.State != payment.PaymentStatePending {
			t.Fatalf("query %s: state=%v err=%v", c.channel, status.State, err)
		}
		if status.AmountFen != payment.Fen(30000) {
			t.Errorf("%s: amount %s, want 300.00", c.channel, status.AmountFen)
		}
	}
}

func TestCreateIdempotentReuse(t *testing.T) {
	p, _ := Open()
	first, err := p.Gateway.CreatePayment(context.Background(), jsapi("ord-idem"))
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	second, err := p.Gateway.CreatePayment(context.Background(), jsapi("ord-idem"))
	if err != nil {
		t.Fatalf("reuse: %v", err)
	}
	if first.Credential != second.Credential {
		t.Errorf("reuse must return the same credential: %v vs %v", first.Credential, second.Credential)
	}
	// Nil-Payer inputs (native) compare equal too.
	n1, err := p.Gateway.CreatePayment(context.Background(), input(ChannelNative, nil, "ord-idem-n"))
	if err != nil {
		t.Fatalf("native first: %v", err)
	}
	n2, err := p.Gateway.CreatePayment(context.Background(), input(ChannelNative, nil, "ord-idem-n"))
	if err != nil {
		t.Fatalf("native reuse: %v", err)
	}
	if n1.Credential != n2.Credential {
		t.Errorf("native reuse must return the same credential: %v vs %v", n1.Credential, n2.Credential)
	}
}

// TestNewSmoke pins the production-shaped constructor: the container it
// returns is fully populated and behaves, and the declared channel set is
// the mirrored three.
func TestNewSmoke(t *testing.T) {
	p := New()
	if p.Gateway == nil || p.Notifier == nil || p.Reports == nil {
		t.Fatal("New() must populate every container field")
	}
	if _, err := p.Gateway.CreatePayment(context.Background(), jsapi("ord-new")); err != nil {
		t.Fatalf("create via New(): %v", err)
	}
	channels := Channels()
	want := []payment.Channel{ChannelJSAPI, ChannelH5, ChannelNative}
	if len(channels) != len(want) {
		t.Fatalf("declared channels: %v, want %v", channels, want)
	}
	for i := range want {
		if channels[i] != want[i] {
			t.Errorf("declared channel %d: %s, want %s", i, channels[i], want[i])
		}
	}
}

func TestCreateDuplicateRefusal(t *testing.T) {
	p, _ := Open()
	if _, err := p.Gateway.CreatePayment(context.Background(), jsapi("ord-dup")); err != nil {
		t.Fatalf("first: %v", err)
	}
	changed := jsapi("ord-dup")
	changed.AmountFen = payment.Fen(99999)
	if _, err := p.Gateway.CreatePayment(context.Background(), changed); !errors.Is(err, payment.ErrDuplicate) {
		t.Errorf("same key different params: want ErrDuplicate, got %v", err)
	}
}

func TestChannelEnforcement(t *testing.T) {
	p, _ := Open()
	if _, err := p.Gateway.CreatePayment(context.Background(), input(payment.Channel("alipay_wap"), nil, "ord-x")); !errors.Is(err, payment.ErrInvalidArgument) {
		t.Errorf("unsupported channel: want ErrInvalidArgument, got %v", err)
	}
	noOpenID := input(ChannelJSAPI, &payment.Payer{}, "ord-x2")
	if _, err := p.Gateway.CreatePayment(context.Background(), noOpenID); !errors.Is(err, payment.ErrInvalidArgument) {
		t.Errorf("jsapi without openid: want ErrInvalidArgument, got %v", err)
	}
	noIP := input(ChannelH5, &payment.Payer{OpenID: "oX"}, "ord-x3")
	if _, err := p.Gateway.CreatePayment(context.Background(), noIP); !errors.Is(err, payment.ErrInvalidArgument) {
		t.Errorf("h5 without client ip: want ErrInvalidArgument, got %v", err)
	}
}

func TestNeutralValidationPassesThrough(t *testing.T) {
	p, _ := Open()
	zero := jsapi("ord-n1")
	zero.AmountFen = 0
	if _, err := p.Gateway.CreatePayment(context.Background(), zero); !errors.Is(err, payment.ErrInvalidArgument) {
		t.Errorf("zero amount: want ErrInvalidArgument, got %v", err)
	}
	bad := jsapi("坏-单号")
	if _, err := p.Gateway.CreatePayment(context.Background(), bad); !errors.Is(err, payment.ErrInvalidArgument) {
		t.Errorf("bad trade number: want ErrInvalidArgument, got %v", err)
	}
}

func TestForcedRejection(t *testing.T) {
	p, ctrl := Open()
	ctrl.RejectWith("ord-rej", "NOTENOUGH", "not enough balance")
	_, err := p.Gateway.CreatePayment(context.Background(), jsapi("ord-rej"))
	if !errors.Is(err, payment.ErrChannelRejected) {
		t.Fatalf("want ErrChannelRejected, got %v", err)
	}
	var ce *payment.ChannelError
	if !errors.As(err, &ce) || ce.Code != "NOTENOUGH" {
		t.Errorf("channel detail lost: %v", err)
	}
}

func TestMarkPaidFlow(t *testing.T) {
	p, ctrl := Open()
	if _, err := p.Gateway.CreatePayment(context.Background(), jsapi("ord-paid")); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := ctrl.MarkPaid("ord-paid"); err != nil {
		t.Fatalf("mark paid: %v", err)
	}
	status, err := p.Gateway.QueryPayment(context.Background(), "ord-paid")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if status.State != payment.PaymentStatePaid || status.TransactionID == "" || status.PaidAt.IsZero() || status.RefundedFen != 0 {
		t.Errorf("paid status malformed: %+v", status)
	}
	// State-machine refusals: unknown order, closed order.
	if _, err := ctrl.MarkPaid("ord-none"); !errors.Is(err, payment.ErrNotFound) {
		t.Errorf("mark paid unknown order: want ErrNotFound, got %v", err)
	}
	if _, err := p.Gateway.QueryPayment(context.Background(), "ord-none"); !errors.Is(err, payment.ErrNotFound) {
		t.Errorf("query unknown order: want ErrNotFound, got %v", err)
	}
	if _, err := p.Gateway.CreatePayment(context.Background(), jsapi("ord-paid-closed")); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := p.Gateway.ClosePayment(context.Background(), "ord-paid-closed"); err != nil {
		t.Fatalf("close: %v", err)
	}
	if _, err := ctrl.MarkPaid("ord-paid-closed"); !errors.Is(err, payment.ErrChannelRejected) {
		t.Errorf("mark paid closed order: want ErrChannelRejected, got %v", err)
	}
	if _, err := ctrl.RefundNotification("ref-none"); !errors.Is(err, payment.ErrNotFound) {
		t.Errorf("refund notification unknown refund: want ErrNotFound, got %v", err)
	}
}

func TestPaymentNotificationParse(t *testing.T) {
	p, ctrl := Open()
	if _, err := p.Gateway.CreatePayment(context.Background(), jsapi("ord-notify")); err != nil {
		t.Fatalf("create: %v", err)
	}
	ni, err := ctrl.MarkPaid("ord-notify")
	if err != nil {
		t.Fatalf("mark paid: %v", err)
	}
	event, err := p.Notifier.Parse(context.Background(), ni)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if event.Type != payment.EventPaymentSucceeded ||
		event.ChannelType != "MOCK.TRANSACTION.SUCCESS" ||
		event.OutTradeNo != "ord-notify" ||
		event.State != payment.PaymentStatePaid ||
		event.AmountFen != payment.Fen(30000) ||
		event.TransactionID == "" ||
		event.SuccessAt.IsZero() {
		t.Errorf("event fields malformed: %+v", event)
	}
}

func TestDuplicateDeliveryStable(t *testing.T) {
	_, ctrl := Open()
	if _, err := ctrl.b.CreatePayment(context.Background(), jsapi("ord-dupn")); err != nil {
		t.Fatalf("create: %v", err)
	}
	first, err := ctrl.MarkPaid("ord-dupn")
	if err != nil {
		t.Fatalf("mark paid: %v", err)
	}
	second, err := ctrl.MarkPaid("ord-dupn")
	if err != nil {
		t.Fatalf("mark paid again: %v", err)
	}
	if !bytes.Equal(first.Body, second.Body) {
		t.Error("redelivery must carry identical bytes, not a remint")
	}
	if first.Headers.Get(signatureHeader) != second.Headers.Get(signatureHeader) {
		t.Error("redelivery signature changed")
	}
}

func TestParseFailureLadder(t *testing.T) {
	p, ctrl := Open()
	if _, err := ctrl.b.CreatePayment(context.Background(), jsapi("ord-ladder")); err != nil {
		t.Fatalf("create: %v", err)
	}
	ni, err := ctrl.MarkPaid("ord-ladder")
	if err != nil {
		t.Fatalf("mark paid: %v", err)
	}
	noSig := &payment.NotificationInput{Headers: http.Header{}, Body: ni.Body}
	if _, err := p.Notifier.Parse(context.Background(), noSig); !errors.Is(err, payment.ErrInvalidSignature) {
		t.Errorf("missing signature: want ErrInvalidSignature, got %v", err)
	}
	badSig := &payment.NotificationInput{Headers: http.Header{}, Body: ni.Body}
	badSig.Headers.Set(signatureHeader, base64.StdEncoding.EncodeToString([]byte("forged")))
	if _, err := p.Notifier.Parse(context.Background(), badSig); !errors.Is(err, payment.ErrInvalidSignature) {
		t.Errorf("forged signature: want ErrInvalidSignature, got %v", err)
	}
	// Correct signature, malformed envelope.
	envelopeBody, _ := json.Marshal(map[string]any{"v": 1, "payload": "?", "nonce": "?"})
	signedMalformed := &payment.NotificationInput{Headers: http.Header{}, Body: envelopeBody}
	signedMalformed.Headers.Set(signatureHeader, ctrl.b.sign(envelopeBody))
	if _, err := p.Notifier.Parse(context.Background(), signedMalformed); !errors.Is(err, payment.ErrInvalidArgument) {
		t.Errorf("malformed envelope: want ErrInvalidArgument, got %v", err)
	}
	// Correct signature and envelope, undecryptable payload.
	garbage, _ := json.Marshal(wireEnvelope{
		V:       wireVersion,
		Nonce:   base64.StdEncoding.EncodeToString(make([]byte, 12)),
		Payload: base64.StdEncoding.EncodeToString([]byte("garbage ciphertext")),
	})
	signedGarbage := &payment.NotificationInput{Headers: http.Header{}, Body: garbage}
	signedGarbage.Headers.Set(signatureHeader, ctrl.b.sign(garbage))
	if _, err := p.Notifier.Parse(context.Background(), signedGarbage); !errors.Is(err, payment.ErrDecryptFailed) {
		t.Errorf("undecryptable payload: want ErrDecryptFailed, got %v", err)
	}
}

func TestCloseLifecycle(t *testing.T) {
	p, ctrl := Open()
	if _, err := p.Gateway.CreatePayment(context.Background(), jsapi("ord-close")); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := p.Gateway.ClosePayment(context.Background(), "ord-close"); err != nil {
		t.Fatalf("close pending: %v", err)
	}
	if status, _ := p.Gateway.QueryPayment(context.Background(), "ord-close"); status.State != payment.PaymentStateClosed {
		t.Errorf("state after close: %v", status.State)
	}
	if err := p.Gateway.ClosePayment(context.Background(), "ord-close"); err != nil {
		t.Errorf("re-close must be idempotent: %v", err)
	}
	if _, err := p.Gateway.CreatePayment(context.Background(), jsapi("ord-close2")); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := ctrl.MarkPaid("ord-close2"); err != nil {
		t.Fatalf("mark paid: %v", err)
	}
	if err := p.Gateway.ClosePayment(context.Background(), "ord-close2"); !errors.Is(err, payment.ErrChannelRejected) {
		t.Errorf("close paid: want ErrChannelRejected, got %v", err)
	}
}

func TestRefundFlow(t *testing.T) {
	p, ctrl := Open()
	if _, err := p.Gateway.CreatePayment(context.Background(), jsapi("ord-refund")); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := ctrl.MarkPaid("ord-refund"); err != nil {
		t.Fatalf("mark paid: %v", err)
	}
	half, err := p.Gateway.Refund(context.Background(), &payment.RefundInput{
		OutRefundNo: "ref-1", OutTradeNo: "ord-refund", AmountFen: payment.Fen(15000), TotalAmountFen: payment.Fen(30000), Reason: "half",
	})
	if err != nil || half.State != payment.RefundStateSucceeded {
		t.Fatalf("first refund: %+v err=%v", half, err)
	}
	status, _ := p.Gateway.QueryPayment(context.Background(), "ord-refund")
	if status.RefundedFen != payment.Fen(15000) {
		t.Errorf("refunded after first: %s, want 150.00", status.RefundedFen)
	}
	if _, err := p.Gateway.Refund(context.Background(), &payment.RefundInput{
		OutRefundNo: "ref-2", OutTradeNo: "ord-refund", AmountFen: payment.Fen(15000), TotalAmountFen: payment.Fen(30000),
	}); err != nil {
		t.Fatalf("second refund: %v", err)
	}
	over, err := p.Gateway.Refund(context.Background(), &payment.RefundInput{
		OutRefundNo: "ref-3", OutTradeNo: "ord-refund", AmountFen: payment.Fen(1), TotalAmountFen: payment.Fen(30000),
	})
	if !errors.Is(err, payment.ErrChannelRejected) || over != nil {
		t.Errorf("over-budget refund must be a channel rejection, got %+v err=%v", over, err)
	}
	var ce *payment.ChannelError
	if !errors.As(err, &ce) || ce.Code != "REFUND_EXCEED_BALANCE" {
		t.Errorf("over-budget code: %v", err)
	}
	// A declared original amount that contradicts the stored order is an
	// input fault, not a channel verdict.
	if _, err := p.Gateway.Refund(context.Background(), &payment.RefundInput{
		OutRefundNo: "ref-wrongtotal", OutTradeNo: "ord-refund", AmountFen: payment.Fen(1), TotalAmountFen: payment.Fen(29999),
	}); !errors.Is(err, payment.ErrInvalidArgument) {
		t.Errorf("contradicting declared total: want ErrInvalidArgument, got %v", err)
	}
	reuse, err := p.Gateway.Refund(context.Background(), &payment.RefundInput{
		OutRefundNo: "ref-1", OutTradeNo: "ord-refund", AmountFen: payment.Fen(15000), TotalAmountFen: payment.Fen(30000), Reason: "half",
	})
	if err != nil || reuse.RefundID != half.RefundID {
		t.Errorf("refund idempotent reuse: %+v err=%v", reuse, err)
	}
	changed, err := p.Gateway.Refund(context.Background(), &payment.RefundInput{
		OutRefundNo: "ref-1", OutTradeNo: "ord-refund", AmountFen: payment.Fen(100), TotalAmountFen: payment.Fen(30000),
	})
	if !errors.Is(err, payment.ErrDuplicate) || changed != nil {
		t.Errorf("same refund key different params: want ErrDuplicate, got %v", err)
	}
	qr, err := p.Gateway.QueryRefund(context.Background(), "ref-1")
	if err != nil || qr.State != payment.RefundStateSucceeded || qr.AmountFen != payment.Fen(15000) {
		t.Errorf("query refund: %+v err=%v", qr, err)
	}
	if _, err := p.Gateway.QueryRefund(context.Background(), "ref-none"); !errors.Is(err, payment.ErrNotFound) {
		t.Errorf("unknown refund: want ErrNotFound, got %v", err)
	}
	// Refunding a pending order is a channel rejection, not an input error.
	if _, err := p.Gateway.CreatePayment(context.Background(), jsapi("ord-pending")); err != nil {
		t.Fatalf("create pending: %v", err)
	}
	if _, err := p.Gateway.Refund(context.Background(), &payment.RefundInput{
		OutRefundNo: "ref-p", OutTradeNo: "ord-pending", AmountFen: payment.Fen(1), TotalAmountFen: payment.Fen(30000),
	}); !errors.Is(err, payment.ErrChannelRejected) {
		t.Errorf("refund pending order: want ErrChannelRejected, got %v", err)
	}
}

func TestRefundNotificationParse(t *testing.T) {
	p, ctrl := Open()
	if _, err := p.Gateway.CreatePayment(context.Background(), jsapi("ord-rnotify")); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := ctrl.MarkPaid("ord-rnotify"); err != nil {
		t.Fatalf("mark paid: %v", err)
	}
	if _, err := p.Gateway.Refund(context.Background(), &payment.RefundInput{
		OutRefundNo: "ref-n1", OutTradeNo: "ord-rnotify", AmountFen: payment.Fen(10000), TotalAmountFen: payment.Fen(30000),
	}); err != nil {
		t.Fatalf("refund: %v", err)
	}
	ni, err := ctrl.RefundNotification("ref-n1")
	if err != nil {
		t.Fatalf("refund notification: %v", err)
	}
	event, err := p.Notifier.Parse(context.Background(), ni)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if event.Type != payment.EventRefundSucceeded ||
		event.OutRefundNo != "ref-n1" ||
		event.AmountFen != payment.Fen(10000) ||
		event.RefundedFen != payment.Fen(10000) ||
		event.RefundState != payment.RefundStateSucceeded {
		t.Errorf("refund event fields malformed: %+v", event)
	}
}

func TestConcurrentCreateSameKey(t *testing.T) {
	p, _ := Open()
	var wg sync.WaitGroup
	credentials := make([]payment.Credential, 16)
	errs := make([]error, 16)
	for i := range credentials {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			res, err := p.Gateway.CreatePayment(context.Background(), jsapi("ord-conc"))
			if err == nil {
				credentials[i] = res.Credential
			}
			errs[i] = err
		}(i)
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("concurrent create %d: %v", i, err)
		}
		if credentials[i] != credentials[0] {
			t.Fatalf("concurrent same-key creates diverged: %v vs %v", credentials[i], credentials[0])
		}
	}
}

func TestConcurrentRefundBudget(t *testing.T) {
	p, ctrl := Open()
	if _, err := p.Gateway.CreatePayment(context.Background(), jsapi("ord-budget")); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := ctrl.MarkPaid("ord-budget"); err != nil {
		t.Fatalf("mark paid: %v", err)
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	succeeded, rejected := 0, 0
	for i := range 3 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := p.Gateway.Refund(context.Background(), &payment.RefundInput{
				OutRefundNo:    "ref-b" + strconv.Itoa(i),
				OutTradeNo:     "ord-budget",
				AmountFen:      payment.Fen(12000),
				TotalAmountFen: payment.Fen(30000),
			})
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				succeeded++
			} else if errors.Is(err, payment.ErrChannelRejected) {
				rejected++
			}
		}(i)
	}
	wg.Wait()
	if succeeded != 2 || rejected != 1 {
		t.Errorf("refund budget under concurrency: succeeded=%d rejected=%d, want 2/1", succeeded, rejected)
	}
	status, _ := p.Gateway.QueryPayment(context.Background(), "ord-budget")
	if status.RefundedFen != payment.Fen(24000) {
		t.Errorf("refunded total: %s, want 240.00", status.RefundedFen)
	}
}

// TestDeliver proves the minted notification travels real HTTP with its
// signature header, repeat times — duplicates a consumer's callback
// endpoint must survive.
func TestDeliver(t *testing.T) {
	_, ctrl := Open()
	if _, err := ctrl.b.CreatePayment(context.Background(), jsapi("ord-deliver")); err != nil {
		t.Fatalf("create: %v", err)
	}
	ni, err := ctrl.MarkPaid("ord-deliver")
	if err != nil {
		t.Fatalf("mark paid: %v", err)
	}
	var mu sync.Mutex
	got := 0
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		_ = r.Body.Close()
		mu.Lock()
		got++
		mu.Unlock()
	}))
	defer srv.Close()
	if err := ctrl.Deliver(ni, srv.URL, 3, 0); err != nil {
		t.Fatalf("deliver: %v", err)
	}
	if got != 3 {
		t.Errorf("repeat deliveries: got %d, want 3", got)
	}
	if err := ctrl.Deliver(ni, "http://127.0.0.1:1/unreachable", 1, 0); err == nil {
		t.Error("unreachable endpoint must error")
	}
	if err := ctrl.Deliver(ni, "http://bad url\x7f", 1, 0); err == nil {
		t.Error("unparseable url must error")
	}
}

func TestBill(t *testing.T) {
	p, ctrl := Open()
	if _, err := p.Gateway.CreatePayment(context.Background(), jsapi("ord-bill")); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := ctrl.MarkPaid("ord-bill"); err != nil {
		t.Fatalf("mark paid: %v", err)
	}
	if _, err := p.Reports.DownloadBill(context.Background(), &payment.BillInput{Kind: payment.BillKindTrade}); !errors.Is(err, payment.ErrInvalidArgument) {
		t.Errorf("zero date: want ErrInvalidArgument, got %v", err)
	}
	now := time.Now()
	trade, err := p.Reports.DownloadBill(context.Background(), &payment.BillInput{Date: now, Kind: payment.BillKindTrade})
	if err != nil {
		t.Fatalf("trade bill: %v", err)
	}
	if !bytes.Contains(trade.Raw, []byte("ord-bill")) || !bytes.Contains(trade.Raw, []byte("paid")) {
		t.Errorf("trade bill content: %s", trade.Raw)
	}
	fund, err := p.Reports.DownloadBill(context.Background(), &payment.BillInput{Date: now, Kind: payment.BillKindFundFlow})
	if err != nil {
		t.Fatalf("fund bill: %v", err)
	}
	if !bytes.Contains(fund.Raw, []byte("ord-bill")) {
		t.Errorf("fund bill content: %s", fund.Raw)
	}
	// A pending-only order appears in the trade bill but not the fund flow.
	if _, err := p.Gateway.CreatePayment(context.Background(), jsapi("ord-bill-pending")); err != nil {
		t.Fatalf("create pending: %v", err)
	}
	trade2, err := p.Reports.DownloadBill(context.Background(), &payment.BillInput{Date: now, Kind: payment.BillKindTrade})
	if err != nil {
		t.Fatalf("trade bill 2: %v", err)
	}
	if !bytes.Contains(trade2.Raw, []byte("ord-bill-pending")) {
		t.Errorf("trade bill must include pending orders: %s", trade2.Raw)
	}
	fund2, err := p.Reports.DownloadBill(context.Background(), &payment.BillInput{Date: now, Kind: payment.BillKindFundFlow})
	if err != nil {
		t.Fatalf("fund bill 2: %v", err)
	}
	if bytes.Contains(fund2.Raw, []byte("ord-bill-pending")) {
		t.Errorf("fund flow must skip unsettled orders: %s", fund2.Raw)
	}
}

func TestBillParse(t *testing.T) {
	p, ctrl := Open()
	if _, err := p.Gateway.CreatePayment(context.Background(), jsapi("ord-bp")); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := ctrl.MarkPaid("ord-bp"); err != nil {
		t.Fatalf("mark paid: %v", err)
	}
	if _, err := p.Gateway.Refund(context.Background(), &payment.RefundInput{
		OutRefundNo: "rf-bp", OutTradeNo: "ord-bp", AmountFen: 100, TotalAmountFen: 30000,
	}); err != nil {
		t.Fatalf("refund: %v", err)
	}
	bill, err := p.Reports.DownloadBill(context.Background(), &payment.BillInput{Date: time.Now(), Kind: payment.BillKindTrade})
	if err != nil {
		t.Fatalf("bill: %v", err)
	}
	lines, err := p.Reports.ParseTradeBill(context.Background(), bill.Raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(lines) != 1 {
		t.Fatalf("lines: want 1, got %d", len(lines))
	}
	l := lines[0]
	if l.OutTradeNo != "ord-bp" || l.State != payment.BillLinePaid {
		t.Errorf("line: %+v", l)
	}
	if int64(l.AmountFen) != 30000 || int64(l.RefundFen) != 100 {
		t.Errorf("amounts: %+v", l)
	}
	if _, err := p.Reports.ParseTradeBill(context.Background(), []byte("a,b\n1,2\n")); !errors.Is(err, payment.ErrInvalidArgument) {
		t.Errorf("malformed: want ErrInvalidArgument, got %v", err)
	}
}

func TestBillParseErrorArms(t *testing.T) {
	p, _ := Open()
	ctx := context.Background()
	parse := func(raw string) error {
		_, err := p.Reports.ParseTradeBill(ctx, []byte(raw))
		return err
	}
	if err := parse("out_trade_no,state,amount_fen,refunded_fen,paid_at\nord-1,paid,1990\n"); !errors.Is(err, payment.ErrInvalidArgument) {
		t.Errorf("row shape: %v", err)
	}
	if err := parse("out_trade_no,state,amount_fen,refunded_fen,paid_at\nord-1,paid,x,0,0001-01-01T00:00:00Z\n"); !errors.Is(err, payment.ErrInvalidArgument) {
		t.Errorf("amount: %v", err)
	}
	if err := parse("out_trade_no,state,amount_fen,refunded_fen,paid_at\nord-1,paid,1990,x,0001-01-01T00:00:00Z\n"); !errors.Is(err, payment.ErrInvalidArgument) {
		t.Errorf("refunded: %v", err)
	}
	if err := parse("out_trade_no,state,amount_fen,refunded_fen,paid_at\nord-1,paid,1990,0,not-a-time\n"); !errors.Is(err, payment.ErrInvalidArgument) {
		t.Errorf("time: %v", err)
	}
	if err := parse("out_trade_no,state,amount_fen,refunded_fen,paid_at\nord-1,weird,1990,0,0001-01-01T00:00:00Z\n"); !errors.Is(err, payment.ErrInvalidArgument) {
		t.Errorf("state: %v", err)
	}
	// A trailing newline is an empty row, not a malformed one.
	if _, err := p.Reports.ParseTradeBill(ctx, []byte("out_trade_no,state,amount_fen,refunded_fen,paid_at\n")); err != nil {
		t.Errorf("header only: %v", err)
	}
}

func TestParseMalformedEnvelopes(t *testing.T) {
	p, ctrl := Open()
	if _, err := p.Gateway.CreatePayment(context.Background(), jsapi("ord-env")); err != nil {
		t.Fatalf("create: %v", err)
	}
	ni, err := ctrl.MarkPaid("ord-env")
	if err != nil {
		t.Fatalf("mark paid: %v", err)
	}
	ctx := context.Background()
	// The genuine envelope replays clean, then each corruption folds onto its sentinel.
	if _, err := p.Notifier.Parse(ctx, ni); err != nil {
		t.Fatalf("genuine: %v", err)
	}
	if _, err := p.Notifier.Parse(ctx, &payment.NotificationInput{Headers: ni.Headers, Body: []byte("not json")}); !errors.Is(err, payment.ErrInvalidSignature) {
		t.Errorf("tampered body: want ErrInvalidSignature, got %v", err)
	}
	tampered := ni.Headers.Clone()
	tampered.Set("X-Mockpay-Signature", "deadbeef")
	if _, err := p.Notifier.Parse(ctx, &payment.NotificationInput{Headers: tampered, Body: ni.Body}); !errors.Is(err, payment.ErrInvalidSignature) {
		t.Errorf("forged signature: want ErrInvalidSignature, got %v", err)
	}
	badNonce := []byte(`{"v":1,"nonce":"!!!","payload":"AAAA"}`)
	forged := http.Header{}
	forged.Set("X-Mockpay-Signature", "x")
	ni2 := &payment.NotificationInput{Headers: forged, Body: badNonce}
	if _, err := p.Notifier.Parse(ctx, ni2); !errors.Is(err, payment.ErrInvalidSignature) {
		t.Errorf("signed garbage: want ErrInvalidSignature, got %v", err)
	}
}

func TestControllerOf(t *testing.T) {
	p, _ := Open()
	ctrl, ok := ControllerOf(p)
	if !ok || ctrl == nil {
		t.Fatalf("mock container: ok=%v ctrl=%v", ok, ctrl)
	}
	// The dev face exists only over a container this package built — a
	// foreign or half-built container never yields one.
	if _, ok := ControllerOf(&payment.Payment{}); ok {
		t.Error("foreign container must not yield a controller")
	}
}
