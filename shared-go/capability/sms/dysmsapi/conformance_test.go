package dysmsapi

// Wire-protocol probes against an httptest edge standing in for the Aliyun
// RPC host: the full round trip with the server verifying the signature by
// the protocol's own rules, the common parameters the host requires, the
// channel-answer classification ladder (OK, business rejection, unparseable
// non-200), the timeout verdict, construction validation, and the family
// shape check running before any dial. A live channel round trip is
// documented at the bottom; it costs a real message, so nothing here dials
// out.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"sync/atomic"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/capability/sms"
)

const (
	testKeyID     = "LTAI-test-key-id"
	testKeySecret = "test-secret-value"
)

func testSender(t *testing.T, endpoint string, client *http.Client) *Sender {
	t.Helper()
	s, err := New(Config{
		AccessKeyID: testKeyID, AccessKeySecret: testKeySecret,
		Endpoint: endpoint, HTTPClient: client,
	})
	if err != nil {
		t.Fatalf("new sender: %v", err)
	}
	return s
}

func codeMessage() *sms.Message {
	return &sms.Message{
		To: "13800138000", SignName: "曦光未来", TemplateCode: "SMS_460000000",
		Params: map[string]string{"code": "654321"},
	}
}

// TestFullRoundTripVerifiesSignature: the fake host recomputes the
// signature from the received form exactly as the protocol prescribes —
// every parameter except the signature itself, sorted and encoded, HMAC'd
// under the known secret. A wrong key, nonce, or canonicalization breaks
// the assertion, so the client's whole signature path is proven, not
// self-attested.
func TestFullRoundTripVerifiesSignature(t *testing.T) {
	var gotForm map[string]string
	var body string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		gotForm = map[string]string{}
		for k, v := range r.PostForm {
			gotForm[k] = v[0]
		}
		body = `{"RequestId":"req-1","Code":"OK","Message":"OK","BizId":"biz-123"}`
		_, _ = fmt.Fprint(w, body)
	}))
	defer srv.Close()

	res, err := testSender(t, srv.URL, nil).Send(context.Background(), codeMessage())
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if res.BizID != "biz-123" {
		t.Fatalf("biz id = %q, want biz-123", res.BizID)
	}

	sig := gotForm["Signature"]
	if sig == "" {
		t.Fatal("no signature parameter on the wire")
	}
	signed := map[string]string{}
	for k, v := range gotForm {
		if k != "Signature" {
			signed[k] = v
		}
	}
	if want := signStringToSign(testKeySecret, stringToSign(canonicalQuery(signed))); sig != want {
		t.Fatalf("signature mismatch: client sent %q, protocol recomputes %q", sig, want)
	}
	for k, want := range map[string]string{
		"AccessKeyId":      testKeyID,
		"Action":           "SendSms",
		"Version":          "2017-05-25",
		"Format":           "JSON",
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureVersion": "1.0",
		"RegionId":         "cn-hangzhou",
		"PhoneNumbers":     "13800138000",
		"SignName":         "曦光未来",
		"TemplateCode":     "SMS_460000000",
	} {
		if gotForm[k] != want {
			t.Errorf("param %s = %q, want %q", k, gotForm[k], want)
		}
	}
	if gotForm["SignatureNonce"] == "" {
		t.Error("missing SignatureNonce — replays would pass")
	}
	// ISO 8601 UTC is the protocol's own clock format — a local-time or
	// offset-carrying timestamp is a rejected request.
	if !regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`).MatchString(gotForm["Timestamp"]) {
		t.Errorf("Timestamp = %q, want ISO 8601 UTC", gotForm["Timestamp"])
	}
	var tpl map[string]string
	if err := json.Unmarshal([]byte(gotForm["TemplateParam"]), &tpl); err != nil || tpl["code"] != "654321" {
		t.Errorf("TemplateParam = %q, want JSON {\"code\":\"654321\"}", gotForm["TemplateParam"])
	}
}

// TestNonceUniquePerRequest: the anti-replay nonce must differ across sends.
func TestNonceUniquePerRequest(t *testing.T) {
	var seen = map[string]bool{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		seen[r.PostForm.Get("SignatureNonce")] = true
		_, _ = fmt.Fprint(w, `{"Code":"OK","BizId":"b"}`)
	}))
	defer srv.Close()
	s := testSender(t, srv.URL, nil)
	for range 3 {
		if _, err := s.Send(context.Background(), codeMessage()); err != nil {
			t.Fatalf("send: %v", err)
		}
	}
	if len(seen) != 3 {
		t.Fatalf("nonce reuse: %d unique nonces across 3 sends", len(seen))
	}
}

// TestChannelAnswerClassification: the channel's own Code is the verdict —
// a business rejection arrives typed with its vocabulary intact, a garbage
// non-200 without a Code is unavailable, and an OK without a receipt id is
// a broken answer, not a success.
func TestChannelAnswerClassification(t *testing.T) {
	cases := []struct {
		name         string
		status       int
		body         string
		wantSentinel error
		wantCode     string
	}{
		{"ok", 200, `{"Code":"OK","BizId":"b1"}`, nil, ""},
		{"business_rejection", 200, `{"Code":"isv.BUSINESS_LIMIT_CONTROL","Message":"flow control"}`, sms.ErrChannelRejected, "isv.BUSINESS_LIMIT_CONTROL"},
		{"auth_error_on_http_400", 400, `{"Code":"SignatureDoesNotMatch","Message":"bad sign"}`, sms.ErrChannelRejected, "SignatureDoesNotMatch"},
		{"garbage_500", 500, `upstream exploded`, sms.ErrUnavailable, ""},
		{"ok_without_bizid", 200, `{"Code":"OK"}`, sms.ErrUnavailable, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = fmt.Fprint(w, tc.body)
			}))
			defer srv.Close()
			_, err := testSender(t, srv.URL, nil).Send(context.Background(), codeMessage())
			if tc.wantSentinel == nil {
				if err != nil {
					t.Fatalf("want success, got %v", err)
				}
				return
			}
			if !errors.Is(err, tc.wantSentinel) {
				t.Fatalf("want %v, got %v", tc.wantSentinel, err)
			}
			if tc.wantCode != "" {
				var ce *sms.ChannelError
				if !errors.As(err, &ce) || ce.Code != tc.wantCode {
					t.Fatalf("channel vocabulary lost: %v", err)
				}
			}
		})
	}
}

// TestTimeoutClassified: a deadline past the client timeout answers
// ErrTimeout — the message's fate is unknown and the caller must see that
// class distinctly from a plain failure to reach the channel.
func TestTimeoutClassified(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		time.Sleep(200 * time.Millisecond)
		_, _ = fmt.Fprint(w, `{"Code":"OK","BizId":"b"}`)
	}))
	defer srv.Close()
	s := testSender(t, srv.URL, &http.Client{Timeout: 50 * time.Millisecond})
	if _, err := s.Send(context.Background(), codeMessage()); !errors.Is(err, sms.ErrTimeout) {
		t.Fatalf("want ErrTimeout, got %v", err)
	}
	if hits.Load() == 0 {
		t.Fatal("the host never saw the request — timeout classified without a dial")
	}
}

// TestContextCancelClassified: caller cancellation is a deadline the same
// way the client timeout is.
func TestContextCancelClassified(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(150 * time.Millisecond)
		_, _ = fmt.Fprint(w, `{"Code":"OK","BizId":"b"}`)
	}))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	if _, err := testSender(t, srv.URL, nil).Send(ctx, codeMessage()); !errors.Is(err, sms.ErrTimeout) {
		t.Fatalf("want ErrTimeout, got %v", err)
	}
}

// TestEndpointInvalidClassified: an endpoint that cannot even form a
// request is the channel being unreachable — Unavailable, not a panic.
func TestEndpointInvalidClassified(t *testing.T) {
	s, err := New(Config{
		AccessKeyID: testKeyID, AccessKeySecret: testKeySecret,
		Endpoint: "://not-a-url",
	})
	if err != nil {
		t.Fatalf("construction should not parse the endpoint: %v", err)
	}
	if _, err = s.Send(context.Background(), codeMessage()); !errors.Is(err, sms.ErrUnavailable) {
		t.Fatalf("want ErrUnavailable, got %v", err)
	}
}

// TestShortBodyClassified: a host that promises more body than it sends is
// a broken answer, not a channel business verdict — the read failure
// classifies as Unavailable.
func TestShortBodyClassified(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "1000")
		_, _ = fmt.Fprint(w, `{"Code":"OK"`)
	}))
	defer srv.Close()
	if _, err := testSender(t, srv.URL, nil).Send(context.Background(), codeMessage()); !errors.Is(err, sms.ErrUnavailable) {
		t.Fatalf("want ErrUnavailable, got %v", err)
	}
}

// TestNetErrorTimeoutClassified: a transport surfacing a net.Error with
// Timeout() true is the deadline verdict — the message's fate is unknown,
// a distinct class from a plain failure to reach the channel.
func TestNetErrorTimeoutClassified(t *testing.T) {
	rt := roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, timeoutNetErr{}
	})
	if _, err := testSender(t, "http://unused.invalid", &http.Client{Transport: rt}).Send(context.Background(), codeMessage()); !errors.Is(err, sms.ErrTimeout) {
		t.Fatalf("want ErrTimeout, got %v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type timeoutNetErr struct{}

func (timeoutNetErr) Error() string   { return "net timeout" }
func (timeoutNetErr) Timeout() bool   { return true }
func (timeoutNetErr) Temporary() bool { return false }

// TestNewRequiresCredentials: a half-configured channel fails construction
// — the boot, not the first send.
func TestNewRequiresCredentials(t *testing.T) {
	if _, err := New(Config{}); err == nil {
		t.Fatal("empty config must fail construction")
	}
	if _, err := New(Config{AccessKeyID: "id"}); err == nil {
		t.Fatal("secretless config must fail construction")
	}
}

// TestShapeCheckRunsBeforeDial: an invalid message never reaches the
// channel — the host counter proves zero dials.
func TestShapeCheckRunsBeforeDial(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		_, _ = fmt.Fprint(w, `{"Code":"OK","BizId":"b"}`)
	}))
	defer srv.Close()
	bad := codeMessage()
	bad.To = "123"
	if _, err := testSender(t, srv.URL, nil).Send(context.Background(), bad); !errors.Is(err, sms.ErrInvalidArgument) {
		t.Fatalf("want ErrInvalidArgument, got %v", err)
	}
	if hits.Load() != 0 {
		t.Fatalf("invalid message dialed the channel %d times", hits.Load())
	}
}

// Live channel round trip — one real message, real money, so it never runs
// from the suite. It is the honest gate before production: the merchant's
// own key signing, the channel accepting it, the phone receiving the code.
//
//	ALIYUN_SMS_CONFORMANCE_ACCESS_KEY_ID=... \
//	ALIYUN_SMS_CONFORMANCE_ACCESS_KEY_SECRET=... \
//	ALIYUN_SMS_CONFORMANCE_SIGN=... \
//	ALIYUN_SMS_CONFORMANCE_TEMPLATE=... \
//	ALIYUN_SMS_CONFORMANCE_PHONE=13... \
//	go test ./shared-go/capability/sms/dysmsapi/ -run TestLiveConformance -count=1 -v
//
// Values come from the operator console — never from the repo.
func TestLiveConformance(t *testing.T) {
	cfg := Config{
		AccessKeyID:     os.Getenv("ALIYUN_SMS_CONFORMANCE_ACCESS_KEY_ID"),
		AccessKeySecret: os.Getenv("ALIYUN_SMS_CONFORMANCE_ACCESS_KEY_SECRET"),
	}
	if cfg.AccessKeyID == "" {
		t.Skip("set ALIYUN_SMS_CONFORMANCE_* to run the live channel round trip")
	}
	s, err := New(cfg)
	if err != nil {
		t.Fatalf("new sender: %v", err)
	}
	msg := &sms.Message{
		To:           os.Getenv("ALIYUN_SMS_CONFORMANCE_PHONE"),
		SignName:     os.Getenv("ALIYUN_SMS_CONFORMANCE_SIGN"),
		TemplateCode: os.Getenv("ALIYUN_SMS_CONFORMANCE_TEMPLATE"),
		Params:       map[string]string{"code": "654321"},
	}
	res, err := s.Send(context.Background(), msg)
	if err != nil {
		t.Fatalf("live send rejected: %v", err)
	}
	t.Logf("live receipt: %s", res.BizID)
}
