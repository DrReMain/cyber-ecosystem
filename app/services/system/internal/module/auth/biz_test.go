package auth

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"cyber-ecosystem/shared-go/kratos/security"
	"cyber-ecosystem/shared-go/utils"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/conf"
	"cyber-ecosystem/app/services/system/internal/module/authz"
	"cyber-ecosystem/app/services/system/internal/module/population"
	"cyber-ecosystem/app/services/system/internal/module/tenant"
	"cyber-ecosystem/app/services/system/internal/module/user"
	"cyber-ecosystem/app/services/system/internal/shared"
)

// fakeTokenRP backs the session store with a map; TTLs and Expire calls are
// recorded per key so the refresh window, the absolute cap, and the revocation
// marker each land on their real branch. A miss answers the classified
// CacheMiss error the real adapter would — fakes simulate the classification
// layer, never raw errors.
type fakeTokenRP struct {
	vals      map[string][]byte
	ttls      map[string]time.Duration
	expire    []time.Duration
	expireErr error
	getErr    error
}

func newTokenStore() *fakeTokenRP {
	return &fakeTokenRP{vals: map[string][]byte{}, ttls: map[string]time.Duration{}}
}

func (f *fakeTokenRP) Get(_ context.Context, key string) ([]byte, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	if v, ok := f.vals[key]; ok {
		return v, nil
	}
	return nil, errorspb.ErrorInfraErrorCacheMiss("")
}

func (f *fakeTokenRP) GetDel(ctx context.Context, key string) ([]byte, error) {
	v, err := f.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	delete(f.vals, key)
	delete(f.ttls, key)
	return v, nil
}

func (f *fakeTokenRP) Set(_ context.Context, key string, val []byte, ttl time.Duration) error {
	f.vals[key] = val
	f.ttls[key] = ttl
	return nil
}

func (f *fakeTokenRP) Del(_ context.Context, key string) error {
	delete(f.vals, key)
	delete(f.ttls, key)
	return nil
}

func (f *fakeTokenRP) Expire(_ context.Context, key string, ttl time.Duration) error {
	if f.expireErr != nil {
		return f.expireErr
	}
	f.expire = append(f.expire, ttl)
	f.ttls[key] = ttl
	return nil
}

// fakeUserRP backs the Login slice: one seeded row, account lookup. updates
// captures what the lazy rehash writes. Seats is the seat store; an explicit
// empty slice means seatless (a nil slice takes the helper's degenerate
// default). credentials is the binding store keyed "kind|identifier";
// createdUsers records the auto-register births.
type fakeUserRP struct {
	user.UserRP
	row     *user.User
	findErr error
	updates []struct {
		fields []string
		u      *user.User
	}
	updateErr error
	seats     []*user.Seat
	seatsErr  error

	credentials    map[string]*user.Credential
	createCredErr  error
	createdCreds   []*user.Credential
	deletedCredFor []string
	findIDErr      error
	findByIDCalls  int

	// merge slice: the poor side's writes, the one-shot ledger, and the
	// follow-ups the disable path owes.
	credCounts map[string]int
	moved      []string
	moveErr    error
	merges     []*user.AccountMerge
	mergeErr   error
	disabled   []string
	revoked    []string
	notified   int

	createdUsers []*user.User
}

func (f *fakeUserRP) ListSeatsByUser(_ context.Context, _ string) ([]*user.Seat, error) {
	if f.seatsErr != nil {
		return nil, f.seatsErr
	}
	return f.seats, nil
}

func (f *fakeUserRP) FindCredentialByIdentifier(_ context.Context, identifier string) (*user.Credential, error) {
	for k, c := range f.credentials {
		if _, id, ok := strings.Cut(k, "|"); ok && id == identifier {
			return c, nil
		}
	}
	return nil, errorspb.ErrorInfraErrorDbNotFound("")
}

func (f *fakeUserRP) FindCredential(_ context.Context, kind, identifier string) (*user.Credential, error) {
	if c, ok := f.credentials[kind+"|"+identifier]; ok {
		return c, nil
	}
	return nil, errorspb.ErrorInfraErrorDbNotFound("")
}

func (f *fakeUserRP) CreateCredential(_ context.Context, c *user.Credential) error {
	if f.createCredErr != nil {
		return f.createCredErr
	}
	f.createdCreds = append(f.createdCreds, c)
	return nil
}

func (f *fakeUserRP) DeleteCredentialsByUser(_ context.Context, userID string) error {
	f.deletedCredFor = append(f.deletedCredFor, userID)
	return nil
}

func (f *fakeUserRP) CountCredentialsByUser(_ context.Context, userID string) (int, error) {
	if f.credCounts != nil {
		return f.credCounts[userID], nil
	}
	n := 0
	for _, c := range f.credentials {
		if c.UserID == userID {
			n++
		}
	}
	return n, nil
}

func (f *fakeUserRP) ListCredentialsByUser(_ context.Context, userID string) ([]*user.Credential, error) {
	var out []*user.Credential
	for _, c := range f.credentials {
		if c.UserID == userID {
			out = append(out, c)
		}
	}
	return out, nil
}

func (f *fakeUserRP) MoveCredential(_ context.Context, kind, identifier, toUserID string) error {
	if f.moveErr != nil {
		return f.moveErr
	}
	c, ok := f.credentials[kind+"|"+identifier]
	if !ok {
		return errorspb.ErrorInfraErrorDbNotFound("")
	}
	c.UserID = toUserID
	f.moved = append(f.moved, kind+"|"+identifier+"->"+toUserID)
	return nil
}

func (f *fakeUserRP) CreateAccountMerge(_ context.Context, m *user.AccountMerge) error {
	if f.mergeErr != nil {
		return f.mergeErr
	}
	f.merges = append(f.merges, m)
	return nil
}

func (f *fakeUserRP) UpdateStatus(_ context.Context, id string, enabled bool) (*user.User, error) {
	if !enabled {
		f.disabled = append(f.disabled, id)
	}
	if f.row != nil && f.row.ID == id {
		row := *f.row
		row.Enabled = enabled
		f.row = &row
		return &row, nil
	}
	return &user.User{ID: id, Enabled: enabled}, nil
}

func (f *fakeUserRP) RevokeUserSessions(_ context.Context, userID string) error {
	f.revoked = append(f.revoked, userID)
	return nil
}

func (f *fakeUserRP) NotifyChanged(context.Context) {
	f.notified++
}

func (f *fakeUserRP) FindByAccount(_ context.Context, account string) (*user.User, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	if f.row != nil && *f.row.Account == account {
		return f.row, nil
	}
	return nil, errorspb.ErrorInfraErrorDbNotFound("")
}

func (f *fakeUserRP) Create(_ context.Context, u *user.User) (*user.User, error) {
	u.ID = "u-born"
	f.createdUsers = append(f.createdUsers, u)
	return u, nil
}

// EnsureSeat mirrors the ensure semantics: the seat becomes visible to the
// seat list, and a second ensure for the same row is idempotent noise.
func (f *fakeUserRP) EnsureSeat(_ context.Context, s *user.Seat) error {
	f.seats = append(f.seats, s)
	return nil
}

func (f *fakeUserRP) FindByID(_ context.Context, id string) (*user.User, error) {
	f.findByIDCalls++
	if f.findIDErr != nil {
		return nil, f.findIDErr
	}
	if f.row != nil && f.row.ID == id {
		return f.row, nil
	}
	return nil, errorspb.ErrorInfraErrorDbNotFound("")
}

func (f *fakeUserRP) Update(_ context.Context, fieldsMask []string, u *user.User) (*user.User, error) {
	f.updates = append(f.updates, struct {
		fields []string
		u      *user.User
	}{fieldsMask, u})
	return u, f.updateErr
}

// fakeAuthzRP answers the me face's grant resolution with a fixed pattern set
// — or the injected engine error, for the infra-failure path. planes answers
// the vocabulary read session verification stamps onto subjects; populations
// absent from the map read as backend.
type fakeAuthzRP struct {
	authz.AuthzRP
	grants    []authz.GrantView
	err       error
	planes    map[string]string
	stamp     int64
	stampErr  error
	stampHits int
}

// StampVersion answers the liveness-stamp channel the verification path
// compares the payload against; stampErr simulates the channel failing so
// the deep-read fallback stays exercised.
func (f *fakeAuthzRP) StampVersion(_ context.Context) (int64, error) {
	f.stampHits++
	if f.stampErr != nil {
		return 0, f.stampErr
	}
	return f.stamp, nil
}

func (f *fakeAuthzRP) EffectiveGrants(_ context.Context, _ *security.Subject) ([]authz.GrantView, error) {
	return f.grants, f.err
}

func (f *fakeAuthzRP) PlaneOf(population string) (string, error) {
	if population == "" {
		return security.PlaneBackend, nil
	}
	if plane, ok := f.planes[population]; ok {
		return plane, nil
	}
	return security.PlaneBackend, nil
}

// fakeTenantRP answers the registry read auth consults: an injectable row
// (nil = the classified miss) and an injectable error.
type fakeTenantRP struct {
	tenant.TenantRP
	row            *tenant.Tenant
	err            error
	findByCodeCall int
}

func (f *fakeTenantRP) FindByCode(_ context.Context, code string) (*tenant.Tenant, error) {
	f.findByCodeCall++
	if f.err != nil {
		return nil, f.err
	}
	if f.row == nil {
		return nil, errorspb.ErrorInfraErrorDbNotFound("")
	}
	return f.row, nil
}

// fakePopulationRP answers the registry read a declared landing costs: rows
// keyed by code, a miss classified as not-found the adapter would return.
type fakePopulationRP struct {
	population.PopulationRP
	rows map[string]*population.Population
}

func (f *fakePopulationRP) FindByCode(_ context.Context, code string) (*population.Population, error) {
	if p, ok := f.rows[code]; ok {
		return p, nil
	}
	return nil, errorspb.ErrorInfraErrorDbNotFound("")
}

// passthroughTx runs the unit inline — the auth slice asserts transaction
// semantics through the fakes, not a real store.
type passthroughTx struct{}

func (passthroughTx) InTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// fakeIdentityProvider answers the exchange seam with a fixed identity; an
// instance keyed "failing" rejects in the domain shape. Registered once
// under its own kind. fakeExchanges counts calls into the exchange seam;
// ordering laws assert it stays put where a refused cash must never reach
// the provider.
type fakeIdentityProvider struct {
	id *exchangedIdentity
}

var fakeExchanges int

func (f fakeIdentityProvider) Exchange(_ context.Context, row *conf.Auth_Provider, _ any) (*exchangedIdentity, error) {
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

func testProviders() map[string]IdentityProvider {
	return map[string]IdentityProvider{
		fakeKind: fakeIdentityProvider{id: &exchangedIdentity{Kind: fakeKind, Identifier: "oid-x",
			Profile: &wechatProfile{Nickname: "昵称", AvatarURL: "https://qlogo.cn/a.png"}}},
		// The resolve phone branch never exchanges; the row only has to
		// carry the channel kind the branch guards on.
		smsKind: fakeIdentityProvider{id: &exchangedIdentity{Kind: smsKind, Identifier: "13900001111"}},
	}
}

func fakeProviderConf() *conf.Auth {
	return &conf.Auth{Providers: []*conf.Auth_Provider{
		{Key: "fake", Kind: fakeKind, Enabled: true},
		{Key: "failing", Kind: fakeKind, Enabled: true},
		{Key: "off", Kind: fakeKind},
		{Key: "sms", Kind: smsKind, Enabled: true},
	}}
}

// Private -------------------------------------------------------------------------------------------------------------

func newAuthUC(u *fakeUserRP, a *fakeAuthzRP, t *fakeTokenRP) *AuthUC {
	return newAuthUCTenant(u, a, t, &fakeTenantRP{row: &tenant.Tenant{Code: "any", Enabled: true}})
}

func newAuthUCTenant(u *fakeUserRP, a *fakeAuthzRP, t *fakeTokenRP, tn *fakeTenantRP) *AuthUC {
	return newAuthUCConf(nil, u, a, t, tn)
}

func newAuthUCConf(auth *conf.Auth, u *fakeUserRP, a *fakeAuthzRP, t *fakeTokenRP, tn *fakeTenantRP) *AuthUC {
	// The default seat mirrors the degenerate world: one seat on t1 for u1.
	if u.seats == nil {
		u.seats = []*user.Seat{{UserID: "u1", TenantID: "t1", Population: security.DefaultPopulation}}
	}
	uc, err := NewAuthUC(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		passthroughTx{}, auth, testProviders(), u, a, tn,
		&fakePopulationRP{rows: map[string]*population.Population{
			"member": {Code: "member", Name: "Member", Plane: security.PlaneMember},
		}},
		t,
	)
	if err != nil {
		panic(err)
	}
	return uc
}

func seedRow(t *testing.T, account, password string) *user.User {
	t.Helper()
	hash, err := utils.Hash(password)
	if err != nil {
		t.Fatalf("seed hash: %v", err)
	}
	return &user.User{
		ID: "u1", Account: &account, PasswordHash: &hash, Enabled: true,
	}
}

// mintSession writes a session key directly through the fake store — the
// authenticate-family tests own their session shapes. A zero LivenessStamp
// becomes -1, the shape a mint writes when the stamp channel failed: the
// legacy specs then ride the deep reads, exactly as they did before the
// stamp path existed.
func mintSession(t *testing.T, rp *fakeTokenRP, sess *Session) string {
	t.Helper()
	if sess.LivenessStamp == 0 {
		sess.LivenessStamp = -1
	}
	token, err := utils.RandomToken()
	if err != nil {
		t.Fatalf("token: %v", err)
	}
	rp.vals[sessionPrefix+token] = utils.MustMarshal(sess)
	rp.ttls[sessionPrefix+token] = sessionIdle
	return token
}

func authCtx(subject *security.Subject) context.Context {
	return security.WithSubject(context.Background(), subject)
}

func identityReq(provider, state string) *systempb.LoginByIdentityRequest {
	code := "c"
	return &systempb.LoginByIdentityRequest{Provider: &provider, State: &state,
		Payload: &systempb.LoginByIdentityRequest_Wechat{Wechat: &systempb.WeChatIdentityPayload{Code: &code}}}
}

// issueState mints a begin-issued state directly into the fake store — the
// law under test is the cash, not the mint (Begin covers that half). The
// anonymous shape is what the public login rungs mint; boundState is the
// bind-rung shape carrying the minting session.
func issueState(t *testing.T, trp *fakeTokenRP, provider string) string {
	t.Helper()
	return issueBoundState(t, trp, provider, "")
}

func issueBoundState(t *testing.T, trp *fakeTokenRP, provider, bindUser string) string {
	t.Helper()
	state := "st-" + provider
	if bindUser != "" {
		state += "-" + bindUser
	}
	val := utils.MustMarshal(identityStateValue{Provider: provider, BindUser: bindUser})
	if err := trp.Set(context.Background(), identityStatePrefix+state, val, identityStateTTL); err != nil {
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
	trp := newTokenStore()
	return newAuthUCConf(fakeProviderConf(), uf, &fakeAuthzRP{}, trp,
		&fakeTenantRP{row: &tenant.Tenant{Code: "t1", Enabled: true}}), trp
}

// autoUC wires the gated row under test: the exchange resolves, nothing
// binds, and the provider row opens the auto-register gate.
func autoUC(t *testing.T, uf *fakeUserRP) (*AuthUC, *fakeTokenRP) {
	t.Helper()
	// The birth owns its seat — no degenerate default for the gated face.
	if uf.seats == nil {
		uf.seats = []*user.Seat{}
	}
	trp := newTokenStore()
	return newAuthUCConf(&conf.Auth{Providers: []*conf.Auth_Provider{
		{Key: "auto", Kind: fakeKind, Enabled: true, AutoRegister: true, Population: "member"},
	}}, uf, &fakeAuthzRP{}, trp, &fakeTenantRP{row: &tenant.Tenant{Code: "t1", Enabled: true}}), trp
}

// Login ---------------------------------------------------------------------------------------------------------------

// Anti-enumeration: an unknown account and a wrong password answer in the same
// shape — which accounts exist is not probeable from the wire (the not-found
// branch burns the dummy digest so timing carries no signal either; that half
// is not unit-assertable, the shape half is).
func TestLoginUnknownAccountSameShapeAsWrongPassword(t *testing.T) {
	uc := newAuthUC(&fakeUserRP{row: seedRow(t, "a@x.co", "right")}, &fakeAuthzRP{}, newTokenStore())

	_, _, _, errUnknown := uc.Login(context.Background(), "ghost@x.co", "right", "")
	if !systempb.IsSystemLoginFailed(errUnknown) {
		t.Fatalf("unknown account: want LOGIN_FAILED, got %v", errUnknown)
	}
	_, _, _, errWrong := uc.Login(context.Background(), "a@x.co", "wrong", "")
	if !systempb.IsSystemLoginFailed(errWrong) {
		t.Fatalf("wrong password: want LOGIN_FAILED, got %v", errWrong)
	}
}

// The credential fallback: an SMS-born member's login handle is their
// phone — a bound credential, not an account handle. The handle misses,
// the identifier resolves, the password answers. Kind-agnostic on
// purpose: whatever kind the identifier bound through names the same
// account, so a future email binding rides this path with zero new
// branches.
func TestLoginCredentialIdentifierFallback(t *testing.T) {
	handle := "dbn18mlthm88rvujd2t0"
	phone := "13800001111"
	digest, err := utils.Hash("right")
	if err != nil {
		t.Fatalf("hash seed: %v", err)
	}
	row := &user.User{ID: "u-sms-born", Account: &handle, PasswordHash: &digest, Enabled: true}
	rp := &fakeUserRP{row: row, credentials: map[string]*user.Credential{
		"sms|" + phone: {Kind: "sms", Identifier: phone, UserID: "u-sms-born"},
	}}
	uc := newAuthUC(rp, &fakeAuthzRP{}, newTokenStore())

	if _, _, _, err := uc.Login(context.Background(), phone, "right", ""); err != nil {
		t.Fatalf("phone + password on sms-born member: %v", err)
	}
	if _, _, _, err := uc.Login(context.Background(), phone, "wrong", ""); !systempb.IsSystemLoginFailed(err) {
		t.Fatalf("wrong password via credential: want LOGIN_FAILED, got %v", err)
	}
	if _, _, _, err := uc.Login(context.Background(), "13900002222", "right", ""); !systempb.IsSystemLoginFailed(err) {
		t.Fatalf("identifier no credential knows: want LOGIN_FAILED, got %v", err)
	}
}

// The deeper order law: a disabled account with a wrong password still reports
// LOGIN_FAILED — status only surfaces once the credential passes, so account
// state is as unprobeable as account existence.
func TestLoginDisabledUserStatusAfterCredential(t *testing.T) {
	row := seedRow(t, "a@x.co", "right")
	row.Enabled = false
	uc := newAuthUC(&fakeUserRP{row: row}, &fakeAuthzRP{}, newTokenStore())

	if _, _, _, err := uc.Login(context.Background(), "a@x.co", "wrong", ""); !systempb.IsSystemLoginFailed(err) {
		t.Fatalf("wrong password on disabled account: want LOGIN_FAILED, got %v", err)
	}
	if _, _, _, err := uc.Login(context.Background(), "a@x.co", "right", ""); !systempb.IsSystemUserDisabled(err) {
		t.Fatalf("right password on disabled account: want USER_DISABLED, got %v", err)
	}
}

// Lazy digest upgrade: a legacy bcrypt row verifies, and the same login
// re-hashes it as argon2id; a failed upgrade warns and never blocks the login.
func TestLoginRehashesLegacyBcrypt(t *testing.T) {
	legacy, err := bcrypt.GenerateFromPassword([]byte("right"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt seed: %v", err)
	}
	account := "a@x.co"
	legacyStr := string(legacy)
	row := &user.User{ID: "u1", Account: &account, PasswordHash: &legacyStr, Enabled: true}
	rp := &fakeUserRP{row: row}
	uc := newAuthUC(rp, &fakeAuthzRP{}, newTokenStore())

	if _, _, _, err := uc.Login(context.Background(), "a@x.co", "right", ""); err != nil {
		t.Fatalf("legacy verify: %v", err)
	}
	if len(rp.updates) != 1 {
		t.Fatalf("rehash: want 1 update, got %d", len(rp.updates))
	}
	newHash := *rp.updates[0].u.PasswordHash
	if utils.NeedsRehash(newHash) || newHash == string(legacy) {
		t.Fatalf("rehash: want an upgraded argon2id digest, got %q", newHash)
	}

	// A failing upgrade is a warning, never a login blocker.
	rp.updateErr = errorspb.ErrorInfraErrorNetworkConnection("")
	if _, _, _, err := uc.Login(context.Background(), "a@x.co", "right", ""); err != nil {
		t.Fatalf("upgrade failure must not block login: %v", err)
	}
}

// The mint's shape is the login contract: the absolute cap derived from
// SessionMax, the idle TTL on the key, and the row's identity in the payload.
func TestLoginMintsSessionShape(t *testing.T) {
	row := seedRow(t, "a@x.co", "right")
	trp := newTokenStore()
	uc := newAuthUC(&fakeUserRP{row: row}, &fakeAuthzRP{}, trp)

	token, issuedAt, expiresAt, err := uc.Login(context.Background(), "a@x.co", "right", "")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if token == "" {
		t.Fatal("mint: empty token")
	}
	if d := expiresAt.Sub(issuedAt); d != shared.SessionMax {
		t.Fatalf("mint: lifetime = %v, want SessionMax", d)
	}
	if got := trp.ttls[sessionPrefix+token]; got != sessionIdle {
		t.Fatalf("mint: key TTL = %v, want sessionIdle", got)
	}
	sess, err := utils.Unmarshal[Session](trp.vals[sessionPrefix+token])
	if err != nil {
		t.Fatalf("mint: payload: %v", err)
	}
	if sess.UserID != row.ID || sess.Account != "a@x.co" {
		t.Fatalf("mint: payload = %+v", sess)
	}
}

// A user lookup that fails beyond not-found (store down) passes through — the
// login path never invents a domain shape for infra trouble.
func TestLoginFindByAccountGenericErrorPassthrough(t *testing.T) {
	uc := newAuthUC(&fakeUserRP{findErr: errorspb.ErrorInfraErrorNetworkConnection("")}, &fakeAuthzRP{}, newTokenStore())

	if _, _, _, err := uc.Login(context.Background(), "a@x.co", "pw", ""); !errorspb.IsInfraErrorNetworkConnection(err) {
		t.Fatalf("generic lookup error: want infra passthrough, got %v", err)
	}
}

// The session's tenant and population come from the resolved seat, never
// from the user row's columns — the user row may carry a tenant column for
// row scoping, but the seat is what the login occupies.
func TestLoginMintsFromSeat(t *testing.T) {
	rp := newTokenStore()
	uc := newAuthUCTenant(&fakeUserRP{row: seedRow(t, "a@x.co", "right"),
		seats: []*user.Seat{{UserID: "u1", TenantID: "t9", Population: "user"}}}, &fakeAuthzRP{}, rp,
		&fakeTenantRP{row: &tenant.Tenant{Code: "t9", Enabled: true}})

	token, _, _, err := uc.Login(context.Background(), "a@x.co", "right", "")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	sess, err := utils.Unmarshal[Session](rp.vals[sessionPrefix+token])
	if err != nil {
		t.Fatal(err)
	}
	if sess.TenantID != "t9" || sess.Population != "user" {
		t.Fatalf("mint: session = %+v, want the seat's t9/user", sess)
	}
}

// A seatless identity is a broken invariant, not a login-failure shape —
// the credential already passed, so the corruption surfaces loudly.
func TestLoginSeatlessIsInternal(t *testing.T) {
	uc := newAuthUCTenant(&fakeUserRP{row: seedRow(t, "a@x.co", "right"), seats: []*user.Seat{}}, &fakeAuthzRP{},
		newTokenStore(), &fakeTenantRP{row: &tenant.Tenant{Code: "t1", Enabled: true}})

	if _, _, _, err := uc.Login(context.Background(), "a@x.co", "right", ""); !errorspb.IsGeneralErrorInternal(err) {
		t.Fatalf("seatless: want INTERNAL, got %v", err)
	}
}

// Several seats without a choice refuse with the dedicated reason — the
// client asks which seat; the server never guesses.
func TestLoginMultiSeatRequiresChoice(t *testing.T) {
	uf := &fakeUserRP{row: seedRow(t, "a@x.co", "right"), seats: []*user.Seat{
		{UserID: "u1", TenantID: "t1", Population: "user"},
		{UserID: "u1", TenantID: "t2", Population: "member"},
	}}
	uc := newAuthUCTenant(uf, &fakeAuthzRP{}, newTokenStore(),
		&fakeTenantRP{row: &tenant.Tenant{Code: "t1", Enabled: true}})

	if _, _, _, err := uc.Login(context.Background(), "a@x.co", "right", ""); !systempb.IsSystemLoginSeatRequired(err) {
		t.Fatalf("multi-seat: want SEAT_REQUIRED, got %v", err)
	}
}

// The hint narrows a multi-seat login to the held seat it names — a
// preference among held seats, never a grant: a code the identity does not
// hold keeps the choice required.
func TestLoginHintPicksHeldSeat(t *testing.T) {
	uf := &fakeUserRP{row: seedRow(t, "a@x.co", "right"), seats: []*user.Seat{
		{UserID: "u1", TenantID: "t1", Population: "user"},
		{UserID: "u1", TenantID: "t1", Population: "member"},
	}}
	rp := newTokenStore()
	uc := newAuthUCTenant(uf, &fakeAuthzRP{}, rp, &fakeTenantRP{row: &tenant.Tenant{Code: "t1", Enabled: true}})

	token, _, _, err := uc.Login(context.Background(), "a@x.co", "right", "member")
	if err != nil {
		t.Fatalf("hinted login: %v", err)
	}
	sess, err := utils.Unmarshal[Session](rp.vals[sessionPrefix+token])
	if err != nil {
		t.Fatal(err)
	}
	if sess.Population != "member" {
		t.Fatalf("hinted login: session population = %q, want member", sess.Population)
	}

	if _, _, _, err := uc.Login(context.Background(), "a@x.co", "right", "ghost"); !systempb.IsSystemLoginSeatRequired(err) {
		t.Fatalf("unheld hint: want SEAT_REQUIRED, got %v", err)
	}
}

// Tenant enforcement --------------------------------------------------------------------------------------------------

// The tenant's disabled state surfaces only after the credential passes —
// tenant state is as unprobeable as account existence.
func TestLoginDisabledTenantAfterCredential(t *testing.T) {
	tok := newTokenStore()
	uc := newAuthUCTenant(&fakeUserRP{row: seedRow(t, "a@x.co", "right")}, &fakeAuthzRP{}, tok,
		&fakeTenantRP{row: &tenant.Tenant{Code: "t1", Enabled: false}})

	if _, _, _, err := uc.Login(context.Background(), "a@x.co", "wrong", ""); !systempb.IsSystemLoginFailed(err) {
		t.Fatalf("wrong password on disabled tenant: want LOGIN_FAILED, got %v", err)
	}
	if _, _, _, err := uc.Login(context.Background(), "a@x.co", "right", ""); !systempb.IsSystemTenantDisabled(err) {
		t.Fatalf("right password on disabled tenant: want TENANT_DISABLED, got %v", err)
	}
	if len(tok.vals) != 0 {
		t.Fatal("disabled tenant: no session may be minted")
	}
}

// A registry row that is gone fails closed into the same rejection — login
// keeps one shape.
func TestLoginMissingTenantFailsClosed(t *testing.T) {
	uc := newAuthUCTenant(&fakeUserRP{row: seedRow(t, "a@x.co", "right")}, &fakeAuthzRP{},
		newTokenStore(), &fakeTenantRP{})

	if _, _, _, err := uc.Login(context.Background(), "a@x.co", "right", ""); !systempb.IsSystemTenantDisabled(err) {
		t.Fatalf("missing registry row: want TENANT_DISABLED, got %v", err)
	}
}

// A generic registry error keeps its infra identity.
func TestLoginTenantGenericPassthrough(t *testing.T) {
	uc := newAuthUCTenant(&fakeUserRP{row: seedRow(t, "a@x.co", "right")}, &fakeAuthzRP{},
		newTokenStore(), &fakeTenantRP{err: errorspb.ErrorInfraErrorNetworkConnection("")})

	if _, _, _, err := uc.Login(context.Background(), "a@x.co", "right", ""); !errorspb.IsInfraErrorNetworkConnection(err) {
		t.Fatalf("generic registry error: want passthrough, got %v", err)
	}
}

// A tenant disabled mid-session dies at verification — a check, not a
// revocation: the key stays, re-enabling resurrects the stored session, and a
// missing row answers in the same shape.
func TestAuthenticateSessionTenantGate(t *testing.T) {
	rp := newTokenStore()
	token := mintSession(t, rp, &Session{UserID: "u1", TenantID: "t1", Account: "a@x.co", IssuedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour)})
	tn := &fakeTenantRP{row: &tenant.Tenant{Code: "t1", Enabled: false}}
	az := &fakeAuthzRP{}
	uc := newAuthUCTenant(&fakeUserRP{row: seedRow(t, "a@x.co", "right")}, az, rp, tn)

	if _, err := uc.AuthenticateSession(context.Background(), token); !errorspb.IsGeneralErrorUnauthenticated(err) {
		t.Fatalf("disabled tenant: want unauthenticated, got %v", err)
	}
	if _, ok := rp.vals[sessionPrefix+token]; !ok {
		t.Fatal("disabled tenant: the gate must not revoke — the key stays for resurrection")
	}

	tn.row.Enabled = true
	az.stamp++ // every gate flip moves the version the payload compares against
	if _, err := uc.AuthenticateSession(context.Background(), token); err != nil {
		t.Fatalf("re-enabled tenant: session must resurrect, got %v", err)
	}

	tn.row = nil
	az.stamp++ // a deleted row bumps on its way out
	if _, err := uc.AuthenticateSession(context.Background(), token); !errorspb.IsGeneralErrorUnauthenticated(err) {
		t.Fatalf("missing registry row: want unauthenticated same shape, got %v", err)
	}

	tn.err = errorspb.ErrorInfraErrorNetworkConnection("")
	az.stamp++ // an unrelated bump forces the deep read the error rides in on
	if _, err := uc.AuthenticateSession(context.Background(), token); !errorspb.IsInfraErrorNetworkConnection(err) {
		t.Fatalf("generic registry error: want passthrough, got %v", err)
	}
}

// A disabled identity dies at verification the same way its tenant does —
// a check, not a revocation: re-enabling resurrects stored sessions, a
// missing row answers in the same shape, and the read backstops the
// disable path's best-effort marker write.
func TestAuthenticateSessionUserGate(t *testing.T) {
	rp := newTokenStore()
	token := mintSession(t, rp, &Session{UserID: "u1", TenantID: "t1", Account: "a@x.co", IssuedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour)})
	row := seedRow(t, "a@x.co", "right")
	uf := &fakeUserRP{row: row}
	az := &fakeAuthzRP{}
	uc := newAuthUCTenant(uf, az, rp, &fakeTenantRP{row: &tenant.Tenant{Code: "t1", Enabled: true}})

	row.Enabled = false
	if _, err := uc.AuthenticateSession(context.Background(), token); !errorspb.IsGeneralErrorUnauthenticated(err) {
		t.Fatalf("disabled identity: want unauthenticated, got %v", err)
	}
	if _, ok := rp.vals[sessionPrefix+token]; !ok {
		t.Fatal("disabled identity: the gate must not revoke — the key stays for resurrection")
	}

	row.Enabled = true
	az.stamp++ // every gate flip moves the version the payload compares against
	if _, err := uc.AuthenticateSession(context.Background(), token); err != nil {
		t.Fatalf("re-enabled identity: session must resurrect, got %v", err)
	}

	uf.row = nil
	az.stamp++ // a deleted row bumps on its way out
	if _, err := uc.AuthenticateSession(context.Background(), token); !errorspb.IsGeneralErrorUnauthenticated(err) {
		t.Fatalf("missing identity row: want unauthenticated same shape, got %v", err)
	}

	uf.findIDErr = errorspb.ErrorInfraErrorNetworkConnection("")
	az.stamp++ // an unrelated bump forces the deep read the error rides in on
	if _, err := uc.AuthenticateSession(context.Background(), token); !errorspb.IsInfraErrorNetworkConnection(err) {
		t.Fatalf("generic identity error: want passthrough, got %v", err)
	}
}

// Session verification -----------------------------------------------------------------------------------------------

// A session inside its lifetime but below the refresh watermark slides its TTL.
func TestAuthenticateSessionSlidesLowTTL(t *testing.T) {
	trp := newTokenStore()
	uc := newAuthUC(&fakeUserRP{row: seedRow(t, "a@x.co", "right")}, &fakeAuthzRP{}, trp)
	token := mintSession(t, trp, &Session{
		UserID: "u1", TenantID: "t1",
		ExpiresAt: time.Now().Add(shared.SessionMax), IssuedAt: time.Now(),
	})
	trp.ttls[sessionPrefix+token] = time.Minute // below the refresh watermark

	if _, err := uc.AuthenticateSession(context.Background(), token); err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if len(trp.expire) != 1 || trp.expire[0] != sessionIdle {
		t.Fatalf("slide: expire calls = %v, want one sessionIdle", trp.expire)
	}
}

// Past the absolute cap the session dies on the spot — the key is deleted, not
// merely rejected.
func TestAuthenticateSessionAbsoluteCapKills(t *testing.T) {
	trp := newTokenStore()
	uc := newAuthUC(&fakeUserRP{}, &fakeAuthzRP{}, trp)
	token := mintSession(t, trp, &Session{
		UserID: "u1", TenantID: "t1",
		ExpiresAt: time.Now().Add(-time.Minute), IssuedAt: time.Now().Add(-time.Hour),
	})

	if _, err := uc.AuthenticateSession(context.Background(), token); !errorspb.IsGeneralErrorUnauthenticated(err) {
		t.Fatalf("absolute cap: want unauthenticated, got %v", err)
	}
	if _, ok := trp.vals[sessionPrefix+token]; ok {
		t.Fatal("absolute cap: session key must be deleted")
	}
}

// The revocation marker is a kick, not a ban: sessions issued at or before the
// marker die, later logins survive.
func TestAuthenticateSessionRevocationMarker(t *testing.T) {
	trp := newTokenStore()
	uc := newAuthUC(&fakeUserRP{row: seedRow(t, "a@x.co", "right")}, &fakeAuthzRP{}, trp)
	issued := time.Now().Add(-time.Hour)
	kicked := mintSession(t, trp, &Session{
		UserID: "u1", TenantID: "t1",
		ExpiresAt: time.Now().Add(shared.SessionMax), IssuedAt: issued,
	})
	survivor := mintSession(t, trp, &Session{
		UserID: "u1", TenantID: "t1",
		ExpiresAt: time.Now().Add(shared.SessionMax), IssuedAt: time.Now(),
	})
	marker := shared.RevokedMarkerPrefix + "u1"
	revokedAt := time.Now().Add(-30 * time.Minute) // between the two issue times
	if err := trp.Set(context.Background(), marker, utils.MustMarshal(revokedAt), shared.SessionMax); err != nil {
		t.Fatalf("marker: %v", err)
	}

	if _, err := uc.AuthenticateSession(context.Background(), kicked); !errorspb.IsGeneralErrorUnauthenticated(err) {
		t.Fatalf("kicked session: want unauthenticated, got %v", err)
	}
	if _, ok := trp.vals[sessionPrefix+kicked]; ok {
		t.Fatal("kicked session: key must be deleted")
	}
	if _, err := uc.AuthenticateSession(context.Background(), survivor); err != nil {
		t.Fatalf("survivor session: %v", err)
	}
}

// Edge inputs: an unrecognized token is a plain 401; a corrupt payload never
// authenticates.
func TestAuthenticateSessionEdgeCacheMissAndDirtyPayload(t *testing.T) {
	trp := newTokenStore()
	uc := newAuthUC(&fakeUserRP{}, &fakeAuthzRP{}, trp)

	if _, err := uc.AuthenticateSession(context.Background(), "no-such-token"); !errorspb.IsGeneralErrorUnauthenticated(err) {
		t.Fatalf("cache miss: want unauthenticated, got %v", err)
	}
	dirty := mintSession(t, trp, &Session{UserID: "u1", TenantID: "t1"})
	trp.vals[sessionPrefix+dirty] = []byte("{not json")
	if _, err := uc.AuthenticateSession(context.Background(), dirty); err == nil {
		t.Fatal("dirty payload: must not authenticate")
	}
}

// A store failure mid-authentication surfaces as itself — infra never
// masquerades as an unauthenticated.
func TestAuthenticateSessionExpireFailureSurfaces(t *testing.T) {
	trp := newTokenStore()
	trp.expireErr = errorspb.ErrorInfraErrorNetworkConnection("")
	uc := newAuthUC(&fakeUserRP{row: seedRow(t, "a@x.co", "right")}, &fakeAuthzRP{}, trp)
	token := mintSession(t, trp, &Session{
		UserID: "u1", TenantID: "t1",
		ExpiresAt: time.Now().Add(shared.SessionMax), IssuedAt: time.Now(),
	})

	if _, err := uc.AuthenticateSession(context.Background(), token); !errorspb.IsInfraErrorNetworkConnection(err) {
		t.Fatalf("expire failure: want infra passthrough, got %v", err)
	}
}

// A corrupt revocation marker is a store integrity problem, not a kick — it
// surfaces instead of silently treating every session as revoked.
func TestAuthenticateSessionDirtyMarkerSurfaces(t *testing.T) {
	trp := newTokenStore()
	uc := newAuthUC(&fakeUserRP{}, &fakeAuthzRP{}, trp)
	token := mintSession(t, trp, &Session{
		UserID: "u1", TenantID: "t1",
		ExpiresAt: time.Now().Add(shared.SessionMax), IssuedAt: time.Now(),
	})
	if err := trp.Set(context.Background(), shared.RevokedMarkerPrefix+"u1", []byte("{dirty"), shared.SessionMax); err != nil {
		t.Fatalf("marker: %v", err)
	}

	if _, err := uc.AuthenticateSession(context.Background(), token); err == nil || errorspb.IsGeneralErrorUnauthenticated(err) {
		t.Fatalf("dirty marker: want a surfaced error, got %v", err)
	}
}

// The verified subject carries the seat's population downstream.
func TestAuthenticateSessionCarriesPopulation(t *testing.T) {
	rp := newTokenStore()
	token := mintSession(t, rp, &Session{UserID: "u1", TenantID: "t1", Population: "user", Account: "a@x.co", IssuedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour)})
	uc := newAuthUCTenant(&fakeUserRP{row: seedRow(t, "a@x.co", "right")}, &fakeAuthzRP{}, rp,
		&fakeTenantRP{row: &tenant.Tenant{Code: "t1", Enabled: true}})

	sub, err := uc.AuthenticateSession(context.Background(), token)
	if err != nil || sub.Population != "user" {
		t.Fatalf("verify: sub=%+v err=%v", sub, err)
	}
}

// An unchanged liveness stamp means no tenant or user gate has moved since
// the last deep read: the two row reads collapse into the single stamp
// comparison, and the deep-read ports stay untouched.
func TestAuthenticateSessionFastPathSkipsDeepReads(t *testing.T) {
	rp := newTokenStore()
	token := mintSession(t, rp, &Session{UserID: "u1", TenantID: "t1", Account: "a@x.co", IssuedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour), LivenessStamp: 7})
	uf := &fakeUserRP{row: seedRow(t, "a@x.co", "right")}
	tn := &fakeTenantRP{row: &tenant.Tenant{Code: "t1", Enabled: true}}
	uc := newAuthUCTenant(uf, &fakeAuthzRP{stamp: 7}, rp, tn)

	for range 2 {
		if _, err := uc.AuthenticateSession(context.Background(), token); err != nil {
			t.Fatalf("fast verify: %v", err)
		}
	}
	if tn.findByCodeCall != 0 || uf.findByIDCalls != 0 {
		t.Fatalf("stamp match must skip the deep reads: tenant=%d user=%d", tn.findByCodeCall, uf.findByIDCalls)
	}
}

// A moved stamp forces the deep reads once; the survivor carries the current
// version back into the payload, so the next request rides the fast path.
func TestAuthenticateSessionStampMoveDeepReadsAndWritesBack(t *testing.T) {
	rp := newTokenStore()
	token := mintSession(t, rp, &Session{UserID: "u1", TenantID: "t1", Account: "a@x.co", IssuedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour), LivenessStamp: 8})
	uf := &fakeUserRP{row: seedRow(t, "a@x.co", "right")}
	tn := &fakeTenantRP{row: &tenant.Tenant{Code: "t1", Enabled: true}}
	uc := newAuthUCTenant(uf, &fakeAuthzRP{stamp: 9}, rp, tn)

	if _, err := uc.AuthenticateSession(context.Background(), token); err != nil {
		t.Fatalf("deep verify: %v", err)
	}
	if tn.findByCodeCall != 1 || uf.findByIDCalls != 1 {
		t.Fatalf("stamp mismatch must deep-read once: tenant=%d user=%d", tn.findByCodeCall, uf.findByIDCalls)
	}
	stored, err := utils.Unmarshal[Session](rp.vals[sessionPrefix+token])
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if stored.LivenessStamp != 9 {
		t.Fatalf("payload must carry the current stamp after a deep read: got %d", stored.LivenessStamp)
	}
	if _, err := uc.AuthenticateSession(context.Background(), token); err != nil {
		t.Fatalf("second verify: %v", err)
	}
	if tn.findByCodeCall != 1 || uf.findByIDCalls != 1 {
		t.Fatalf("write-back must put the next request on the fast path: tenant=%d user=%d", tn.findByCodeCall, uf.findByIDCalls)
	}
}

// A failed stamp read degrades the optimization, never the guarantee: the
// deep reads answer as they did before the channel existed, and the payload
// keeps its old stamp.
func TestAuthenticateSessionStampFailureFallsBackToDeepReads(t *testing.T) {
	rp := newTokenStore()
	token := mintSession(t, rp, &Session{UserID: "u1", TenantID: "t1", Account: "a@x.co", IssuedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour), LivenessStamp: 5})
	uf := &fakeUserRP{row: seedRow(t, "a@x.co", "right")}
	tn := &fakeTenantRP{row: &tenant.Tenant{Code: "t1", Enabled: true}}
	uc := newAuthUCTenant(uf, &fakeAuthzRP{stampErr: errorspb.ErrorInfraErrorNetworkConnection("")}, rp, tn)

	if _, err := uc.AuthenticateSession(context.Background(), token); err != nil {
		t.Fatalf("fallback verify: %v", err)
	}
	if tn.findByCodeCall != 1 || uf.findByIDCalls != 1 {
		t.Fatalf("channel failure must deep-read: tenant=%d user=%d", tn.findByCodeCall, uf.findByIDCalls)
	}
	stored, err := utils.Unmarshal[Session](rp.vals[sessionPrefix+token])
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if stored.LivenessStamp != 5 {
		t.Fatalf("a failed channel must not rewrite the payload: got %d", stored.LivenessStamp)
	}
}

// The mint writes the stamp it saw, so a fresh session rides the fast path
// from its first verification.
func TestLoginMintsCurrentLivenessStamp(t *testing.T) {
	rp := newTokenStore()
	uc := newAuthUC(&fakeUserRP{row: seedRow(t, "a@x.co", "right")}, &fakeAuthzRP{stamp: 3}, rp)

	token, _, _, err := uc.Login(context.Background(), "a@x.co", "right", "")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	stored, err := utils.Unmarshal[Session](rp.vals[sessionPrefix+token])
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if stored.LivenessStamp != 3 {
		t.Fatalf("minted payload must carry the current stamp: got %d", stored.LivenessStamp)
	}
}

// Logout and me -------------------------------------------------------------------------------------------------------

// The token is the session's only key, so it doubles as the deletion handle —
// logout removes exactly that key; a callerless logout is unauthenticated.
func TestLogoutDeletesSessionKey(t *testing.T) {
	trp := newTokenStore()
	uc := newAuthUC(&fakeUserRP{}, &fakeAuthzRP{}, trp)
	subject := &security.Subject{UserID: "u1", SessionID: "tok-1"}
	trp.vals[sessionPrefix+"tok-1"] = []byte("{}")

	if err := uc.Logout(context.Background()); !errorspb.IsGeneralErrorUnauthenticated(err) {
		t.Fatalf("callerless logout: want unauthenticated, got %v", err)
	}
	if err := uc.Logout(authCtx(subject)); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if _, ok := trp.vals[sessionPrefix+"tok-1"]; ok {
		t.Fatal("logout: session key must be deleted")
	}
}

// Engine failure on the me face is infra, not an empty permission set — a
// session whose grants cannot resolve must never read as "no permissions".
func TestGetCurrentUserEngineFailureIsInternal(t *testing.T) {
	uc := newAuthUC(&fakeUserRP{}, &fakeAuthzRP{err: errorspb.ErrorInfraErrorNetworkConnection("")}, newTokenStore())
	subject := &security.Subject{UserID: "u1", SessionID: "tok-1", Extra: map[string]any{shared.ExtraAccount: "a@x.co"}}

	if _, err := uc.GetCurrentUser(authCtx(subject)); !errorspb.IsGeneralErrorInternal(err) {
		t.Fatalf("engine failure: want INTERNAL, got %v", err)
	}
}

// The me face hydrates from the session's extra, the engine's patterns, and
// the row's avatar; a missing avatar row degrades to nil, never an error.
func TestGetCurrentUserHydrates(t *testing.T) {
	row := seedRow(t, "a@x.co", "right")
	avatar := "f1"
	row.Avatar = &avatar
	uc := newAuthUC(&fakeUserRP{row: row}, &fakeAuthzRP{grants: []authz.GrantView{{Pattern: "/svc/*"}}}, newTokenStore())
	subject := &security.Subject{UserID: "u1", SessionID: "tok-1", Extra: map[string]any{shared.ExtraAccount: "a@x.co"}}

	out, err := uc.GetCurrentUser(authCtx(subject))
	if err != nil {
		t.Fatalf("me: %v", err)
	}
	if out.Account != "a@x.co" || len(out.Permissions) != 1 || out.Permissions[0] != "/svc/*" || out.Avatar == nil || *out.Avatar != "f1" {
		t.Fatalf("me: out = %+v", out)
	}

	// A missing avatar row degrades to nil.
	uc2 := newAuthUC(&fakeUserRP{}, &fakeAuthzRP{grants: []authz.GrantView{{Pattern: "/svc/*"}}}, newTokenStore())
	out2, err := uc2.GetCurrentUser(authCtx(subject))
	if err != nil {
		t.Fatalf("me degraded: %v", err)
	}
	if out2.Avatar != nil {
		t.Fatalf("me degraded: avatar = %v, want nil", *out2.Avatar)
	}
}

// The me face carries the caller's seats — tenancy lives on seats, never
// on the identity row.
func TestGetCurrentUserCarriesSeats(t *testing.T) {
	subject := &security.Subject{UserID: "u1", TenantID: "t1", SessionID: "s1"}
	uf := &fakeUserRP{seats: []*user.Seat{
		{UserID: "u1", TenantID: "default", Population: "user"},
	}}
	uc := newAuthUCTenant(uf, &fakeAuthzRP{grants: []authz.GrantView{{Pattern: "/svc/*"}}}, newTokenStore(),
		&fakeTenantRP{row: &tenant.Tenant{Code: "t1", Enabled: true}})

	out, err := uc.GetCurrentUser(authCtx(subject))
	if err != nil {
		t.Fatalf("me: %v", err)
	}
	if len(out.Seats) != 1 || out.Seats[0].TenantID != "default" || out.Seats[0].Population != "user" {
		t.Fatalf("me: seats = %+v", out.Seats)
	}

	// A seat-read failure surfaces — a broken store must not masquerade as
	// an empty seat list.
	uf.seatsErr = errorspb.ErrorInfraErrorNetworkConnection("")
	if _, err := uc.GetCurrentUser(authCtx(subject)); !errorspb.IsInfraErrorNetworkConnection(err) {
		t.Fatalf("me broken store: want passthrough, got %v", err)
	}
}

// Identity login ------------------------------------------------------------------------------------------------------

// The exchange resolves the identity, the binding resolves the row, and the
// session mints from the seat the provider's landing picks — here the
// undeclared landing of the default seat, the staff-binding shape.
func TestLoginByIdentityMintsFromSeat(t *testing.T) {
	trp := newTokenStore()
	uf := boundUser(t, []*user.Seat{{UserID: "u9", TenantID: "t1", Population: security.DefaultPopulation}})
	uc := newAuthUCConf(fakeProviderConf(), uf, &fakeAuthzRP{}, trp,
		&fakeTenantRP{row: &tenant.Tenant{Code: "t1", Enabled: true}})
	state := issueState(t, trp, "fake")

	token, _, _, _, err := uc.LoginByIdentity(context.Background(), "fake", state, identityReq("fake", state))
	if err != nil {
		t.Fatalf("login-by-identity: %v", err)
	}
	sess, err := utils.Unmarshal[Session](trp.vals[sessionPrefix+token])
	if err != nil {
		t.Fatal(err)
	}
	if sess.UserID != "u9" || sess.TenantID != "t1" || sess.Population != security.DefaultPopulation {
		t.Fatalf("mint: session = %+v, want the seat's t1/%s for u9", sess, security.DefaultPopulation)
	}
}

// The person behind the identity rides the login's return verbatim — the
// wizard's nickname prefill source — without touching the session mint.
func TestLoginByIdentityCarriesProfile(t *testing.T) {
	trp := newTokenStore()
	uf := boundUser(t, []*user.Seat{{UserID: "u9", TenantID: "t1", Population: security.DefaultPopulation}})
	uc := newAuthUCConf(fakeProviderConf(), uf, &fakeAuthzRP{}, trp,
		&fakeTenantRP{row: &tenant.Tenant{Code: "t1", Enabled: true}})
	state := issueState(t, trp, "fake")

	_, _, _, profile, err := uc.LoginByIdentity(context.Background(), "fake", state, identityReq("fake", state))
	if err != nil {
		t.Fatalf("login-by-identity: %v", err)
	}
	if profile == nil || profile.Nickname != "昵称" || profile.AvatarURL != "https://qlogo.cn/a.png" {
		t.Fatalf("profile = %+v, want the exchange's person verbatim", profile)
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
	if _, _, _, _, err := uc.LoginByIdentity(context.Background(), "fake", "st-missing", identityReq("fake", "st-missing")); !systempb.IsSystemIdentityExchangeFailed(err) {
		t.Fatalf("missing state: want EXCHANGE_FAILED, got %v", err)
	}
	if fakeExchanges != before {
		t.Fatal("missing state: the exchange ran on a refused cash")
	}

	other := issueState(t, trp, "failing")
	before = fakeExchanges
	if _, _, _, _, err := uc.LoginByIdentity(context.Background(), "fake", other, identityReq("fake", other)); !systempb.IsSystemIdentityExchangeFailed(err) {
		t.Fatalf("foreign state: want EXCHANGE_FAILED, got %v", err)
	}
	if fakeExchanges != before {
		t.Fatal("foreign state: the exchange ran on a refused cash")
	}

	state := issueState(t, trp, "fake")
	if _, _, _, _, err := uc.LoginByIdentity(context.Background(), "fake", state, identityReq("fake", state)); err != nil {
		t.Fatalf("first cash: %v", err)
	}
	if _, _, _, _, err := uc.LoginByIdentity(context.Background(), "fake", state, identityReq("fake", state)); !systempb.IsSystemIdentityExchangeFailed(err) {
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
	got, err := trp.Get(context.Background(), identityStatePrefix+state)
	if err != nil {
		t.Fatalf("begin: stored state err=%v", err)
	}
	if val, uerr := utils.Unmarshal[identityStateValue](got); uerr != nil || val.Provider != "fake" || val.BindUser != "" {
		t.Fatalf("begin: stored state = %q (%v), want an anonymous fake-provider mint", got, uerr)
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
		if _, _, _, _, err := uc.LoginByIdentity(context.Background(), key, "st-x", identityReq(key, "st-x")); !systempb.IsSystemIdentityProviderUnknown(err) {
			t.Fatalf("provider %q: want PROVIDER_UNKNOWN, got %v", key, err)
		}
	}
}

// A completed exchange with no binding reads as a failed exchange — the
// face never confirms which external identities exist.
func TestLoginByIdentityUnboundFailsAsExchange(t *testing.T) {
	uc, trp := identityUC(t, &fakeUserRP{})
	state := issueState(t, trp, "fake")

	if _, _, _, _, err := uc.LoginByIdentity(context.Background(), "fake", state, identityReq("fake", state)); !systempb.IsSystemIdentityExchangeFailed(err) {
		t.Fatalf("unbound: want EXCHANGE_FAILED, got %v", err)
	}
}

// The auto-register gate births the member the C-end product promises: an
// unbound exchange on a gated row creates the generated-account user — no
// password, so the provider is the only way in — lands the declared seat,
// binds the credential, and mints from that seat, the same derivation and
// minter as every other login face.
func TestLoginByIdentityAutoRegisterBirthsMember(t *testing.T) {
	uf := &fakeUserRP{}
	uc, trp := autoUC(t, uf)
	state := issueState(t, trp, "auto")

	token, _, _, _, err := uc.LoginByIdentity(context.Background(), "auto", state, identityReq("auto", state))
	if err != nil {
		t.Fatalf("auto-register: %v", err)
	}
	if len(uf.createdUsers) != 1 {
		t.Fatalf("auto-register: %d users created, want the one birth", len(uf.createdUsers))
	}
	born := uf.createdUsers[0]
	if born.Account == nil || *born.Account == "" {
		t.Fatalf("auto-register: account = %v, want a generated handle", born.Account)
	}
	if born.PasswordHash == nil || *born.PasswordHash != "" {
		t.Fatalf("auto-register: hash = %v, want empty for provider-only auth", born.PasswordHash)
	}
	if !born.Enabled {
		t.Fatal("auto-register: the birth lands disabled")
	}
	if len(uf.seats) != 1 || uf.seats[0].UserID != born.ID ||
		uf.seats[0].TenantID != security.DefaultTenant || uf.seats[0].Population != "member" {
		t.Fatalf("auto-register: seats = %+v, want the declared seat for %s", uf.seats, born.ID)
	}
	if len(uf.createdCreds) != 1 || uf.createdCreds[0].UserID != born.ID ||
		uf.createdCreds[0].Kind != fakeKind || uf.createdCreds[0].Identifier != "oid-x" {
		t.Fatalf("auto-register: credentials = %+v, want the identity bound to the birth", uf.createdCreds)
	}
	sess, err := utils.Unmarshal[Session](trp.vals[sessionPrefix+token])
	if err != nil {
		t.Fatal(err)
	}
	if sess.UserID != born.ID || sess.Population != "member" {
		t.Fatalf("auto-register: session = %+v, want the born member", sess)
	}
}

// A deleted binding severs the tie: a later login of the same external
// identity births a fresh user instead of resurrecting the unclaimable row —
// generated accounts are not derivable, so nothing correlates the two.
func TestLoginByIdentityUnboundBirthsFresh(t *testing.T) {
	row := seedRow(t, "leftover", "pw")
	row.ID = "u-syn"
	uf := &fakeUserRP{row: row}
	uc, trp := autoUC(t, uf)
	state := issueState(t, trp, "auto")

	token, _, _, _, err := uc.LoginByIdentity(context.Background(), "auto", state, identityReq("auto", state))
	if err != nil {
		t.Fatalf("unbound re-login: %v", err)
	}
	if len(uf.createdUsers) != 1 {
		t.Fatalf("unbound re-login: %d users created, want the fresh birth", len(uf.createdUsers))
	}
	born := uf.createdUsers[0]
	if born.ID != "u-born" || born.ID == row.ID {
		t.Fatalf("unbound re-login: born = %s, want a row distinct from the severed one", born.ID)
	}
	if len(uf.createdCreds) != 1 || uf.createdCreds[0].UserID != born.ID {
		t.Fatalf("unbound re-login: credentials = %+v, want the identity on the birth", uf.createdCreds)
	}
	sess, err := utils.Unmarshal[Session](trp.vals[sessionPrefix+token])
	if err != nil {
		t.Fatal(err)
	}
	if sess.UserID != born.ID || sess.Population != "member" {
		t.Fatalf("unbound re-login: session = %+v, want the fresh birth on the declared seat", sess)
	}
}

// The provider's own rejection surfaces in its shape.
func TestLoginByIdentityExchangeFailureSurfaces(t *testing.T) {
	uc, trp := identityUC(t, &fakeUserRP{})
	state := issueState(t, trp, "failing")

	if _, _, _, _, err := uc.LoginByIdentity(context.Background(), "failing", state, identityReq("failing", state)); !systempb.IsSystemIdentityExchangeFailed(err) {
		t.Fatalf("exchange failure: want EXCHANGE_FAILED, got %v", err)
	}
}

// The provider's declared landing directs the pick among an identity's
// seats: the member seat serves the member-declaring provider while the
// engine seat stays invisible to it — a held seat of another population is
// another face's key, not ambiguity.
func TestLoginByIdentityPicksProviderSeat(t *testing.T) {
	uf := boundUser(t, []*user.Seat{
		{UserID: "u9", TenantID: "t1", Population: "user"},
		{UserID: "u9", TenantID: "t1", Population: "member"},
	})
	trp := newTokenStore()
	uc := newAuthUCConf(&conf.Auth{Providers: []*conf.Auth_Provider{
		{Key: "auto", Kind: fakeKind, Enabled: true, Population: "member"},
	}}, uf, &fakeAuthzRP{}, trp, &fakeTenantRP{row: &tenant.Tenant{Code: "t1", Enabled: true}})
	state := issueState(t, trp, "auto")

	token, _, _, _, err := uc.LoginByIdentity(context.Background(), "auto", state, identityReq("auto", state))
	if err != nil {
		t.Fatalf("directed seat: %v", err)
	}
	sess, err := utils.Unmarshal[Session](trp.vals[sessionPrefix+token])
	if err != nil {
		t.Fatal(err)
	}
	if sess.Population != "member" || len(uf.seats) != 2 {
		t.Fatalf("directed seat: session = %+v seats = %+v, want the member seat without an ensure", sess, uf.seats)
	}
}

// A bound identity without a seat in the provider's population gains it on
// entry — entering the face is the provider's doing, never a prior
// administrative act. The same identity keeps its other seats untouched.
func TestLoginByIdentityEnsuresSeatForBoundIdentity(t *testing.T) {
	uf := boundUser(t, []*user.Seat{{UserID: "u9", TenantID: "t1", Population: "user"}})
	trp := newTokenStore()
	uc := newAuthUCConf(&conf.Auth{Providers: []*conf.Auth_Provider{
		{Key: "auto", Kind: fakeKind, Enabled: true, Population: "member"},
	}}, uf, &fakeAuthzRP{}, trp, &fakeTenantRP{row: &tenant.Tenant{Code: "t1", Enabled: true}})
	state := issueState(t, trp, "auto")

	token, _, _, _, err := uc.LoginByIdentity(context.Background(), "auto", state, identityReq("auto", state))
	if err != nil {
		t.Fatalf("bound entry: %v", err)
	}
	if len(uf.seats) != 2 || uf.seats[1].Population != "member" || uf.seats[1].TenantID != security.DefaultTenant {
		t.Fatalf("bound entry: seats = %+v, want the member seat ensured on the default tenant", uf.seats)
	}
	sess, err := utils.Unmarshal[Session](trp.vals[sessionPrefix+token])
	if err != nil {
		t.Fatal(err)
	}
	if sess.UserID != "u9" || sess.Population != "member" {
		t.Fatalf("bound entry: session = %+v, want u9 on the member seat", sess)
	}
}

// A declared landing that does not resolve is a config fact: the login names
// it precisely instead of seating into a silent default — and the refusal
// lands before the exchange, so a code is never spent on a broken provider.
func TestLoginByIdentityUnknownPopulation(t *testing.T) {
	uf := &fakeUserRP{}
	trp := newTokenStore()
	uc := newAuthUCConf(&conf.Auth{Providers: []*conf.Auth_Provider{
		{Key: "auto", Kind: fakeKind, Enabled: true, AutoRegister: true, Population: "ghost-pop"},
	}}, uf, &fakeAuthzRP{}, trp, &fakeTenantRP{row: &tenant.Tenant{Code: "t1", Enabled: true}})
	state := issueState(t, trp, "auto")

	before := fakeExchanges
	if _, _, _, _, err := uc.LoginByIdentity(context.Background(), "auto", state, identityReq("auto", state)); !systempb.IsSystemPopulationNotFound(err) {
		t.Fatalf("unknown population: want POPULATION_NOT_FOUND, got %v", err)
	}
	if fakeExchanges != before || len(uf.createdCreds) != 0 {
		t.Fatal("unknown population: the exchange ran or a credential landed")
	}
}

// Binding -------------------------------------------------------------------------------------------------------------

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
	if _, err := uc.BindCredential(ctx, "fake", "st-missing", newReq("st-missing")); !systempb.IsSystemIdentityExchangeFailed(err) {
		t.Fatalf("stateless bind: want EXCHANGE_FAILED, got %v", err)
	}
	if fakeExchanges != before || len(uf.createdCreds) != 0 {
		t.Fatal("stateless bind: the exchange ran or a credential landed")
	}

	// The session-tie law: an anonymous mint (the public login rung's
	// shape) and a foreign session's mint both answer as a failed
	// exchange, before any provider call.
	anon := issueState(t, trp, "fake")
	before = fakeExchanges
	if _, err := uc.BindCredential(ctx, "fake", anon, newReq(anon)); !systempb.IsSystemIdentityExchangeFailed(err) {
		t.Fatalf("anonymous state: want EXCHANGE_FAILED, got %v", err)
	}
	if fakeExchanges != before || len(uf.createdCreds) != 0 {
		t.Fatal("anonymous state: the exchange ran or a credential landed")
	}
	foreign := issueBoundState(t, trp, "fake", "u9")
	before = fakeExchanges
	if _, err := uc.BindCredential(ctx, "fake", foreign, newReq(foreign)); !systempb.IsSystemIdentityExchangeFailed(err) {
		t.Fatalf("foreign state: want EXCHANGE_FAILED, got %v", err)
	}
	if fakeExchanges != before || len(uf.createdCreds) != 0 {
		t.Fatal("foreign state: the exchange ran or a credential landed")
	}

	state := issueBoundState(t, trp, "fake", "u1")
	if _, err := uc.BindCredential(ctx, "fake", state, newReq(state)); err != nil {
		t.Fatalf("bind: %v", err)
	}
	if len(uf.createdCreds) != 1 || uf.createdCreds[0].UserID != "u1" {
		t.Fatalf("bind: created = %+v, want one binding to u1", uf.createdCreds)
	}

	bind := func(uf *fakeUserRP) error {
		uc, trp := identityUC(t, uf)
		state := issueBoundState(t, trp, "fake", "u1")
		_, err := uc.BindCredential(ctx, "fake", state, newReq(state))
		return err
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
	noState := issueBoundState(t, noTrp, "fake", "u1")
	if _, err := noUC.BindCredential(context.Background(), "fake", noState, newReq(noState)); !errorspb.IsGeneralErrorUnauthenticated(err) {
		t.Fatalf("no session: want 401, got %v", err)
	}
}

// The begin twin of the bind flow: an authenticated rung mints a state
// carrying the session, and the mint is refused without one.
func TestBeginBindCredentialBindsMint(t *testing.T) {
	uc, trp := identityUC(t, &fakeUserRP{})

	ctx := security.WithSubject(context.Background(), &security.Subject{UserID: "u1", TenantID: "t1"})
	url, err := uc.BeginBindCredential(ctx, "fake", "https://app.example/cb")
	if err != nil {
		t.Fatalf("begin bind: %v", err)
	}
	state := strings.TrimPrefix(url, "https://idp.example/auth/fake?state=")
	got, gerr := trp.Get(context.Background(), identityStatePrefix+state)
	if gerr != nil {
		t.Fatalf("begin bind: stored state err=%v", gerr)
	}
	if val, uerr := utils.Unmarshal[identityStateValue](got); uerr != nil || val.Provider != "fake" || val.BindUser != "u1" {
		t.Fatalf("begin bind: stored state = %q (%v), want a u1-bound fake mint", got, uerr)
	}

	if _, err := uc.BeginBindCredential(context.Background(), "fake", "https://app.example/cb"); !errorspb.IsGeneralErrorUnauthenticated(err) {
		t.Fatalf("begin bind without session: want 401, got %v", err)
	}
}

// Merge laws: proof + session-bound state, the poor side must be
// system-side weightless, the absorption is one write set (move +
// disable + ledger) with the disable's follow-ups, and every refusal
// answers with its own reason before any write lands.
func TestMergeAccountsLaws(t *testing.T) {
	code, key := "c", "fake"
	mergeReq := func(state string) *systempb.MergeAccountsRequest {
		return &systempb.MergeAccountsRequest{Provider: &key, State: &state,
			Payload: &systempb.MergeAccountsRequest_Wechat{Wechat: &systempb.WeChatIdentityPayload{Code: &code}}}
	}
	ctx := security.WithSubject(context.Background(), &security.Subject{UserID: "u1", TenantID: "t1"})
	memberPlane := &fakeAuthzRP{planes: map[string]string{"member": security.PlaneMember}}

	poorRow := func(mutate func(*user.User)) *user.User {
		empty := ""
		row := &user.User{ID: "u9", Account: &empty, PasswordHash: &empty, Enabled: true}
		if mutate != nil {
			mutate(row)
		}
		return row
	}
	merge := func(uf *fakeUserRP, trp *fakeTokenRP, state string) error {
		uc := newAuthUCConf(fakeProviderConf(), uf, memberPlane, trp,
			&fakeTenantRP{row: &tenant.Tenant{Code: "t1", Enabled: true}})
		return uc.MergeAccounts(ctx, "fake", state, mergeReq(state))
	}
	mergeFresh := func(uf *fakeUserRP) error {
		trp := newTokenStore()
		return merge(uf, trp, issueBoundState(t, trp, "fake", "u1"))
	}

	// Happy path: one binding owned by a weightless u9, one member seat.
	uf := &fakeUserRP{
		row:   poorRow(nil),
		seats: []*user.Seat{{UserID: "u9", TenantID: "t1", Population: "member"}},
		credentials: map[string]*user.Credential{
			fakeKind + "|oid-x": {Kind: fakeKind, Identifier: "oid-x", UserID: "u9"},
		},
	}
	if err := mergeFresh(uf); err != nil {
		t.Fatalf("merge: %v", err)
	}
	if len(uf.moved) != 1 || uf.credentials[fakeKind+"|oid-x"].UserID != "u1" {
		t.Fatalf("merge: moved = %v, credential owner = %+v", uf.moved, uf.credentials[fakeKind+"|oid-x"])
	}
	if len(uf.disabled) != 1 || uf.disabled[0] != "u9" {
		t.Fatalf("merge: disabled = %v, want u9", uf.disabled)
	}
	if len(uf.merges) != 1 || uf.merges[0].PoorID != "u9" || uf.merges[0].RichID != "u1" ||
		uf.merges[0].Kind != fakeKind || uf.merges[0].Identifier != "oid-x" {
		t.Fatalf("merge: ledger = %+v, want one u9→u1 row for the proven credential", uf.merges)
	}
	if len(uf.revoked) != 1 || uf.revoked[0] != "u9" || uf.notified != 1 {
		t.Fatalf("merge: revoked = %v notified = %d, want u9's sessions gone and one notify", uf.revoked, uf.notified)
	}

	// A free credential and the caller's own credential refuse before writes.
	free := &fakeUserRP{}
	if err := mergeFresh(free); !systempb.IsSystemMergeNotBound(err) {
		t.Fatalf("unbound credential: want NOT_BOUND, got %v", err)
	}
	own := &fakeUserRP{credentials: map[string]*user.Credential{
		fakeKind + "|oid-x": {Kind: fakeKind, Identifier: "oid-x", UserID: "u1"},
	}}
	if err := mergeFresh(own); !systempb.IsSystemIdentityAlreadyBound(err) {
		t.Fatalf("own credential: want ALREADY_BOUND, got %v", err)
	}

	// Weight refuses: a password, a second credential, an operator seat,
	// an already-dead account.
	pw := "set"
	withPw := &fakeUserRP{
		row: poorRow(func(r *user.User) { r.PasswordHash = &pw }),
		credentials: map[string]*user.Credential{
			fakeKind + "|oid-x": {Kind: fakeKind, Identifier: "oid-x", UserID: "u9"},
		},
	}
	if err := mergeFresh(withPw); !systempb.IsSystemMergeNotMergeable(err) {
		t.Fatalf("password weight: want NOT_MERGEABLE, got %v", err)
	}
	multi := &fakeUserRP{
		row: poorRow(nil),
		credentials: map[string]*user.Credential{
			fakeKind + "|oid-x": {Kind: fakeKind, Identifier: "oid-x", UserID: "u9"},
			fakeKind + "|oid-y": {Kind: fakeKind, Identifier: "oid-y", UserID: "u9"},
		},
	}
	if err := mergeFresh(multi); !systempb.IsSystemMergeNotMergeable(err) {
		t.Fatalf("multi credential: want NOT_MERGEABLE, got %v", err)
	}
	operator := &fakeUserRP{
		row:   poorRow(nil),
		seats: []*user.Seat{{UserID: "u9", TenantID: "t1", Population: "ops"}},
		credentials: map[string]*user.Credential{
			fakeKind + "|oid-x": {Kind: fakeKind, Identifier: "oid-x", UserID: "u9"},
		},
	}
	if err := mergeFresh(operator); !systempb.IsSystemMergeNotMergeable(err) {
		t.Fatalf("operator seat: want NOT_MERGEABLE, got %v", err)
	}
	dead := &fakeUserRP{
		row: poorRow(func(r *user.User) { r.Enabled = false }),
		credentials: map[string]*user.Credential{
			fakeKind + "|oid-x": {Kind: fakeKind, Identifier: "oid-x", UserID: "u9"},
		},
	}
	if err := mergeFresh(dead); !systempb.IsSystemMergeNotMergeable(err) {
		t.Fatalf("dead account: want NOT_MERGEABLE, got %v", err)
	}

	// The one-shot ledger's race backstop and the vanished credential.
	raced := &fakeUserRP{
		row:      poorRow(nil),
		seats:    []*user.Seat{{UserID: "u9", TenantID: "t1", Population: "member"}},
		mergeErr: errorspb.ErrorInfraErrorDbConstraint(""),
		credentials: map[string]*user.Credential{
			fakeKind + "|oid-x": {Kind: fakeKind, Identifier: "oid-x", UserID: "u9"},
		},
	}
	if err := mergeFresh(raced); !systempb.IsSystemMergeNotMergeable(err) {
		t.Fatalf("ledger race: want NOT_MERGEABLE, got %v", err)
	}
	vanished := &fakeUserRP{
		row:     poorRow(nil),
		seats:   []*user.Seat{{UserID: "u9", TenantID: "t1", Population: "member"}},
		moveErr: errorspb.ErrorInfraErrorDbNotFound(""),
		credentials: map[string]*user.Credential{
			fakeKind + "|oid-x": {Kind: fakeKind, Identifier: "oid-x", UserID: "u9"},
		},
	}
	if err := mergeFresh(vanished); !systempb.IsSystemMergeNotBound(err) {
		t.Fatalf("vanished credential: want NOT_BOUND, got %v", err)
	}

	// The session-tie law rides the merge too: an anonymous state never
	// opens an absorption.
	trp := newTokenStore()
	anon := issueState(t, trp, "fake")
	anonUf := &fakeUserRP{row: poorRow(nil), credentials: map[string]*user.Credential{
		fakeKind + "|oid-x": {Kind: fakeKind, Identifier: "oid-x", UserID: "u9"},
	}}
	if err := merge(anonUf, trp, anon); !systempb.IsSystemIdentityExchangeFailed(err) {
		t.Fatalf("anonymous state: want EXCHANGE_FAILED, got %v", err)
	}
	if len(anonUf.moved) != 0 || len(anonUf.disabled) != 0 || len(anonUf.merges) != 0 {
		t.Fatal("anonymous state: a write landed on a refused cash")
	}

	noUC := func() {
		trp := newTokenStore()
		uc := newAuthUCConf(fakeProviderConf(), &fakeUserRP{row: poorRow(nil)}, memberPlane, trp,
			&fakeTenantRP{row: &tenant.Tenant{Code: "t1", Enabled: true}})
		state := issueBoundState(t, trp, "fake", "u1")
		if err := uc.MergeAccounts(context.Background(), "fake", state, mergeReq(state)); !errorspb.IsGeneralErrorUnauthenticated(err) {
			t.Fatalf("no session: want 401, got %v", err)
		}
	}
	noUC()
}

func TestResolveCredentialLaws(t *testing.T) {
	smsKey, fakeKey := "sms", "fake"
	memberPlane := &fakeAuthzRP{planes: map[string]string{"member": security.PlaneMember}}
	tn := &fakeTenantRP{row: &tenant.Tenant{Code: "t1", Enabled: true}}
	ctx := security.WithSubject(context.Background(), &security.Subject{UserID: "u1", TenantID: "t1"})

	phoneReq := func(phone string) *systempb.ResolveCredentialRequest {
		return &systempb.ResolveCredentialRequest{Provider: &smsKey,
			Credential: &systempb.ResolveCredentialRequest_Phone{Phone: phone}}
	}
	resolve := func(uf *fakeUserRP, providerKey, state string, in *systempb.ResolveCredentialRequest) (*CredentialOwner, error) {
		uc := newAuthUCConf(fakeProviderConf(), uf, memberPlane, newTokenStore(), tn)
		return uc.ResolveCredential(ctx, providerKey, state, in)
	}

	// Free phone: the plain-bind path, no owner facts at all.
	if owner, err := resolve(&fakeUserRP{}, smsKey, "", phoneReq("13900001111")); err != nil || owner != nil {
		t.Fatalf("free phone: want nil owner, got %v %v", owner, err)
	}

	// A weightless member owner answers with the mergeable verdict and a
	// masked credential list; the +86 form canonicalizes to the same row.
	empty := ""
	owned := &fakeUserRP{
		row:   &user.User{ID: "u9", Account: &empty, PasswordHash: &empty, Enabled: true},
		seats: []*user.Seat{{UserID: "u9", TenantID: "t1", Population: "member"}},
		credentials: map[string]*user.Credential{
			smsKind + "|13900001111": {Kind: smsKind, Identifier: "13900001111", UserID: "u9"},
		},
	}
	for _, form := range []string{"13900001111", "+8613900001111"} {
		owner, err := resolve(owned, smsKey, "", phoneReq(form))
		if err != nil {
			t.Fatalf("owned phone %s: %v", form, err)
		}
		if owner == nil || owner.UserID != "u9" || !owner.Enabled || !owner.Mergeable {
			t.Fatalf("owned phone %s: owner = %+v", form, owner)
		}
		if len(owner.Credentials) != 1 || owner.Credentials[0].Identifier != "139****1111" {
			t.Fatalf("owned phone %s: credentials = %+v, want masked", form, owner.Credentials)
		}
	}

	// Weight answers as a fact, not a refusal: the offer face renders the
	// blocked pane from the same shape.
	pw := "set"
	heavy := &fakeUserRP{
		row: &user.User{ID: "u9", Account: &empty, PasswordHash: &pw, Enabled: true},
		credentials: map[string]*user.Credential{
			smsKind + "|13900001111": {Kind: smsKind, Identifier: "13900001111", UserID: "u9"},
		},
	}
	if owner, err := resolve(heavy, smsKey, "", phoneReq("13900001111")); err != nil || owner == nil || owner.Mergeable {
		t.Fatalf("heavy owner: want mergeable=false, got %v %v", owner, err)
	}

	// A non-mainland phone and a phone branch on a redirect provider are
	// both wiring refusals.
	if _, err := resolve(&fakeUserRP{}, smsKey, "", phoneReq("12345")); !errorspb.IsGeneralErrorValidationFailed(err) {
		t.Fatalf("non-mainland: want VALIDATION_FAILED, got %v", err)
	}
	if _, err := resolve(&fakeUserRP{}, fakeKey, "", &systempb.ResolveCredentialRequest{Provider: &fakeKey,
		Credential: &systempb.ResolveCredentialRequest_Phone{Phone: "13900001111"}}); !systempb.IsSystemIdentityProviderUnknown(err) {
		t.Fatalf("phone on redirect provider: want PROVIDER_UNKNOWN, got %v", err)
	}

	// The wechat branch is proof-carrying: a bind-minted state exchanges to
	// the owner, an anonymous state refuses before any lookup.
	code := "c"
	wechatReq := func(state string) *systempb.ResolveCredentialRequest {
		return &systempb.ResolveCredentialRequest{Provider: &fakeKey, State: &state,
			Credential: &systempb.ResolveCredentialRequest_Wechat{Wechat: &systempb.WeChatIdentityPayload{Code: &code}}}
	}
	bound := &fakeUserRP{
		row:   &user.User{ID: "u9", Account: &empty, PasswordHash: &empty, Enabled: true},
		seats: []*user.Seat{{UserID: "u9", TenantID: "t1", Population: "member"}},
		credentials: map[string]*user.Credential{
			fakeKind + "|oid-x": {Kind: fakeKind, Identifier: "oid-x", UserID: "u9"},
		},
	}
	trp := newTokenStore()
	state := issueBoundState(t, trp, "fake", "u1")
	owner, err := newAuthUCConf(fakeProviderConf(), bound, memberPlane, trp, tn).
		ResolveCredential(ctx, "fake", state, wechatReq(state))
	if err != nil || owner == nil || owner.UserID != "u9" || !owner.Mergeable {
		t.Fatalf("wechat branch: owner = %+v err = %v", owner, err)
	}
	anonState := issueState(t, trp, "fake")
	if _, err := newAuthUCConf(fakeProviderConf(), bound, memberPlane, trp, tn).
		ResolveCredential(ctx, "fake", anonState, wechatReq(anonState)); !systempb.IsSystemIdentityExchangeFailed(err) {
		t.Fatalf("anonymous state: want EXCHANGE_FAILED, got %v", err)
	}

	// A credential whose account vanished reads as free, with the dangling
	// row left to the log.
	dangling := &fakeUserRP{
		row: &user.User{ID: "u1", Account: &empty, PasswordHash: &empty, Enabled: true},
		credentials: map[string]*user.Credential{
			smsKind + "|13900001111": {Kind: smsKind, Identifier: "13900001111", UserID: "u9"},
		},
	}
	if owner, err := resolve(dangling, smsKey, "", phoneReq("13900001111")); err != nil || owner != nil {
		t.Fatalf("dangling credential: want free, got %v %v", owner, err)
	}

	if _, err := newAuthUCConf(fakeProviderConf(), &fakeUserRP{}, memberPlane, newTokenStore(), tn).
		ResolveCredential(context.Background(), "sms", "", phoneReq("13900001111")); !errorspb.IsGeneralErrorUnauthenticated(err) {
		t.Fatalf("no session: want 401, got %v", err)
	}
}

func TestResetPasswordLaws(t *testing.T) {
	key := "fake"
	phone := "13900001111"
	code := "123456"
	newPw := "secret1"
	req := func(state string) *systempb.ResetPasswordRequest {
		return &systempb.ResetPasswordRequest{Provider: &key, State: &state, NewPassword: &newPw,
			Payload: &systempb.ResetPasswordRequest_Sms{Sms: &systempb.SmsIdentityPayload{Phone: &phone, Code: &code}}}
	}
	reset := func(uf *fakeUserRP, trp *fakeTokenRP, state string) error {
		uc := newAuthUCConf(fakeProviderConf(), uf, &fakeAuthzRP{}, trp, &fakeTenantRP{})
		return uc.ResetPassword(context.Background(), "fake", state, "secret1", req(state))
	}

	// The proof names the account: the sms credential's owner is rehashed,
	// sessions stand (a rotation is not a revocation — no revoke call).
	empty := ""
	uf := &fakeUserRP{
		row: &user.User{ID: "u9", Account: &empty, PasswordHash: &empty, Enabled: true},
		credentials: map[string]*user.Credential{
			fakeKind + "|oid-x": {Kind: fakeKind, Identifier: "oid-x", UserID: "u9"},
		},
	}
	trp := newTokenStore()
	if err := reset(uf, trp, issueState(t, trp, "fake")); err != nil {
		t.Fatalf("reset: %v", err)
	}
	if len(uf.updates) != 1 || uf.updates[0].u.ID != "u9" || uf.updates[0].u.PasswordHash == nil ||
		!slices.Contains(uf.updates[0].fields, "password") {
		t.Fatalf("reset: updates = %+v, want one password write on u9", uf.updates)
	}
	if len(uf.revoked) != 0 {
		t.Fatalf("reset: revoked = %v, want sessions standing", uf.revoked)
	}

	// A phone no account holds and a state nobody minted both refuse.
	free := &fakeUserRP{}
	if err := reset(free, trp, issueState(t, trp, "fake")); !systempb.IsSystemUserNotFound(err) {
		t.Fatalf("no owner: want USER_NOT_FOUND, got %v", err)
	}
	if err := reset(uf, newTokenStore(), "st-unknown"); !systempb.IsSystemIdentityExchangeFailed(err) {
		t.Fatalf("unknown state: want EXCHANGE_FAILED, got %v", err)
	}

	// The composition law (Q3-①): a proof-backed but weak candidate is
	// refused after the exchange, before any digest write.
	weak := "secret"
	wstate := issueState(t, trp, "fake")
	wreq := &systempb.ResetPasswordRequest{Provider: &key, State: &wstate, NewPassword: &weak,
		Payload: &systempb.ResetPasswordRequest_Sms{Sms: &systempb.SmsIdentityPayload{Phone: &phone, Code: &code}}}
	if err := newAuthUCConf(fakeProviderConf(), uf, &fakeAuthzRP{}, trp, &fakeTenantRP{}).
		ResetPassword(context.Background(), "fake", wstate, weak, wreq); !systempb.IsSystemPasswordPolicy(err) {
		t.Fatalf("weak new password: want PASSWORD_POLICY, got %v", err)
	}
	if len(uf.updates) != 1 {
		t.Fatalf("weak new password: updates = %d, want still one (the earlier strong reset only)", len(uf.updates))
	}
}

// The pay-enrichment read's two laws: the answer is raw, not the masked
// display form (the unified order consumes the identifier itself), and a
// caller with no wechat binding answers nil — the consumer's refusal to
// act on it lives in the order service.
func TestWechatOpenIDRawAndUnbound(t *testing.T) {
	ctx := security.WithSubject(context.Background(), &security.Subject{UserID: "u1", TenantID: "t1"})
	uc := newAuthUC(&fakeUserRP{credentials: map[string]*user.Credential{
		smsKind + "|13900001111":   {Kind: smsKind, Identifier: "13900001111", UserID: "u1"},
		wechatKind + "|oid-real-1": {Kind: wechatKind, Identifier: "oid-real-1", UserID: "u1"},
	}}, &fakeAuthzRP{}, newTokenStore())

	openid, err := uc.WechatOpenID(ctx)
	if err != nil {
		t.Fatalf("wechat openid: %v", err)
	}
	if openid == nil || *openid != "oid-real-1" {
		t.Fatalf("openid: want oid-real-1 raw, got %v", openid)
	}

	unbound := newAuthUC(&fakeUserRP{}, &fakeAuthzRP{}, newTokenStore())
	if openid, err := unbound.WechatOpenID(ctx); err != nil || openid != nil {
		t.Fatalf("unbound: want nil,nil got %v,%v", openid, err)
	}
}

func TestGetSecurityStatusLaws(t *testing.T) {
	ctx := security.WithSubject(context.Background(), &security.Subject{UserID: "u1", TenantID: "t1"})
	status := func(hash string) *SecurityStatus {
		empty := ""
		row := &user.User{ID: "u1", Account: &empty, Enabled: true}
		if hash != "" {
			row.PasswordHash = &hash
		}
		uf := &fakeUserRP{
			row: row,
			credentials: map[string]*user.Credential{
				smsKind + "|13900001111":            {Kind: smsKind, Identifier: "13900001111", UserID: "u1"},
				fakeKind + "|oABCDEFG1234567890xyz": {Kind: fakeKind, Identifier: "oABCDEFG1234567890xyz", UserID: "u1"},
			},
		}
		s, err := newAuthUCConf(fakeProviderConf(), uf, &fakeAuthzRP{}, newTokenStore(), &fakeTenantRP{}).
			GetSecurityStatus(ctx)
		if err != nil {
			t.Fatalf("status: %v", err)
		}
		return s
	}

	withPw := status("hashed")
	if !withPw.HasPassword || len(withPw.Credentials) != 2 {
		t.Fatalf("with password: %+v", withPw)
	}
	masks := map[string]string{}
	for _, c := range withPw.Credentials {
		masks[c.Kind] = c.Identifier
	}
	if masks[smsKind] != "139****1111" || masks[fakeKind] != "oAB****0xyz" {
		t.Fatalf("masks = %v, want phone 3-****-4 and openid head-tail", masks)
	}
	if status("").HasPassword {
		t.Fatal("without password: want has_password=false")
	}

	if _, err := newAuthUCConf(fakeProviderConf(), &fakeUserRP{}, &fakeAuthzRP{}, newTokenStore(), &fakeTenantRP{}).
		GetSecurityStatus(context.Background()); !errorspb.IsGeneralErrorUnauthenticated(err) {
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

	if _, _, _, _, err := uc.LoginByIdentity(context.Background(), "fake", state, identityReq("fake", state)); !errorspb.IsInfraErrorNetworkConnection(err) {
		t.Fatalf("cash outage: want infra passthrough, got %v", err)
	}
}

// Provider rows are wiring facts: a row referencing an unregistered kind
// fails the construction, not the first login.
func TestNewAuthUCRejectsUnregisteredKind(t *testing.T) {
	bad := &conf.Auth{Providers: []*conf.Auth_Provider{{Key: "x", Kind: "ghost"}}}
	if _, err := NewAuthUC(slog.New(slog.NewTextHandler(io.Discard, nil)), nil, bad, nil, nil, nil, nil, nil, nil); err == nil {
		t.Fatal("unregistered kind: want construction error")
	}
}
