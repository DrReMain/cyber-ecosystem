package auth

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/kratos/security"
	"cyber-ecosystem/shared-go/utils"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/conf"
	"cyber-ecosystem/app/services/system/internal/module/tenant"
	"cyber-ecosystem/app/services/system/internal/module/user"
)

// fakeIdentityProvider answers the exchange seam with a fixed identity; an
// instance keyed "failing" rejects in the domain shape. Registered once
// under its own kind.
type fakeIdentityProvider struct {
	id *exchangedIdentity
}

// fakeExchanges counts calls into the exchange seam; ordering laws assert it
// stays put where a refused cash must never reach the provider.
var fakeExchanges int

func (f fakeIdentityProvider) Exchange(_ context.Context, row *conf.Auth_Provider, _ identityPayload) (*exchangedIdentity, error) {
	fakeExchanges++
	if row.GetKey() == "failing" {
		return nil, systempb.ErrorSystemIdentityExchangeFailed("").WithCause(fmt.Errorf("provider rejected"))
	}
	return f.id, nil
}

func (f fakeIdentityProvider) AuthorizeURL(row *conf.Auth_Provider, _, state string) string {
	return "https://idp.example/auth/" + row.GetKey() + "?state=" + state
}

const fakeKind = "fakekind"

func init() {
	registerIdentityProvider(fakeKind, fakeIdentityProvider{id: &exchangedIdentity{Kind: fakeKind, Identifier: "oid-x"}})
}

func fakeProviderConf() *conf.Auth {
	return &conf.Auth{Providers: []*conf.Auth_Provider{
		{Key: "fake", Kind: fakeKind, Enabled: true},
		{Key: "failing", Kind: fakeKind, Enabled: true},
		{Key: "off", Kind: fakeKind},
	}}
}

func identityReq(provider, state string) *systempb.LoginByIdentityRequest {
	code := "c"
	return &systempb.LoginByIdentityRequest{Provider: &provider, State: &state,
		Payload: &systempb.LoginByIdentityRequest_Wechat{Wechat: &systempb.WeChatIdentityPayload{Code: &code}}}
}

// issueState mints a begin-issued state directly into the fake store — the
// law under test is the cash, not the mint (Begin covers that half).
func issueState(t *testing.T, trp *fakeTokenRP, provider string) string {
	t.Helper()
	state := "st-" + provider
	if err := trp.Set(context.Background(), identityStatePrefix+state, []byte(provider), identityStateTTL); err != nil {
		t.Fatalf("state: %v", err)
	}
	return state
}

func boundUser(t *testing.T, seats []*user.Seat) *fakeUserRP {
	t.Helper()
	row := seedRow(t, "bound@x.co", "pw")
	row.ID = "u9"
	return &fakeUserRP{
		row:   row,
		seats: seats,
		credentials: map[string]*user.Credential{
			fakeKind + "|oid-x": {Kind: fakeKind, Identifier: "oid-x", UserID: "u9"},
		},
	}
}

func identityUC(t *testing.T, uf *fakeUserRP) (*AuthUC, *fakeTokenRP) {
	t.Helper()
	trp := &fakeTokenRP{vals: map[string][]byte{}, ttls: map[string]time.Duration{}}
	return newAuthUCConf(fakeProviderConf(), uf, &fakeAuthzRP{}, trp,
		&fakeTenantRP{row: &tenant.Tenant{Code: "t1", Enabled: true}}), trp
}

// The exchange resolves the identity, the binding resolves the row, and the
// session mints from the seat — the same derivation and minter as the
// password plane; the face is population-agnostic because the seat decides.
func TestLoginByIdentityMintsFromSeat(t *testing.T) {
	trp := &fakeTokenRP{vals: map[string][]byte{}, ttls: map[string]time.Duration{}}
	uf := boundUser(t, []*user.Seat{{UserID: "u9", TenantID: "t1", Population: "guest"}})
	uc := newAuthUCConf(fakeProviderConf(), uf, &fakeAuthzRP{}, trp,
		&fakeTenantRP{row: &tenant.Tenant{Code: "t1", Enabled: true}})
	state := issueState(t, trp, "fake")

	token, _, _, err := uc.LoginByIdentity(context.Background(), "fake", state, identityReq("fake", state))
	if err != nil {
		t.Fatalf("login-by-identity: %v", err)
	}
	sess, err := utils.Unmarshal[Session](trp.vals[sessionPrefix+token])
	if err != nil {
		t.Fatal(err)
	}
	if sess.UserID != "u9" || sess.TenantID != "t1" || sess.Population != "guest" {
		t.Fatalf("mint: session = %+v, want the seat's t1/guest for u9", sess)
	}
}

// The one-time state is the begin-to-callback binding: a miss, a state
// minted for another provider, and a second cash of the same state all read
// as the same failed exchange — the defense confirms nothing — and each
// refusal lands before the provider seam, so an injected code is never
// spent on a refused cash.
func TestLoginByIdentityStateLaws(t *testing.T) {
	uf := boundUser(t, []*user.Seat{{UserID: "u9", TenantID: "t1", Population: "guest"}})
	uc, trp := identityUC(t, uf)

	before := fakeExchanges
	if _, _, _, err := uc.LoginByIdentity(context.Background(), "fake", "st-missing", identityReq("fake", "st-missing")); !systempb.IsSystemIdentityExchangeFailed(err) {
		t.Fatalf("missing state: want EXCHANGE_FAILED, got %v", err)
	}
	if fakeExchanges != before {
		t.Fatal("missing state: the exchange ran on a refused cash")
	}

	other := issueState(t, trp, "failing")
	before = fakeExchanges
	if _, _, _, err := uc.LoginByIdentity(context.Background(), "fake", other, identityReq("fake", other)); !systempb.IsSystemIdentityExchangeFailed(err) {
		t.Fatalf("foreign state: want EXCHANGE_FAILED, got %v", err)
	}
	if fakeExchanges != before {
		t.Fatal("foreign state: the exchange ran on a refused cash")
	}

	state := issueState(t, trp, "fake")
	if _, _, _, err := uc.LoginByIdentity(context.Background(), "fake", state, identityReq("fake", state)); err != nil {
		t.Fatalf("first cash: %v", err)
	}
	if _, _, _, err := uc.LoginByIdentity(context.Background(), "fake", state, identityReq("fake", state)); !systempb.IsSystemIdentityExchangeFailed(err) {
		t.Fatalf("second cash: want EXCHANGE_FAILED (single use), got %v", err)
	}
}

// Begin mints the state into the store and rides it on the redirect — the
// authorize URL a client navigates carries the proof its callback must cash.
func TestBeginIdentityLoginIssuesState(t *testing.T) {
	uc, trp := identityUC(t, &fakeUserRP{})

	url, err := uc.BeginIdentityLogin(context.Background(), "fake", "https://app.example/cb")
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	if !strings.HasPrefix(url, "https://idp.example/auth/fake?state=") {
		t.Fatalf("begin: url = %q", url)
	}
	state := strings.TrimPrefix(url, "https://idp.example/auth/fake?state=")
	if state == "" {
		t.Fatal("begin: empty state")
	}
	if got, err := trp.Get(context.Background(), identityStatePrefix+state); err != nil || string(got) != "fake" {
		t.Fatalf("begin: stored state = %q err=%v, want provider key", got, err)
	}
	if ttl := trp.ttls[identityStatePrefix+state]; ttl != 10*time.Minute {
		t.Fatalf("begin: state ttl = %v, want 10m", ttl)
	}

	// An unknown provider refuses before any state is minted — a refused
	// begin must leave zero one-time keys behind.
	if _, err := uc.BeginIdentityLogin(context.Background(), "ghost", "https://app.example/cb"); !systempb.IsSystemIdentityProviderUnknown(err) {
		t.Fatalf("begin unknown provider: want PROVIDER_UNKNOWN, got %v", err)
	}
	if len(trp.vals) != 1 {
		t.Fatalf("begin unknown provider: %d keys stored, want only the first mint", len(trp.vals))
	}
}

// A missing key and a disabled instance answer the same way — the face
// never confirms which instances are configured.
func TestLoginByIdentityUnknownProvider(t *testing.T) {
	uc, _ := identityUC(t, &fakeUserRP{})

	for _, key := range []string{"ghost", "off"} {
		if _, _, _, err := uc.LoginByIdentity(context.Background(), key, "st-x", identityReq(key, "st-x")); !systempb.IsSystemIdentityProviderUnknown(err) {
			t.Fatalf("provider %q: want PROVIDER_UNKNOWN, got %v", key, err)
		}
	}
}

// A completed exchange with no binding reads as a failed exchange — the
// face never confirms which external identities exist. The rows here leave
// the auto-register gate closed; the gated row births instead.
func TestLoginByIdentityUnboundFailsAsExchange(t *testing.T) {
	uc, trp := identityUC(t, &fakeUserRP{})
	state := issueState(t, trp, "fake")

	if _, _, _, err := uc.LoginByIdentity(context.Background(), "fake", state, identityReq("fake", state)); !systempb.IsSystemIdentityExchangeFailed(err) {
		t.Fatalf("unbound: want EXCHANGE_FAILED, got %v", err)
	}
}

// autoUC wires the gated row under test: the exchange resolves, nothing
// binds, and the provider row opens the auto-register gate.
func autoUC(t *testing.T, uf *fakeUserRP) (*AuthUC, *fakeTokenRP) {
	t.Helper()
	// The birth owns its seat — no degenerate default for the gated face.
	if uf.seats == nil {
		uf.seats = []*user.Seat{}
	}
	trp := &fakeTokenRP{vals: map[string][]byte{}, ttls: map[string]time.Duration{}}
	return newAuthUCConf(&conf.Auth{Providers: []*conf.Auth_Provider{
		{Key: "auto", Kind: fakeKind, Enabled: true, AutoRegister: true},
	}}, uf, &fakeAuthzRP{}, trp, &fakeTenantRP{row: &tenant.Tenant{Code: "t1", Enabled: true}}), trp
}

// The auto-register gate births the member the C-end product promises: an
// unbound exchange on a gated row creates the synthetic-email user — no
// password, so the provider is the only way in — ensures the default seat,
// binds the credential, and mints from that seat, the same derivation and
// minter as every other login face.
func TestLoginByIdentityAutoRegisterBirthsMember(t *testing.T) {
	uf := &fakeUserRP{}
	uc, trp := autoUC(t, uf)
	state := issueState(t, trp, "auto")

	token, _, _, err := uc.LoginByIdentity(context.Background(), "auto", state, identityReq("auto", state))
	if err != nil {
		t.Fatalf("auto-register: %v", err)
	}
	if len(uf.createdUsers) != 1 {
		t.Fatalf("auto-register: %d users created, want the one birth", len(uf.createdUsers))
	}
	born := uf.createdUsers[0]
	if born.Email == nil || *born.Email != fakeKind+".oid-x@id.local" {
		t.Fatalf("auto-register: email = %v, want the synthetic %s.oid-x@id.local", born.Email, fakeKind)
	}
	if born.PasswordHash == nil || *born.PasswordHash != "" {
		t.Fatalf("auto-register: hash = %v, want empty for provider-only auth", born.PasswordHash)
	}
	if !born.Enabled {
		t.Fatal("auto-register: the birth lands disabled")
	}
	if len(uf.seats) != 1 || uf.seats[0].UserID != born.ID ||
		uf.seats[0].TenantID != security.DefaultTenant || uf.seats[0].Population != security.DefaultPopulation {
		t.Fatalf("auto-register: seats = %+v, want the default seat for %s", uf.seats, born.ID)
	}
	if len(uf.createdCreds) != 1 || uf.createdCreds[0].UserID != born.ID ||
		uf.createdCreds[0].Kind != fakeKind || uf.createdCreds[0].Identifier != "oid-x" {
		t.Fatalf("auto-register: credentials = %+v, want the identity bound to the birth", uf.createdCreds)
	}
	sess, err := utils.Unmarshal[Session](trp.vals[sessionPrefix+token])
	if err != nil {
		t.Fatal(err)
	}
	if sess.UserID != born.ID || sess.Population != security.DefaultPopulation {
		t.Fatalf("auto-register: session = %+v, want the born member", sess)
	}
}

// A synthetic-email row whose credential is gone (a deleted binding)
// converges instead of duplicating: the existing user is reused, the seat
// ensure stays idempotent, and the fresh binding re-attaches the identity.
func TestLoginByIdentityAutoRegisterConverges(t *testing.T) {
	row := seedRow(t, fakeKind+".oid-x@id.local", "pw")
	row.ID = "u-syn"
	uf := &fakeUserRP{row: row}
	uc, trp := autoUC(t, uf)
	state := issueState(t, trp, "auto")

	token, _, _, err := uc.LoginByIdentity(context.Background(), "auto", state, identityReq("auto", state))
	if err != nil {
		t.Fatalf("converge: %v", err)
	}
	if len(uf.createdUsers) != 0 {
		t.Fatalf("converge: %d users created, want the existing row reused", len(uf.createdUsers))
	}
	if len(uf.createdCreds) != 1 || uf.createdCreds[0].UserID != "u-syn" {
		t.Fatalf("converge: credentials = %+v, want the identity re-bound to u-syn", uf.createdCreds)
	}
	sess, err := utils.Unmarshal[Session](trp.vals[sessionPrefix+token])
	if err != nil {
		t.Fatal(err)
	}
	if sess.UserID != "u-syn" {
		t.Fatalf("converge: session user = %s, want u-syn", sess.UserID)
	}
}

// The provider's own rejection surfaces in its shape.
func TestLoginByIdentityExchangeFailureSurfaces(t *testing.T) {
	uc, trp := identityUC(t, &fakeUserRP{})
	state := issueState(t, trp, "failing")

	if _, _, _, err := uc.LoginByIdentity(context.Background(), "failing", state, identityReq("failing", state)); !systempb.IsSystemIdentityExchangeFailed(err) {
		t.Fatalf("exchange failure: want EXCHANGE_FAILED, got %v", err)
	}
}

// Several seats refuse with the seat-required reason — the client chooses
// the seat; the server never guesses. Same law as the password plane.
func TestLoginByIdentityMultiSeatRequiresChoice(t *testing.T) {
	uf := boundUser(t, []*user.Seat{
		{UserID: "u9", TenantID: "t1", Population: "user"},
		{UserID: "u9", TenantID: "t2", Population: "member"},
	})
	uc, trp := identityUC(t, uf)
	state := issueState(t, trp, "fake")

	if _, _, _, err := uc.LoginByIdentity(context.Background(), "fake", state, identityReq("fake", state)); !systempb.IsSystemLoginSeatRequired(err) {
		t.Fatalf("multi-seat: want SEAT_REQUIRED, got %v", err)
	}
}

// Binding attaches the exchanged identity to the caller; rebinding the
// same identity is cosmetic, another identity's binding is refused, and
// the unique is the race backstop of the same shape. The state law mirrors
// the login face — an injected code must not bind.
func TestBindCredentialLaws(t *testing.T) {
	code, key := "c", "fake"
	newReq := func(state string) *systempb.BindCredentialRequest {
		return &systempb.BindCredentialRequest{Provider: &key, State: &state,
			Payload: &systempb.BindCredentialRequest_Wechat{Wechat: &systempb.WeChatIdentityPayload{Code: &code}}}
	}
	ctx := security.WithSubject(context.Background(), &security.Subject{UserID: "u1", TenantID: "t1"})

	uf := &fakeUserRP{}
	uc, trp := identityUC(t, uf)
	before := fakeExchanges
	if err := uc.BindCredential(ctx, "fake", "st-missing", newReq("st-missing")); !systempb.IsSystemIdentityExchangeFailed(err) {
		t.Fatalf("stateless bind: want EXCHANGE_FAILED, got %v", err)
	}
	if fakeExchanges != before || len(uf.createdCreds) != 0 {
		t.Fatal("stateless bind: the exchange ran or a credential landed")
	}

	state := issueState(t, trp, "fake")
	if err := uc.BindCredential(ctx, "fake", state, newReq(state)); err != nil {
		t.Fatalf("bind: %v", err)
	}
	if len(uf.createdCreds) != 1 || uf.createdCreds[0].UserID != "u1" {
		t.Fatalf("bind: created = %+v, want one binding to u1", uf.createdCreds)
	}

	bind := func(uf *fakeUserRP) error {
		uc, trp := identityUC(t, uf)
		state := issueState(t, trp, "fake")
		return uc.BindCredential(ctx, "fake", state, newReq(state))
	}
	same := &fakeUserRP{credentials: map[string]*user.Credential{
		fakeKind + "|oid-x": {Kind: fakeKind, Identifier: "oid-x", UserID: "u1"},
	}}
	if err := bind(same); err != nil {
		t.Fatalf("rebind: %v", err)
	}
	if len(same.createdCreds) != 0 {
		t.Fatal("rebind: no second binding")
	}

	other := &fakeUserRP{credentials: map[string]*user.Credential{
		fakeKind + "|oid-x": {Kind: fakeKind, Identifier: "oid-x", UserID: "someone"},
	}}
	if err := bind(other); !systempb.IsSystemIdentityAlreadyBound(err) {
		t.Fatalf("conflict: want ALREADY_BOUND, got %v", err)
	}

	race := &fakeUserRP{createCredErr: errorspb.ErrorInfraErrorDbConstraint("")}
	if err := bind(race); !systempb.IsSystemIdentityAlreadyBound(err) {
		t.Fatalf("race: want ALREADY_BOUND, got %v", err)
	}

	noUC, noTrp := identityUC(t, &fakeUserRP{})
	noState := issueState(t, noTrp, "fake")
	if err := noUC.BindCredential(context.Background(), "fake", noState, newReq(noState)); !errorspb.IsGeneralErrorUnauthenticated(err) {
		t.Fatalf("no session: want 401, got %v", err)
	}
}

// A store outage during the cash surfaces as infra — the state law must not
// mask a broken dependency behind a failed exchange.
func TestLoginByIdentityStateCashInfraPassthrough(t *testing.T) {
	uf := boundUser(t, []*user.Seat{{UserID: "u9", TenantID: "t1", Population: "guest"}})
	uc, trp := identityUC(t, uf)
	state := issueState(t, trp, "fake")
	trp.getErr = errorspb.ErrorInfraErrorNetworkConnection("")

	if _, _, _, err := uc.LoginByIdentity(context.Background(), "fake", state, identityReq("fake", state)); !errorspb.IsInfraErrorNetworkConnection(err) {
		t.Fatalf("cash outage: want infra passthrough, got %v", err)
	}
}

// Provider rows are wiring facts: a row referencing an unregistered kind
// fails the construction, not the first login.
func TestNewAuthUCRejectsUnregisteredKind(t *testing.T) {
	bad := &conf.Auth{Providers: []*conf.Auth_Provider{{Key: "x", Kind: "ghost"}}}
	if _, err := NewAuthUC(slog.New(slog.NewTextHandler(io.Discard, nil)), nil, bad, nil, nil, nil, nil); err == nil {
		t.Fatal("unregistered kind: want construction error")
	}
}
