package auth

// Unit probes over the sms provider's two seams: IssueCode (throttle order,
// canonical phone, code storage, channel carry, state minted only on a
// delivered send) and Exchange (burn-on-read, uniform invalid answer,
// canonical identity). The channel, limiter, and counter are fakes; the
// code store rides the same fakeTokenRP the session tests use — fakes
// simulate the classification layer, never raw errors.

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/capability/cache"
	"cyber-ecosystem/shared-go/capability/sms"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/conf"
	"cyber-ecosystem/app/services/system/internal/platform"
)

type fakeSmsSender struct {
	sent []*sms.Message
	err  error
}

func (f *fakeSmsSender) Send(_ context.Context, msg *sms.Message) (*sms.SendResult, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.sent = append(f.sent, msg)
	return &sms.SendResult{BizID: "fake-biz"}, nil
}

type fakeLimiter struct {
	allowed bool
	calls   int
}

func (f *fakeLimiter) Allow(context.Context, string, int64, time.Duration) (*cache.RateResult, error) {
	f.calls++
	return &cache.RateResult{Allowed: f.allowed}, nil
}

type fakeCounter struct {
	next int
	seq  []int64 // successive Incr answers when scripted
	i    int
	exp  []string
}

func (f *fakeCounter) Incr(context.Context, string, int64) (int64, error) {
	if f.seq != nil {
		v := f.seq[f.i]
		f.i++
		return v, nil
	}
	return int64(f.next), nil
}
func (f *fakeCounter) Decr(context.Context, string, int64) (int64, error) { return 0, nil }
func (f *fakeCounter) Get(context.Context, string) (int64, error)         { return 0, nil }
func (f *fakeCounter) Set(context.Context, string, int64) error           { return nil }
func (f *fakeCounter) Expire(_ context.Context, key string, _ time.Duration) error {
	f.exp = append(f.exp, key)
	return nil
}

func testHandleSms() platform.SmsErrorHandler {
	return func(err error) error { return err }
}

func smsTestCache(lim cache.RateLimiter, ctr cache.Counter) *cache.Cache {
	return &cache.Cache{RateLimiter: lim, Counter: ctr}
}

func newSmsProviderForTest(sender *fakeSmsSender, lim *fakeLimiter, ctr *fakeCounter, trp TokenRP) *smsProvider {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewSmsProvider(logger, sender, testHandleSms(), func(err error) error { return err }, smsTestCache(lim, ctr), trp)
}

func smsRow() *conf.Auth_Provider {
	return &conf.Auth_Provider{
		Key: "sms", Kind: smsKind, Enabled: true,
		Sms: &conf.Auth_Provider_Sms{SignName: "曦光未来", TemplateCode: "SMS_TEST"},
	}
}

func smsReq(phone, code string) *systempb.LoginByIdentityRequest {
	p, c := phone, code
	return &systempb.LoginByIdentityRequest{
		Payload: &systempb.LoginByIdentityRequest_Sms{Sms: &systempb.SmsIdentityPayload{Phone: &p, Code: &c}},
	}
}

// TestIssueCodeHappyPath: one send passes the window, stores the code under
// the canonical phone with the code TTL, carries sign/template/params to
// the channel, and returns a state bound to the provider key.
func TestIssueCodeHappyPath(t *testing.T) {
	sender, lim, ctr, trp := &fakeSmsSender{}, &fakeLimiter{allowed: true}, &fakeCounter{next: 1}, newTokenStore()
	p := newSmsProviderForTest(sender, lim, ctr, trp)

	state, err := p.IssueCode(context.Background(), smsRow(), "+8613800138000")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if lim.calls != 1 {
		t.Fatalf("limiter calls = %d, want 1", lim.calls)
	}
	if len(ctr.exp) != 1 {
		t.Fatalf("daily counter expiry not set on first increment: %v", ctr.exp)
	}
	code, ok := trp.vals[smsCodePrefix+"13800138000"]
	if !ok {
		t.Fatal("code not stored under the canonical phone")
	}
	if len(code) != 6 {
		t.Fatalf("code %q not 6 digits", code)
	}
	if trp.ttls[smsCodePrefix+"13800138000"] != smsCodeTTL {
		t.Fatalf("code TTL = %v, want %v", trp.ttls[smsCodePrefix+"13800138000"], smsCodeTTL)
	}
	if len(sender.sent) != 1 {
		t.Fatalf("channel saw %d sends, want 1", len(sender.sent))
	}
	m := sender.sent[0]
	if m.To != "13800138000" || m.SignName != "曦光未来" || m.TemplateCode != "SMS_TEST" || m.Params["code"] != string(code) {
		t.Fatalf("channel message = %+v", m)
	}
	if string(trp.vals[identityStatePrefix+state]) != "sms" {
		t.Fatal("state not bound to the provider key")
	}
}

// TestIssueCodeThrottleOrder: a non-mainland phone rejects before the
// limiter ever runs, and a closed window rejects before any code is stored
// or sent.
func TestIssueCodeThrottleOrder(t *testing.T) {
	sender, lim, ctr, trp := &fakeSmsSender{}, &fakeLimiter{allowed: true}, &fakeCounter{next: 1}, newTokenStore()
	p := newSmsProviderForTest(sender, lim, ctr, trp)

	if _, err := p.IssueCode(context.Background(), smsRow(), "12345"); !errorspb.IsGeneralErrorValidationFailed(err) {
		t.Fatalf("non-mainland: want validation failed, got %v", err)
	}
	if lim.calls != 0 {
		t.Fatal("invalid phone consumed the send window")
	}

	lim.allowed = false
	if _, err := p.IssueCode(context.Background(), smsRow(), "13800138000"); !systempb.IsSystemSmsRateLimited(err) {
		t.Fatalf("closed window: want rate limited, got %v", err)
	}
	if len(sender.sent) != 0 || len(trp.vals) != 0 {
		t.Fatal("a throttled send stored a code or dialed the channel")
	}

	lim.allowed = true
	ctr.next = smsSendPerDay + 1
	if _, err := p.IssueCode(context.Background(), smsRow(), "13800138000"); !systempb.IsSystemSmsRateLimited(err) {
		t.Fatalf("daily cap: want rate limited, got %v", err)
	}
	if len(sender.sent) != 0 {
		t.Fatal("a capped send dialed the channel")
	}
}

// TestIssueCodeNoStateOnFailedSend: a channel rejection answers the mapped
// infra error and mints no state — the caller never presents a state whose
// code never traveled.
func TestIssueCodeNoStateOnFailedSend(t *testing.T) {
	sender := &fakeSmsSender{err: &sms.ChannelError{Code: "isv.BUSINESS_LIMIT_CONTROL", Message: "flow control"}}
	lim, ctr, trp := &fakeLimiter{allowed: true}, &fakeCounter{next: 1}, newTokenStore()
	p := newSmsProviderForTest(sender, lim, ctr, trp)

	state, err := p.IssueCode(context.Background(), smsRow(), "13800138000")
	if state != "" || !errors.Is(err, sms.ErrChannelRejected) {
		t.Fatalf("want empty state + channel rejection, got %q / %v", state, err)
	}
	for k := range trp.vals {
		if len(k) > len(identityStatePrefix) && k[:len(identityStatePrefix)] == identityStatePrefix {
			t.Fatalf("failed send left a state behind: %s", k)
		}
	}
}

// TestExchangeBurnsOnce: a correct code exchanges to the canonical identity
// and consumes the stored code; a second presentation with the same digits
// answers invalid — the burn happened at the read.
func TestExchangeBurnsOnce(t *testing.T) {
	trp := newTokenStore()
	p := newSmsProviderForTest(&fakeSmsSender{}, &fakeLimiter{allowed: true}, &fakeCounter{next: 1}, trp)
	if err := trp.Set(context.Background(), smsCodePrefix+"13800138000", []byte("654321"), smsCodeTTL); err != nil {
		t.Fatal(err)
	}

	id, err := p.Exchange(context.Background(), smsRow(), smsReq("13800138000", "654321"))
	if err != nil {
		t.Fatalf("exchange: %v", err)
	}
	if id.Kind != smsKind || id.Identifier != "13800138000" {
		t.Fatalf("identity = %+v", id)
	}
	if _, err = p.Exchange(context.Background(), smsRow(), smsReq("13800138000", "654321")); !systempb.IsSystemSmsCodeInvalid(err) {
		t.Fatalf("replay: want code invalid, got %v", err)
	}
}

// TestExchangeUniformInvalidAnswer: a wrong code and a missing code answer
// the same error — the caller learns nothing about which.
func TestExchangeUniformInvalidAnswer(t *testing.T) {
	trp := newTokenStore()
	p := newSmsProviderForTest(&fakeSmsSender{}, &fakeLimiter{allowed: true}, &fakeCounter{next: 1}, trp)

	_, miss := p.Exchange(context.Background(), smsRow(), smsReq("13800138000", "111111"))
	if !systempb.IsSystemSmsCodeInvalid(miss) {
		t.Fatalf("no live code: want code invalid, got %v", miss)
	}

	if serr := trp.Set(context.Background(), smsCodePrefix+"13800138000", []byte("222222"), smsCodeTTL); serr != nil {
		t.Fatal(serr)
	}
	_, wrong := p.Exchange(context.Background(), smsRow(), smsReq("13800138000", "111111"))
	if !systempb.IsSystemSmsCodeInvalid(wrong) {
		t.Fatalf("wrong code: want code invalid, got %v", wrong)
	}
	// Uniform to the client means the reason class; the causes stay
	// distinct server-side for the audit trail.
	// The wrong guess consumed the stored code — one stored code, one
	// presentation, whatever the verdict.
	if _, replay := p.Exchange(context.Background(), smsRow(), smsReq("13800138000", "222222")); !systempb.IsSystemSmsCodeInvalid(replay) {
		t.Fatalf("code survived a wrong guess: %v", replay)
	}
}

// TestExchangeRejectsPayloadWithoutSmsView: a request whose oneof carries
// no sms member (BindCredential today) narrows to a validation failure,
// never a panic.
func TestExchangeRejectsPayloadWithoutSmsView(t *testing.T) {
	p := newSmsProviderForTest(&fakeSmsSender{}, &fakeLimiter{allowed: true}, &fakeCounter{next: 1}, newTokenStore())
	bindOnly := &systempb.BindCredentialRequest{}
	if _, err := p.Exchange(context.Background(), smsRow(), bindOnly); !errorspb.IsGeneralErrorValidationFailed(err) {
		t.Fatalf("want validation failed, got %v", err)
	}
	// A nil sms member inside a sms-capable request is the same verdict.
	if _, err := p.Exchange(context.Background(), smsRow(), &systempb.LoginByIdentityRequest{}); !errorspb.IsGeneralErrorValidationFailed(err) {
		t.Fatalf("nil sms member: want validation failed, got %v", err)
	}
}

// TestBeginIdentityLoginRefusesSmsKind: the direct-code provider has no
// redirect flow; Begin refuses it with the provider-unknown class rather
// than a type panic.
func TestBeginIdentityLoginRefusesSmsKind(t *testing.T) {
	trp := newTokenStore()
	p := newSmsProviderForTest(&fakeSmsSender{}, &fakeLimiter{allowed: true}, &fakeCounter{next: 1}, trp)
	providers := NewIdentityProviders(NewWechatProvider(), p)
	authConf := &conf.Auth{Providers: []*conf.Auth_Provider{{Key: "sms", Kind: smsKind, Enabled: true}}}
	uc, err := NewAuthUC(slog.New(slog.NewTextHandler(io.Discard, nil)), nil, authConf, providers, nil, nil, nil, nil, trp)
	if err != nil {
		t.Fatalf("uc: %v", err)
	}
	if _, err = uc.BeginIdentityLogin(context.Background(), "sms", "https://x.example/cb"); !systempb.IsSystemIdentityProviderUnknown(err) {
		t.Fatalf("want provider unknown, got %v", err)
	}
}

// TestSendSmsCodeRefusesNonIssuer: a redirect-only provider (wechat) issues
// no codes — the mirror of Begin refusing the sms kind.
func TestSendSmsCodeRefusesNonIssuer(t *testing.T) {
	p := newSmsProviderForTest(&fakeSmsSender{}, &fakeLimiter{allowed: true}, &fakeCounter{next: 1}, newTokenStore())
	providers := NewIdentityProviders(NewWechatProvider(), p)
	authConf := &conf.Auth{Providers: []*conf.Auth_Provider{
		{Key: "wechat", Kind: wechatKind, Enabled: true},
	}}
	uc, err := NewAuthUC(slog.New(slog.NewTextHandler(io.Discard, nil)), nil, authConf, providers, nil, nil, nil, nil, newTokenStore())
	if err != nil {
		t.Fatalf("uc: %v", err)
	}
	if _, err = uc.SendSmsCode(context.Background(), "wechat", "13800138000"); !systempb.IsSystemIdentityProviderUnknown(err) {
		t.Fatalf("want provider unknown, got %v", err)
	}
}

// TestIssueCodeDailyCapBoundary: the cap admits exactly smsSendPerDay sends
// a day — the 10th passes, the 11th throttles; an off-by-one here flips a
// whole phone's daily budget.
func TestIssueCodeDailyCapBoundary(t *testing.T) {
	sender := &fakeSmsSender{}
	lim := &fakeLimiter{allowed: true}
	seq := make([]int64, smsSendPerDay)
	for i := range seq {
		seq[i] = int64(i + 1)
	}
	ctr := &fakeCounter{seq: seq}
	p := newSmsProviderForTest(sender, lim, ctr, newTokenStore())

	for range smsSendPerDay {
		if _, err := p.IssueCode(context.Background(), smsRow(), "13800138000"); err != nil {
			t.Fatalf("send inside the daily budget: %v", err)
		}
	}
	if len(sender.sent) != smsSendPerDay {
		t.Fatalf("channel saw %d sends, want %d", len(sender.sent), smsSendPerDay)
	}
	ctr.seq, ctr.i = []int64{smsSendPerDay + 1}, 0
	if _, err := p.IssueCode(context.Background(), smsRow(), "13800138000"); !systempb.IsSystemSmsRateLimited(err) {
		t.Fatalf("send past the daily budget: want rate limited, got %v", err)
	}
	if len(sender.sent) != smsSendPerDay {
		t.Fatal("an over-budget send dialed the channel")
	}
}

// TestFailedSendInvalidatesPriorCode: the store holds one live code per
// phone, so a failed resend overwrites the prior code — the earlier pair is
// dead even though its state still lives, and a fresh successful send arms
// a new pair.
func TestFailedSendInvalidatesPriorCode(t *testing.T) {
	sender := &fakeSmsSender{}
	lim := &fakeLimiter{allowed: true}
	trp := newTokenStore()
	p := newSmsProviderForTest(sender, lim, &fakeCounter{next: 1}, trp)

	if _, err := p.IssueCode(context.Background(), smsRow(), "13800138000"); err != nil {
		t.Fatal(err)
	}
	first := string(trp.vals[smsCodePrefix+"13800138000"])

	sender.err = &sms.ChannelError{Code: "isv.BUSINESS_LIMIT_CONTROL", Message: "flow control"}
	if _, err := p.IssueCode(context.Background(), smsRow(), "13800138000"); !errors.Is(err, sms.ErrChannelRejected) {
		t.Fatalf("failed resend: %v", err)
	}
	if cur := string(trp.vals[smsCodePrefix+"13800138000"]); cur == first {
		t.Fatal("a failed send left the prior code armed")
	}

	// The overwritten code no longer exchanges — one live code per phone.
	if _, err := p.Exchange(context.Background(), smsRow(), smsReq("13800138000", first)); !systempb.IsSystemSmsCodeInvalid(err) {
		t.Fatalf("prior code still exchanges: %v", err)
	}

	sender.err = nil
	if _, err := p.IssueCode(context.Background(), smsRow(), "13800138000"); err != nil {
		t.Fatal(err)
	}
	fresh := string(trp.vals[smsCodePrefix+"13800138000"])
	if _, err := p.Exchange(context.Background(), smsRow(), smsReq("13800138000", fresh)); err != nil {
		t.Fatalf("fresh pair exchange: %v", err)
	}
}

// TestMaskPhoneTable: log and audit faces show three head digits, four
// tail, stars between; anything too short to split hides entirely.
func TestMaskPhoneTable(t *testing.T) {
	cases := []struct{ in, want string }{
		{"13800138000", "138****8000"},
		{"+8613800138000", "+86****8000"},
		{"123456", "******"},
		{"", ""},
	}
	for _, tc := range cases {
		if got := maskPhone(tc.in); got != tc.want {
			t.Errorf("maskPhone(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestNormalizePhoneTable: the +86 prefix strips to the bare mainland form;
// anything that is not a mainland mobile rejects.
func TestNormalizePhoneTable(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"13800138000", "13800138000"},
		{"+8613800138000", "13800138000"},
		{"+85212345678", ""},
		{"1380013800", ""},
		{"23800138000", ""},
		{"", ""},
	}
	for _, tc := range cases {
		if got := normalizePhone(tc.in); got != tc.want {
			t.Errorf("normalizePhone(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestRandomDigitsShape: six decimal digits every draw, no digit locked
// out (the rejection-sampled decade has no unreachable residue).
func TestRandomDigitsShape(t *testing.T) {
	seen := map[byte]bool{}
	for range 500 {
		s, err := randomDigits(6)
		if err != nil {
			t.Fatalf("draw: %v", err)
		}
		if len(s) != 6 {
			t.Fatalf("length = %d", len(s))
		}
		for i := 0; i < 6; i++ {
			if s[i] < '0' || s[i] > '9' {
				t.Fatalf("non-digit %q", s)
			}
			seen[s[i]] = true
		}
	}
	if len(seen) != 10 {
		t.Fatalf("only %d distinct digits drawn", len(seen))
	}
}
