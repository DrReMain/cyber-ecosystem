package wechat

// Statement probes: the two-hop download (signed ticket, unsigned bill
// host), the gzip-unwrapping, the hash gate, and the miss shapes — a day
// with no bill, a late JSON-wrapped error, a hash that does not match.

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/capability/payment"
)

func gzipBytes(t *testing.T, data []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(data); err != nil {
		t.Fatalf("gzip write: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("gzip close: %v", err)
	}
	return buf.Bytes()
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func routeBill(t *testing.T, f *fakeChannel, kindPath string, csv []byte) {
	t.Helper()
	zipped := gzipBytes(t, csv)
	f.route(http.MethodGet, kindPath, http.StatusOK,
		`{"download_url":"`+f.srv.URL+`/v3/download/bill","hash_type":"SHA256","hash_value":"`+sha256Hex(csv)+`"}`)
	f.routeCT(http.MethodGet, "/v3/download/bill", http.StatusOK, string(zipped), "application/x-gzip")
}

func newTestReports(t *testing.T) (*fakeChannel, *reports) {
	t.Helper()
	f := newFakeChannel(t)
	c, err := newClient(f.publicKeyConfig(t))
	if err != nil {
		t.Fatalf("newClient: %v", err)
	}
	return f, &reports{c: c}
}

func TestReportsDownloadBill(t *testing.T) {
	f, r := newTestReports(t)

	t.Run("trade", func(t *testing.T) {
		csv := []byte("2026-10-24,ord-1,tx-1,100.00\n")
		routeBill(t, f, "/v3/bill/tradebill", csv)
		day := time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC)
		res, err := r.DownloadBill(context.Background(), &payment.BillInput{Date: day, Kind: payment.BillKindTrade})
		if err != nil {
			t.Fatalf("download: %v", err)
		}
		if !bytes.Equal(res.Raw, csv) {
			t.Errorf("raw: %q", res.Raw)
		}
		captured := f.lastRequest(t, http.MethodGet, "/v3/bill/tradebill")
		if captured.Query != "account_type=BASIC&bill_date=2026-10-24" {
			t.Errorf("query: %q", captured.Query)
		}
	})

	t.Run("fund_flow", func(t *testing.T) {
		csv := []byte("2026-10-24,100.00,CNY\n")
		routeBill(t, f, "/v3/bill/fundflowbill", csv)
		res, err := r.DownloadBill(context.Background(), &payment.BillInput{
			Date: time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC), Kind: payment.BillKindFundFlow,
		})
		if err != nil {
			t.Fatalf("download: %v", err)
		}
		if !bytes.Equal(res.Raw, csv) {
			t.Errorf("raw: %q", res.Raw)
		}
	})
}

func TestReportsHashMismatch(t *testing.T) {
	f, r := newTestReports(t)
	csv := []byte("row\n")
	f.route(http.MethodGet, "/v3/bill/tradebill", http.StatusOK,
		`{"download_url":"`+f.srv.URL+`/v3/download/bill","hash_type":"SHA256","hash_value":"deadbeef"}`)
	f.routeCT(http.MethodGet, "/v3/download/bill", http.StatusOK, string(gzipBytes(t, csv)), "application/x-gzip")
	_, err := r.DownloadBill(context.Background(), &payment.BillInput{
		Date: time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC), Kind: payment.BillKindTrade,
	})
	if !errors.Is(err, payment.ErrInvalidSignature) {
		t.Fatalf("want ErrInvalidSignature, got %v", err)
	}
}

// A statement day with no bill is a query miss.
func TestReportsNoBill(t *testing.T) {
	f, r := newTestReports(t)
	f.route(http.MethodGet, "/v3/bill/tradebill", http.StatusBadRequest,
		`{"code":"NO_BILL_EXIST","message":"no bill for the day"}`)
	_, err := r.DownloadBill(context.Background(), &payment.BillInput{
		Date: time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC), Kind: payment.BillKindTrade,
	})
	if !errors.Is(err, payment.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

// The bill host wraps late failures (an expired ticket) in a 2xx JSON body.
func TestReportsLateError(t *testing.T) {
	f, r := newTestReports(t)
	f.route(http.MethodGet, "/v3/bill/tradebill", http.StatusOK,
		`{"download_url":"`+f.srv.URL+`/v3/download/bill","hash_type":"SHA256"}`)
	f.routeCT(http.MethodGet, "/v3/download/bill", http.StatusOK,
		`{"code":"SYSTEM_ERROR","message":"ticket expired"}`, "application/json")
	_, err := r.DownloadBill(context.Background(), &payment.BillInput{
		Date: time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC), Kind: payment.BillKindTrade,
	})
	if !errors.Is(err, payment.ErrUnavailable) {
		t.Fatalf("want ErrUnavailable, got %v", err)
	}
}

// A non-2xx from the bill host folds through the same classification as the
// API host.
func TestReportsBillHostRejection(t *testing.T) {
	f, r := newTestReports(t)
	f.route(http.MethodGet, "/v3/bill/tradebill", http.StatusOK,
		`{"download_url":"`+f.srv.URL+`/v3/download/bill","hash_type":"SHA256"}`)
	f.routeCT(http.MethodGet, "/v3/download/bill", http.StatusForbidden,
		`{"code":"NO_AUTH","message":"ticket rejected"}`, "application/json")
	_, err := r.DownloadBill(context.Background(), &payment.BillInput{
		Date: time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC), Kind: payment.BillKindTrade,
	})
	if !errors.Is(err, payment.ErrChannelRejected) {
		t.Fatalf("want ErrChannelRejected, got %v", err)
	}
}

// A gzip-tagged body that will not gunzip refuses rather than passing the
// bytes through as if they were a bill.
func TestReportsCorruptGzip(t *testing.T) {
	f, r := newTestReports(t)
	f.route(http.MethodGet, "/v3/bill/tradebill", http.StatusOK,
		`{"download_url":"`+f.srv.URL+`/v3/download/bill","hash_type":"SHA256"}`)
	f.routeCT(http.MethodGet, "/v3/download/bill", http.StatusOK, "not gzip at all", "application/x-gzip")
	_, err := r.DownloadBill(context.Background(), &payment.BillInput{
		Date: time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC), Kind: payment.BillKindTrade,
	})
	if !errors.Is(err, payment.ErrUnavailable) {
		t.Fatalf("want ErrUnavailable, got %v", err)
	}
}

// A ticket without a download URL is channel-state refuse, not a bill.
func TestReportsEmptyTicket(t *testing.T) {
	f, r := newTestReports(t)
	f.route(http.MethodGet, "/v3/bill/tradebill", http.StatusOK, `{}`)
	_, err := r.DownloadBill(context.Background(), &payment.BillInput{
		Date: time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC), Kind: payment.BillKindTrade,
	})
	if !errors.Is(err, payment.ErrUnavailable) {
		t.Fatalf("want ErrUnavailable, got %v", err)
	}
}

func TestNewAssembles(t *testing.T) {
	f := newFakeChannel(t)
	p, err := New(f.publicKeyConfig(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if p.Gateway == nil || p.Notifier == nil || p.Reports == nil {
		t.Fatalf("container faces must all be assembled: %+v", p)
	}
	bad := f.publicKeyConfig(t)
	bad.MchID = ""
	if _, err := New(bad); !errors.Is(err, payment.ErrInvalidArgument) {
		t.Fatalf("bad config: want ErrInvalidArgument, got %v", err)
	}
}

func TestParseTradeBill(t *testing.T) {
	_, r := newTestReports(t)

	header := "交易时间,公众账号ID,商户号,特约商户号,设备号,微信订单号,商户订单号,用户标识,交易类型,交易状态,付款银行,货币种类,应结订单金额,代金券金额,微信退款单号,商户退款单号,退款金额,充值券退款金额,费率,手续费金额,订单金额,费率备注,商户数据包,支付场景"
	paid := "2026-10-24 12:34:56,wx-app,mch-1,,,tx-1,ord-1,o-user,NATIVE,SUCCESS,CMB,CNY,19.83,0.00,,,,,0.60%,0.07,19.90,,,SHOP"
	refund := "2026-10-25 09:00:00,wx-app,mch-1,,,tx-1,ord-1,o-user,NATIVE,REFUND,CMB,CNY,0.00,0.00,rf-1,ord-r-1,19.90,0.00,,0.00,19.90,,,"

	t.Run("rows", func(t *testing.T) {
		raw := []byte("\ufeff" + header + "\n" + paid + "\n" + refund + "\n")
		lines, err := r.ParseTradeBill(context.Background(), raw)
		if err != nil {
			t.Fatalf("parse: %v", err)
		}
		if len(lines) != 2 {
			t.Fatalf("lines: want 2, got %d", len(lines))
		}
		l := lines[0]
		if l.OutTradeNo != "ord-1" || l.TransactionID != "tx-1" || l.State != payment.BillLinePaid {
			t.Errorf("paid row: %+v", l)
		}
		if int64(l.AmountFen) != 1990 || int64(l.FeeFen) != 7 || int64(l.IncomeFen) != 1983 || int64(l.RefundFen) != 0 {
			t.Errorf("paid amounts: %+v", l)
		}
		want := time.Date(2026, 10, 24, 12, 34, 56, 0, cstZone)
		if !l.At.Equal(want) {
			t.Errorf("paid time: got %v want %v", l.At, want)
		}
		if lines[1].State != payment.BillLineRefund || int64(lines[1].RefundFen) != 1990 {
			t.Errorf("refund row: %+v", lines[1])
		}
	})

	t.Run("column reorder", func(t *testing.T) {
		// Same facts, columns moved — lookup is by name, position never means anything.
		raw := []byte("商户订单号,交易状态,订单金额,交易时间,微信订单号,退款金额,手续费金额,应结订单金额\nord-9,SUCCESS,0.05,2026-10-24 08:00:00,tx-9,0.00,0.00,0.05\n")
		lines, err := r.ParseTradeBill(context.Background(), raw)
		if err != nil {
			t.Fatalf("parse: %v", err)
		}
		if len(lines) != 1 || int64(lines[0].AmountFen) != 5 || lines[0].TransactionID != "tx-9" {
			t.Fatalf("lines: %+v err %v", lines, err)
		}
	})

	t.Run("missing column", func(t *testing.T) {
		_, err := r.ParseTradeBill(context.Background(), []byte("交易时间,商户订单号\n2026-10-24 12:00:00,ord-1\n"))
		if !errors.Is(err, payment.ErrInvalidArgument) {
			t.Errorf("missing column: want ErrInvalidArgument, got %v", err)
		}
	})

	t.Run("bad amount", func(t *testing.T) {
		raw := []byte("商户订单号,交易状态,订单金额\nord-1,SUCCESS,19.9.0\n")
		if _, err := r.ParseTradeBill(context.Background(), raw); !errors.Is(err, payment.ErrInvalidArgument) {
			t.Errorf("bad amount: want ErrInvalidArgument, got %v", err)
		}
	})

	t.Run("unknown state", func(t *testing.T) {
		raw := []byte("商户订单号,交易状态,订单金额\nord-1,WEIRD,19.90\n")
		if _, err := r.ParseTradeBill(context.Background(), raw); !errors.Is(err, payment.ErrInvalidArgument) {
			t.Errorf("unknown state: want ErrInvalidArgument, got %v", err)
		}
	})
}

func TestParseYuanFen(t *testing.T) {
	cases := []struct {
		in   string
		want int64
	}{
		{"19.90", 1990},
		{"0.05", 5},
		{"0.5", 50},
		{"100", 10000},
		{"", 0},
		{"-2.00", -200},
	}
	for _, c := range cases {
		got, err := parseYuanFen(c.in)
		if err != nil {
			t.Fatalf("%q: %v", c.in, err)
		}
		if int64(got) != c.want {
			t.Errorf("%q: want %d got %d", c.in, c.want, int64(got))
		}
	}
	if _, err := parseYuanFen("abc"); !errors.Is(err, payment.ErrInvalidArgument) {
		t.Errorf("abc: want ErrInvalidArgument, got %v", err)
	}
	// Corrupted cells refuse rather than guess: double signs (whose inner
	// negative once folded to a positive), bare signs, partial decimals,
	// sub-fen precision, and an overflow that would wrap.
	for _, bad := range []string{"--1.00", "+1.00", "-", ".5", "5.", "19.904", "92233720368547758.08"} {
		if _, err := parseYuanFen(bad); !errors.Is(err, payment.ErrInvalidArgument) {
			t.Errorf("%q: want ErrInvalidArgument, got %v", bad, err)
		}
	}
}
