package wechat

// Gateway probes against the fake channel: the wire each access context
// produces (payer for JSAPI, scene for H5, bare body for Native), the state
// folding from the channel's vocabulary onto the contract's, and the local
// refusals that never touch the network.

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/capability/payment"
)

func newTestGateway(t *testing.T) (*fakeChannel, payment.Gateway) {
	t.Helper()
	f := newFakeChannel(t)
	c, err := newClient(f.publicKeyConfig(t))
	if err != nil {
		t.Fatalf("newClient: %v", err)
	}
	return f, &gateway{c: c, appID: testAppID}
}

func jsonMap(t *testing.T, body string) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(body), &m); err != nil {
		t.Fatalf("decode request body %q: %v", body, err)
	}
	return m
}

func wantString(t *testing.T, m map[string]any, key, want string) {
	t.Helper()
	if got, _ := m[key].(string); got != want {
		t.Errorf("%s: want %q, got %v", key, want, m[key])
	}
}

func TestChannelsDeclaration(t *testing.T) {
	got := Channels()
	want := []payment.Channel{"wechat_jsapi", "wechat_h5", "wechat_native"}
	if len(got) != len(want) {
		t.Fatalf("Channels(): %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Channels()[%d]: want %s, got %s", i, want[i], got[i])
		}
	}
}

func TestGatewayCreatePayment(t *testing.T) {
	base := func() *payment.CreatePaymentInput {
		return &payment.CreatePaymentInput{
			Channel:     ChannelJSAPI,
			OutTradeNo:  "ord-1",
			AmountFen:   payment.Fen(100),
			Description: "membership",
			NotifyURL:   "https://xgwl-edu.com/pay/notify",
			ExpireAt:    time.Now().Add(time.Hour).Truncate(time.Second),
		}
	}

	t.Run("jsapi", func(t *testing.T) {
		f, g := newTestGateway(t)
		f.route(http.MethodPost, "/v3/pay/transactions/jsapi", http.StatusOK, `{"prepay_id":"wx01"}`)
		in := base()
		in.Payer = &payment.Payer{OpenID: "oX-123"}
		res, err := g.CreatePayment(context.Background(), in)
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if res.Credential != (payment.Credential{Kind: payment.CredentialInvoke, Value: "wx01"}) {
			t.Errorf("credential: %+v", res.Credential)
		}
		m := jsonMap(t, f.lastRequest(t, http.MethodPost, "/v3/pay/transactions/jsapi").Body)
		wantString(t, m, "appid", testAppID)
		wantString(t, m, "mchid", testMchID)
		wantString(t, m, "out_trade_no", "ord-1")
		wantString(t, m, "notify_url", in.NotifyURL)
		amount, _ := m["amount"].(map[string]any)
		if amount == nil || amount["total"] != float64(100) {
			t.Errorf("amount: %v", m["amount"])
		}
		payer, _ := m["payer"].(map[string]any)
		if payer == nil || payer["openid"] != "oX-123" {
			t.Errorf("payer: %v", m["payer"])
		}
		expire, _ := m["time_expire"].(string)
		if !strings.Contains(expire, "T") {
			t.Errorf("time_expire must be RFC3339, got %q", expire)
		}
	})

	t.Run("h5", func(t *testing.T) {
		f, g := newTestGateway(t)
		f.route(http.MethodPost, "/v3/pay/transactions/h5", http.StatusOK, `{"h5_url":"https://wx.tenpay.com/v1/jump"}`)
		in := base()
		in.Channel = ChannelH5
		in.Payer = &payment.Payer{ClientIP: "203.0.113.7"}
		res, err := g.CreatePayment(context.Background(), in)
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if res.Credential != (payment.Credential{Kind: payment.CredentialJumpURL, Value: "https://wx.tenpay.com/v1/jump"}) {
			t.Errorf("credential: %+v", res.Credential)
		}
		m := jsonMap(t, f.lastRequest(t, http.MethodPost, "/v3/pay/transactions/h5").Body)
		if _, has := m["payer"]; has {
			t.Error("h5 wire must not carry payer")
		}
		scene, _ := m["scene_info"].(map[string]any)
		if scene == nil || scene["payer_client_ip"] != "203.0.113.7" {
			t.Errorf("scene_info: %v", m["scene_info"])
		}
		h5, _ := scene["h5_info"].(map[string]any)
		if h5 == nil || h5["type"] != "Wap" {
			t.Errorf("h5_info: %v", scene["h5_info"])
		}
	})

	t.Run("native", func(t *testing.T) {
		f, g := newTestGateway(t)
		f.route(http.MethodPost, "/v3/pay/transactions/native", http.StatusOK, `{"code_url":"weixin://wxpay/bizpayurl?pr=abc"}`)
		in := base()
		in.Channel = ChannelNative
		in.Payer = nil
		res, err := g.CreatePayment(context.Background(), in)
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if res.Credential != (payment.Credential{Kind: payment.CredentialQRCode, Value: "weixin://wxpay/bizpayurl?pr=abc"}) {
			t.Errorf("credential: %+v", res.Credential)
		}
		m := jsonMap(t, f.lastRequest(t, http.MethodPost, "/v3/pay/transactions/native").Body)
		if _, has := m["payer"]; has {
			t.Error("native wire must not carry payer")
		}
		if _, has := m["scene_info"]; has {
			t.Error("native wire must not carry scene_info")
		}
		if _, has := m["time_expire"]; !has {
			t.Error("a set ExpireAt must ride the wire as time_expire")
		}
	})
}

func TestGatewayCreatePaymentRefusals(t *testing.T) {
	cases := []struct {
		name string
		mut  func(*payment.CreatePaymentInput)
	}{
		{"unknown_channel", func(in *payment.CreatePaymentInput) { in.Channel = payment.Channel("alipay_wap") }},
		{"jsapi_missing_openid", func(in *payment.CreatePaymentInput) { in.Payer = &payment.Payer{} }},
		{"h5_missing_ip", func(in *payment.CreatePaymentInput) {
			in.Channel = ChannelH5
			in.Payer = &payment.Payer{OpenID: "irrelevant"}
		}},
		{"zero_amount", func(in *payment.CreatePaymentInput) { in.AmountFen = 0 }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f, g := newTestGateway(t)
			in := &payment.CreatePaymentInput{
				Channel:     ChannelJSAPI,
				OutTradeNo:  "ord-x",
				AmountFen:   payment.Fen(100),
				Description: "membership",
				NotifyURL:   "https://xgwl-edu.com/pay/notify",
				Payer:       &payment.Payer{OpenID: "oX-123"},
			}
			c.mut(in)
			if _, err := g.CreatePayment(context.Background(), in); !errors.Is(err, payment.ErrInvalidArgument) {
				t.Fatalf("want ErrInvalidArgument, got %v", err)
			}
			if len(f.requests) != 0 {
				t.Errorf("refusal must be local, but %d requests left", len(f.requests))
			}
		})
	}
}

func TestGatewayQueryPayment(t *testing.T) {
	cases := []struct {
		name       string
		body       string
		wantState  payment.PaymentState
		wantPaid   bool
		wantRefund payment.AmountFen
	}{
		{"success", `{"transaction_id":"tx1","trade_state":"SUCCESS","success_time":"2026-10-25T12:00:00+08:00","amount":{"total":100}}`, payment.PaymentStatePaid, true, 0},
		{"refund_stays_paid", `{"transaction_id":"tx1","trade_state":"REFUND","success_time":"2026-10-25T12:00:00+08:00","amount":{"total":100,"refund_total":40}}`, payment.PaymentStatePaid, true, payment.Fen(40)},
		{"notpay", `{"trade_state":"NOTPAY","amount":{"total":100}}`, payment.PaymentStatePending, false, 0},
		{"userpaying", `{"trade_state":"USERPAYING","amount":{"total":100}}`, payment.PaymentStatePending, false, 0},
		{"closed", `{"trade_state":"CLOSED","amount":{"total":100}}`, payment.PaymentStateClosed, false, 0},
		{"revoked", `{"trade_state":"REVOKED","amount":{"total":100}}`, payment.PaymentStateClosed, false, 0},
		{"payerror", `{"trade_state":"PAYERROR","amount":{"total":100}}`, payment.PaymentStateClosed, false, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f, g := newTestGateway(t)
			f.route(http.MethodGet, "/v3/pay/transactions/out-trade-no/ord-1", http.StatusOK, c.body)
			res, err := g.QueryPayment(context.Background(), "ord-1")
			if err != nil {
				t.Fatalf("query: %v", err)
			}
			if res.State != c.wantState {
				t.Errorf("state: want %s, got %s", c.wantState, res.State)
			}
			if res.AmountFen != payment.Fen(100) {
				t.Errorf("amount: %s", res.AmountFen)
			}
			if res.RefundedFen != c.wantRefund {
				t.Errorf("refunded: want %s, got %s", c.wantRefund, res.RefundedFen)
			}
			if c.wantPaid != !res.PaidAt.IsZero() {
				t.Errorf("paid at: %v", res.PaidAt)
			}
		})
	}

	t.Run("unknown_state_refuses", func(t *testing.T) {
		f, g := newTestGateway(t)
		f.route(http.MethodGet, "/v3/pay/transactions/out-trade-no/ord-1", http.StatusOK, `{"trade_state":"FROM_THE_FUTURE","amount":{"total":100}}`)
		if _, err := g.QueryPayment(context.Background(), "ord-1"); !errors.Is(err, payment.ErrUnavailable) {
			t.Fatalf("want ErrUnavailable for unknown trade_state, got %v", err)
		}
	})

	t.Run("bad_success_time_refuses", func(t *testing.T) {
		f, g := newTestGateway(t)
		f.route(http.MethodGet, "/v3/pay/transactions/out-trade-no/ord-1", http.StatusOK,
			`{"trade_state":"SUCCESS","success_time":"not-a-time","amount":{"total":100}}`)
		if _, err := g.QueryPayment(context.Background(), "ord-1"); !errors.Is(err, payment.ErrUnavailable) {
			t.Fatalf("want ErrUnavailable for bad success_time, got %v", err)
		}
	})

	t.Run("order_absent", func(t *testing.T) {
		f, g := newTestGateway(t)
		f.route(http.MethodGet, "/v3/pay/transactions/out-trade-no/ord-none", http.StatusNotFound,
			`{"code":"ORDER_NOT_EXISTS","message":"no such order"}`)
		if _, err := g.QueryPayment(context.Background(), "ord-none"); !errors.Is(err, payment.ErrNotFound) {
			t.Fatalf("want ErrNotFound, got %v", err)
		}
	})
}

func TestGatewayClosePayment(t *testing.T) {
	t.Run("closed_ok", func(t *testing.T) {
		f, g := newTestGateway(t)
		f.route(http.MethodPost, "/v3/pay/transactions/out-trade-no/ord-1/close", http.StatusNoContent, "")
		if err := g.ClosePayment(context.Background(), "ord-1"); err != nil {
			t.Fatalf("close: %v", err)
		}
	})

	t.Run("already_paid_is_channel_rejection", func(t *testing.T) {
		f, g := newTestGateway(t)
		f.route(http.MethodPost, "/v3/pay/transactions/out-trade-no/ord-1/close", http.StatusForbidden,
			`{"code":"ORDERPAID","message":"order paid"}`)
		err := g.ClosePayment(context.Background(), "ord-1")
		if !errors.Is(err, payment.ErrChannelRejected) {
			t.Fatalf("want ErrChannelRejected, got %v", err)
		}
		var ce *payment.ChannelError
		if !errors.As(err, &ce) || ce.Code != "ORDERPAID" {
			t.Fatalf("channel code: %v", err)
		}
	})
}

func TestGatewayRefund(t *testing.T) {
	cases := []struct {
		name      string
		body      string
		wantState payment.RefundState
	}{
		{"processing", `{"refund_id":"rf1","status":"PROCESSING","amount":{"refund":40}}`, payment.RefundStateProcessing},
		{"succeeded", `{"refund_id":"rf1","status":"SUCCESS","amount":{"refund":40}}`, payment.RefundStateSucceeded},
		{"abnormal", `{"refund_id":"rf1","status":"ABNORMAL","amount":{"refund":40}}`, payment.RefundStateAbnormal},
		{"closed_is_abnormal", `{"refund_id":"rf1","status":"CLOSED","amount":{"refund":40}}`, payment.RefundStateAbnormal},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f, g := newTestGateway(t)
			f.route(http.MethodPost, "/v3/refund/domestic/refunds", http.StatusOK, c.body)
			res, err := g.Refund(context.Background(), &payment.RefundInput{
				OutRefundNo: "ref-1", OutTradeNo: "ord-1",
				AmountFen: payment.Fen(40), TotalAmountFen: payment.Fen(100),
				Reason: "damage", NotifyURL: "https://xgwl-edu.com/refund/notify",
			})
			if err != nil {
				t.Fatalf("refund: %v", err)
			}
			if res.State != c.wantState || res.RefundID != "rf1" {
				t.Errorf("result: %+v", res)
			}
			m := jsonMap(t, f.lastRequest(t, http.MethodPost, "/v3/refund/domestic/refunds").Body)
			wantString(t, m, "out_trade_no", "ord-1")
			wantString(t, m, "out_refund_no", "ref-1")
			wantString(t, m, "reason", "damage")
			amount, _ := m["amount"].(map[string]any)
			if amount == nil || amount["refund"] != float64(40) || amount["total"] != float64(100) {
				t.Errorf("amount: %v", m["amount"])
			}
		})
	}

	t.Run("refund_over_declared_total_is_local", func(t *testing.T) {
		f, g := newTestGateway(t)
		if _, err := g.Refund(context.Background(), &payment.RefundInput{
			OutRefundNo: "ref-x", OutTradeNo: "ord-1",
			AmountFen: payment.Fen(101), TotalAmountFen: payment.Fen(100),
		}); !errors.Is(err, payment.ErrInvalidArgument) {
			t.Fatalf("want ErrInvalidArgument, got %v", err)
		}
		if len(f.requests) != 0 {
			t.Error("refusal must be local")
		}
	})

	t.Run("unknown_status_refuses", func(t *testing.T) {
		f, g := newTestGateway(t)
		f.route(http.MethodPost, "/v3/refund/domestic/refunds", http.StatusOK,
			`{"refund_id":"rf1","status":"FROM_THE_FUTURE","amount":{"refund":40}}`)
		if _, err := g.Refund(context.Background(), &payment.RefundInput{
			OutRefundNo: "ref-1", OutTradeNo: "ord-1",
			AmountFen: payment.Fen(40), TotalAmountFen: payment.Fen(100),
		}); !errors.Is(err, payment.ErrUnavailable) {
			t.Fatalf("want ErrUnavailable for unknown refund status, got %v", err)
		}
	})
}

func TestGatewayQueryRefund(t *testing.T) {
	f, g := newTestGateway(t)
	f.route(http.MethodGet, "/v3/refund/domestic/refunds/ref-1", http.StatusOK,
		`{"refund_id":"rf1","status":"SUCCESS","success_time":"2026-10-25T12:05:00+08:00","amount":{"refund":40}}`)
	res, err := g.QueryRefund(context.Background(), "ref-1")
	if err != nil {
		t.Fatalf("query refund: %v", err)
	}
	if res.State != payment.RefundStateSucceeded || res.AmountFen != payment.Fen(40) || res.RefundID != "rf1" {
		t.Errorf("result: %+v", res)
	}
	if res.SuccessAt.IsZero() {
		t.Error("success time must parse")
	}

	f.route(http.MethodGet, "/v3/refund/domestic/refunds/ref-none", http.StatusNotFound,
		`{"code":"RESOURCE_NOT_EXISTS","message":"no refund"}`)
	if _, err := g.QueryRefund(context.Background(), "ref-none"); !errors.Is(err, payment.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}

	// The truth read refuses an unparseable success time rather than
	// settling on a zero it minted itself.
	f.route(http.MethodGet, "/v3/refund/domestic/refunds/ref-bad-time", http.StatusOK,
		`{"refund_id":"rf2","status":"SUCCESS","success_time":"not-a-time","amount":{"refund":40}}`)
	if _, err := g.QueryRefund(context.Background(), "ref-bad-time"); !errors.Is(err, payment.ErrUnavailable) {
		t.Fatalf("want ErrUnavailable for bad success_time, got %v", err)
	}

	f.route(http.MethodGet, "/v3/refund/domestic/refunds/ref-future", http.StatusOK,
		`{"refund_id":"rf3","status":"FROM_THE_FUTURE","amount":{"refund":40}}`)
	if _, err := g.QueryRefund(context.Background(), "ref-future"); !errors.Is(err, payment.ErrUnavailable) {
		t.Fatalf("want ErrUnavailable for unknown refund status, got %v", err)
	}
}
