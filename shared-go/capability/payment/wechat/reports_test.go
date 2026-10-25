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
