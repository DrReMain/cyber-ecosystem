package payment

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// TestIsTradeNo is the exhaustive shape check for merchant-side trade/refund
// numbers: ASCII alphanumerics, dash and underscore only, 1-32 bytes (32 is
// every known channel's tightest cap — WeChat's own limit).
func TestIsTradeNo(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "empty", input: "", want: false},
		{name: "single", input: "a", want: true},
		{name: "alphanumeric_mixed", input: "ORD20260101aB9", want: true},
		{name: "dash", input: "o-1", want: true},
		{name: "underscore", input: "o_1", want: true},
		{name: "at_cap_32", input: strings.Repeat("a", 32), want: true},
		{name: "over_cap_33", input: strings.Repeat("a", 33), want: false},
		{name: "dot", input: "a.b", want: false},
		{name: "slash", input: "a/b", want: false},
		{name: "space", input: "a b", want: false},
		{name: "cjk", input: "单号1", want: false},
		{name: "plus", input: "a+b", want: false},
		{name: "nul_control", input: "a\x00b", want: false},
	}
	for _, c := range cases {
		if got := isTradeNo(c.input); got != c.want {
			t.Errorf("isTradeNo(%q): want %v, got %v", c.input, c.want, got)
		}
	}
}

func validInput() *CreatePaymentInput {
	return &CreatePaymentInput{
		Channel:     Channel("any_provider_mode"),
		OutTradeNo:  "ord-1",
		AmountFen:   Fen(100),
		Description: "membership",
		NotifyURL:   "https://xgwl-edu.com/pay/notify",
		ExpireAt:    time.Time{},
		Payer:       &Payer{OpenID: "oX-123"},
	}
}

// TestValidateCreatePaymentInput walks the channel-neutral rules only:
// channel existence and Payer hints are provider knowledge, so an
// arbitrary channel value passes here untouched.
func TestValidateCreatePaymentInput(t *testing.T) {
	if err := ValidateCreatePaymentInput(validInput()); err != nil {
		t.Fatalf("baseline: want nil, got %v", err)
	}

	future := time.Now().Add(time.Hour)
	past := time.Now().Add(-time.Hour)
	cases := []struct {
		name string
		mut  func(*CreatePaymentInput)
	}{
		{name: "nil_input", mut: func(*CreatePaymentInput) {}},
		{name: "empty_trade_no", mut: func(in *CreatePaymentInput) { in.OutTradeNo = "" }},
		{name: "bad_trade_no_charset", mut: func(in *CreatePaymentInput) { in.OutTradeNo = "单-1" }},
		{name: "overlong_trade_no", mut: func(in *CreatePaymentInput) { in.OutTradeNo = strings.Repeat("a", 33) }},
		{name: "zero_amount", mut: func(in *CreatePaymentInput) { in.AmountFen = 0 }},
		{name: "negative_amount", mut: func(in *CreatePaymentInput) { in.AmountFen = Fen(-1) }},
		{name: "empty_notify_url", mut: func(in *CreatePaymentInput) { in.NotifyURL = "" }},
		{name: "expired_deadline", mut: func(in *CreatePaymentInput) { in.ExpireAt = past }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var in *CreatePaymentInput
			if c.name != "nil_input" {
				in = validInput()
			}
			c.mut(in)
			if err := ValidateCreatePaymentInput(in); !errors.Is(err, ErrInvalidArgument) {
				t.Errorf("%s: want ErrInvalidArgument, got %v", c.name, err)
			}
		})
	}

	// A future deadline passes; the zero deadline (channel default) already
	// passed in the baseline above.
	scheduled := validInput()
	scheduled.ExpireAt = future
	if err := ValidateCreatePaymentInput(scheduled); err != nil {
		t.Errorf("future ExpireAt: want nil, got %v", err)
	}
}

func TestValidateRefundInput(t *testing.T) {
	base := &RefundInput{
		OutRefundNo:    "ref-1",
		OutTradeNo:     "ord-1",
		AmountFen:      Fen(100),
		TotalAmountFen: Fen(100),
	}
	if err := ValidateRefundInput(base); err != nil {
		t.Fatalf("baseline: want nil, got %v", err)
	}
	cases := []struct {
		name string
		mut  func(*RefundInput)
	}{
		{name: "nil", mut: func(*RefundInput) {}},
		{name: "bad_refund_no", mut: func(in *RefundInput) { in.OutRefundNo = "r.1" }},
		{name: "bad_trade_no", mut: func(in *RefundInput) { in.OutTradeNo = "!" }},
		{name: "zero_amount", mut: func(in *RefundInput) { in.AmountFen = 0 }},
		{name: "negative_amount", mut: func(in *RefundInput) { in.AmountFen = Fen(-100) }},
		{name: "zero_total", mut: func(in *RefundInput) { in.TotalAmountFen = 0 }},
		{name: "refund_exceeds_total", mut: func(in *RefundInput) { in.TotalAmountFen = Fen(99) }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var in *RefundInput
			if c.name != "nil" {
				in = &RefundInput{OutRefundNo: "ref-1", OutTradeNo: "ord-1", AmountFen: Fen(100), TotalAmountFen: Fen(100)}
			}
			c.mut(in)
			if err := ValidateRefundInput(in); !errors.Is(err, ErrInvalidArgument) {
				t.Errorf("%s: want ErrInvalidArgument, got %v", c.name, err)
			}
		})
	}
	// A partial refund inside the declared total passes; NotifyURL is
	// optional for refunds — settlement converges via QueryRefund.
	base.Reason = "manual refund"
	base.NotifyURL = ""
	base.AmountFen = Fen(40)
	if err := ValidateRefundInput(base); err != nil {
		t.Errorf("partial refund: want nil, got %v", err)
	}
}

func TestValidateNotificationInput(t *testing.T) {
	if err := ValidateNotificationInput(nil); !errors.Is(err, ErrInvalidArgument) {
		t.Errorf("nil input: want ErrInvalidArgument, got %v", err)
	}
	if err := ValidateNotificationInput(&NotificationInput{Headers: nil, Body: nil}); !errors.Is(err, ErrInvalidArgument) {
		t.Errorf("nil body: want ErrInvalidArgument, got %v", err)
	}
	if err := ValidateNotificationInput(&NotificationInput{Body: []byte("{}")}); err != nil {
		t.Errorf("present body: want nil, got %v", err)
	}
}

func TestValidateBillInput(t *testing.T) {
	if err := ValidateBillInput(nil); !errors.Is(err, ErrInvalidArgument) {
		t.Errorf("nil input: want ErrInvalidArgument, got %v", err)
	}
	if err := ValidateBillInput(&BillInput{Kind: BillKindTrade}); !errors.Is(err, ErrInvalidArgument) {
		t.Errorf("zero date: want ErrInvalidArgument, got %v", err)
	}
	if err := ValidateBillInput(&BillInput{Date: time.Now(), Kind: BillKind("random")}); !errors.Is(err, ErrInvalidArgument) {
		t.Errorf("unknown kind: want ErrInvalidArgument, got %v", err)
	}
	for _, kind := range []BillKind{BillKindTrade, BillKindFundFlow} {
		if err := ValidateBillInput(&BillInput{Date: time.Now(), Kind: kind}); err != nil {
			t.Errorf("kind %q baseline: want nil, got %v", kind, err)
		}
	}
}
