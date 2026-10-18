package auth

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"cyber-ecosystem/shared-go/kratos/security"
	"cyber-ecosystem/shared-go/utils"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/conf"
	"cyber-ecosystem/app/services/system/internal/module/authz"
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

// fakeUserRP backs the Login slice: one seeded row, email lookup. updates
// captures what the lazy rehash writes.
type fakeUserRP struct {
	user.UserRP
	row     *user.User
	findErr error
	updates []struct {
		fields []string
		u      *user.User
	}
	updateErr error
	// Seats is the seat store; an explicit empty slice means seatless (a
	// nil slice takes the helper's degenerate default).
	seats    []*user.Seat
	seatsErr error

	// credentials is the binding store keyed "kind|identifier".
	credentials    map[string]*user.Credential
	createCredErr  error
	createdCreds   []*user.Credential
	deletedCredFor []string
	findIDErr      error

	// createdUsers records the auto-register births.
	createdUsers []*user.User
}

func (f *fakeUserRP) ListSeatsByUser(_ context.Context, _ string) ([]*user.Seat, error) {
	if f.seatsErr != nil {
		return nil, f.seatsErr
	}
	return f.seats, nil
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

func (f *fakeUserRP) FindByEmail(_ context.Context, email string) (*user.User, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	if f.row != nil && *f.row.Email == email {
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
// — or the injected engine error, for the infra-failure path.
type fakeAuthzRP struct {
	authz.AuthzRP
	grants []authz.GrantView
	err    error
}

func (f *fakeAuthzRP) EffectiveGrants(_ context.Context, _ *security.Subject) ([]authz.GrantView, error) {
	return f.grants, f.err
}

// Private -------------------------------------------------------------------------------------------------------------

func newAuthUC(u *fakeUserRP, a *fakeAuthzRP, t *fakeTokenRP) *AuthUC {
	return newAuthUCTenant(u, a, t, &fakeTenantRP{row: &tenant.Tenant{Code: "any", Enabled: true}})
}

func newAuthUCTenant(u *fakeUserRP, a *fakeAuthzRP, t *fakeTokenRP, tn *fakeTenantRP) *AuthUC {
	return newAuthUCConf(nil, u, a, t, tn)
}

// passthroughTx runs the unit inline — the auth slice asserts transaction
// semantics through the fakes, not a real store.
type passthroughTx struct{}

func (passthroughTx) InTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func newAuthUCConf(auth *conf.Auth, u *fakeUserRP, a *fakeAuthzRP, t *fakeTokenRP, tn *fakeTenantRP) *AuthUC { // The default seat mirrors the degenerate world: one seat on t1 for u1.
	if u.seats == nil {
		u.seats = []*user.Seat{{UserID: "u1", TenantID: "t1", Population: security.DefaultPopulation}}
	}
	uc, err := NewAuthUC(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		passthroughTx{}, auth, u, a, tn, t,
	)
	if err != nil {
		panic(err)
	}
	return uc
}

// fakeTenantRP answers the registry read auth consults: an injectable row
// (nil = the classified miss) and an injectable error.
type fakeTenantRP struct {
	tenant.TenantRP
	row *tenant.Tenant
	err error
}

func (f *fakeTenantRP) FindByCode(_ context.Context, code string) (*tenant.Tenant, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.row == nil {
		return nil, errorspb.ErrorInfraErrorDbNotFound("")
	}
	return f.row, nil
}

func seedRow(t *testing.T, email, password string) *user.User {
	t.Helper()
	hash, err := utils.Hash(password)
	if err != nil {
		t.Fatalf("seed hash: %v", err)
	}
	return &user.User{
		ID: "u1", Email: &email, PasswordHash: &hash, Enabled: true,
	}
}

// mintSession writes a session key directly through the fake store — the
// authenticate-family tests own their session shapes.
func mintSession(t *testing.T, rp *fakeTokenRP, sess *Session) string {
	t.Helper()
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

// Login ---------------------------------------------------------------------------------------------------------------

// Anti-enumeration: an unknown email and a wrong password answer in the same
// shape — which emails exist is not probeable from the wire (the not-found
// branch burns the dummy digest so timing carries no signal either; that half
// is not unit-assertable, the shape half is).
func TestLoginUnknownEmailSameShapeAsWrongPassword(t *testing.T) {
	uc := newAuthUC(&fakeUserRP{row: seedRow(t, "a@x.co", "right")}, &fakeAuthzRP{}, &fakeTokenRP{vals: map[string][]byte{}, ttls: map[string]time.Duration{}})

	_, _, _, errUnknown := uc.Login(context.Background(), "ghost@x.co", "right")
	if !systempb.IsSystemLoginFailed(errUnknown) {
		t.Fatalf("unknown email: want LOGIN_FAILED, got %v", errUnknown)
	}
	_, _, _, errWrong := uc.Login(context.Background(), "a@x.co", "wrong")
	if !systempb.IsSystemLoginFailed(errWrong) {
		t.Fatalf("wrong password: want LOGIN_FAILED, got %v", errWrong)
	}
}

// The deeper order law: a disabled account with a wrong password still reports
// LOGIN_FAILED — status only surfaces once the credential passes, so account
// state is as unprobeable as account existence.
func TestLoginDisabledUserStatusAfterCredential(t *testing.T) {
	row := seedRow(t, "a@x.co", "right")
	row.Enabled = false
	uc := newAuthUC(&fakeUserRP{row: row}, &fakeAuthzRP{}, &fakeTokenRP{vals: map[string][]byte{}, ttls: map[string]time.Duration{}})

	if _, _, _, err := uc.Login(context.Background(), "a@x.co", "wrong"); !systempb.IsSystemLoginFailed(err) {
		t.Fatalf("wrong password on disabled account: want LOGIN_FAILED, got %v", err)
	}
	if _, _, _, err := uc.Login(context.Background(), "a@x.co", "right"); !systempb.IsSystemUserDisabled(err) {
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
	email := "a@x.co"
	legacyStr := string(legacy)
	row := &user.User{ID: "u1", Email: &email, PasswordHash: &legacyStr, Enabled: true}
	rp := &fakeUserRP{row: row}
	uc := newAuthUC(rp, &fakeAuthzRP{}, &fakeTokenRP{vals: map[string][]byte{}, ttls: map[string]time.Duration{}})

	if _, _, _, err := uc.Login(context.Background(), "a@x.co", "right"); err != nil {
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
	if _, _, _, err := uc.Login(context.Background(), "a@x.co", "right"); err != nil {
		t.Fatalf("upgrade failure must not block login: %v", err)
	}
}

// The mint's shape is the login contract: the absolute cap derived from
// SessionMax, the idle TTL on the key, and the row's identity in the payload.
func TestLoginMintsSessionShape(t *testing.T) {
	row := seedRow(t, "a@x.co", "right")
	trp := &fakeTokenRP{vals: map[string][]byte{}, ttls: map[string]time.Duration{}}
	uc := newAuthUC(&fakeUserRP{row: row}, &fakeAuthzRP{}, trp)

	token, issuedAt, expiresAt, err := uc.Login(context.Background(), "a@x.co", "right")
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
	if sess.UserID != row.ID || sess.Email != "a@x.co" {
		t.Fatalf("mint: payload = %+v", sess)
	}
}

// AuthenticateSession -------------------------------------------------------------------------------------------------

// A session inside its lifetime but below the refresh watermark slides its TTL.
func TestAuthenticateSessionSlidesLowTTL(t *testing.T) {
	trp := &fakeTokenRP{vals: map[string][]byte{}, ttls: map[string]time.Duration{}}
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
	trp := &fakeTokenRP{vals: map[string][]byte{}, ttls: map[string]time.Duration{}}
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
	trp := &fakeTokenRP{vals: map[string][]byte{}, ttls: map[string]time.Duration{}}
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
	trp := &fakeTokenRP{vals: map[string][]byte{}, ttls: map[string]time.Duration{}}
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

// Logout / me ---------------------------------------------------------------------------------------------------------

// The token is the session's only key, so it doubles as the deletion handle —
// logout removes exactly that key; a callerless logout is unauthenticated.
func TestLogoutDeletesSessionKey(t *testing.T) {
	trp := &fakeTokenRP{vals: map[string][]byte{}, ttls: map[string]time.Duration{}}
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
	uc := newAuthUC(&fakeUserRP{}, &fakeAuthzRP{err: errorspb.ErrorInfraErrorNetworkConnection("")}, &fakeTokenRP{vals: map[string][]byte{}, ttls: map[string]time.Duration{}})
	subject := &security.Subject{UserID: "u1", SessionID: "tok-1", Extra: map[string]any{shared.ExtraEmail: "a@x.co"}}

	if _, err := uc.GetCurrentUser(authCtx(subject)); !errorspb.IsGeneralErrorInternal(err) {
		t.Fatalf("engine failure: want INTERNAL, got %v", err)
	}
}

// A store failure mid-authentication surfaces as itself — infra never
// masquerades as an unauthenticated.
func TestAuthenticateSessionExpireFailureSurfaces(t *testing.T) {
	trp := &fakeTokenRP{vals: map[string][]byte{}, ttls: map[string]time.Duration{}, expireErr: errorspb.ErrorInfraErrorNetworkConnection("")}
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
	trp := &fakeTokenRP{vals: map[string][]byte{}, ttls: map[string]time.Duration{}}
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

// A user lookup that fails beyond not-found (store down) passes through — the
// login path never invents a domain shape for infra trouble.
func TestLoginFindByEmailGenericErrorPassthrough(t *testing.T) {
	uc := newAuthUC(&fakeUserRP{findErr: errorspb.ErrorInfraErrorNetworkConnection("")}, &fakeAuthzRP{}, &fakeTokenRP{vals: map[string][]byte{}, ttls: map[string]time.Duration{}})

	if _, _, _, err := uc.Login(context.Background(), "a@x.co", "pw"); !errorspb.IsInfraErrorNetworkConnection(err) {
		t.Fatalf("generic lookup error: want infra passthrough, got %v", err)
	}
}

// The me face hydrates from the session's extra, the engine's patterns, and
// the row's avatar; a missing avatar row degrades to nil, never an error.
func TestGetCurrentUserHydrates(t *testing.T) {
	row := seedRow(t, "a@x.co", "right")
	avatar := "f1"
	row.Avatar = &avatar
	uc := newAuthUC(&fakeUserRP{row: row}, &fakeAuthzRP{grants: []authz.GrantView{{Pattern: "/svc/*"}}}, &fakeTokenRP{vals: map[string][]byte{}, ttls: map[string]time.Duration{}})
	subject := &security.Subject{UserID: "u1", SessionID: "tok-1", Extra: map[string]any{shared.ExtraEmail: "a@x.co"}}

	out, err := uc.GetCurrentUser(authCtx(subject))
	if err != nil {
		t.Fatalf("me: %v", err)
	}
	if out.Email != "a@x.co" || len(out.Permissions) != 1 || out.Permissions[0] != "/svc/*" || out.Avatar == nil || *out.Avatar != "f1" {
		t.Fatalf("me: out = %+v", out)
	}

	// A missing avatar row degrades to nil.
	uc2 := newAuthUC(&fakeUserRP{}, &fakeAuthzRP{grants: []authz.GrantView{{Pattern: "/svc/*"}}}, &fakeTokenRP{vals: map[string][]byte{}, ttls: map[string]time.Duration{}})
	out2, err := uc2.GetCurrentUser(authCtx(subject))
	if err != nil {
		t.Fatalf("me degraded: %v", err)
	}
	if out2.Avatar != nil {
		t.Fatalf("me degraded: avatar = %v, want nil", *out2.Avatar)
	}
}

// Tenant enforcement --------------------------------------------------------------------------------------------------

// The tenant's disabled state surfaces only after the credential passes —
// tenant state is as unprobeable as account existence.
func TestLoginDisabledTenantAfterCredential(t *testing.T) {
	tok := &fakeTokenRP{vals: map[string][]byte{}, ttls: map[string]time.Duration{}}
	uc := newAuthUCTenant(&fakeUserRP{row: seedRow(t, "a@x.co", "right")}, &fakeAuthzRP{}, tok,
		&fakeTenantRP{row: &tenant.Tenant{Code: "t1", Enabled: false}})

	if _, _, _, err := uc.Login(context.Background(), "a@x.co", "wrong"); !systempb.IsSystemLoginFailed(err) {
		t.Fatalf("wrong password on disabled tenant: want LOGIN_FAILED, got %v", err)
	}
	if _, _, _, err := uc.Login(context.Background(), "a@x.co", "right"); !systempb.IsSystemTenantDisabled(err) {
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
		&fakeTokenRP{vals: map[string][]byte{}, ttls: map[string]time.Duration{}}, &fakeTenantRP{})

	if _, _, _, err := uc.Login(context.Background(), "a@x.co", "right"); !systempb.IsSystemTenantDisabled(err) {
		t.Fatalf("missing registry row: want TENANT_DISABLED, got %v", err)
	}
}

// A generic registry error keeps its infra identity.
func TestLoginTenantGenericPassthrough(t *testing.T) {
	uc := newAuthUCTenant(&fakeUserRP{row: seedRow(t, "a@x.co", "right")}, &fakeAuthzRP{},
		&fakeTokenRP{vals: map[string][]byte{}, ttls: map[string]time.Duration{}},
		&fakeTenantRP{err: errorspb.ErrorInfraErrorNetworkConnection("")})

	if _, _, _, err := uc.Login(context.Background(), "a@x.co", "right"); !errorspb.IsInfraErrorNetworkConnection(err) {
		t.Fatalf("generic registry error: want passthrough, got %v", err)
	}
}

// A tenant disabled mid-session dies at verification — a check, not a
// revocation: the key stays, re-enabling resurrects the stored session, and a
// missing row answers in the same shape.
func TestAuthenticateSessionTenantGate(t *testing.T) {
	rp := &fakeTokenRP{vals: map[string][]byte{}, ttls: map[string]time.Duration{}}
	token := mintSession(t, rp, &Session{UserID: "u1", TenantID: "t1", Email: "a@x.co", IssuedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour)})
	tn := &fakeTenantRP{row: &tenant.Tenant{Code: "t1", Enabled: false}}
	uc := newAuthUCTenant(&fakeUserRP{row: seedRow(t, "a@x.co", "right")}, &fakeAuthzRP{}, rp, tn)

	if _, err := uc.AuthenticateSession(context.Background(), token); !errorspb.IsGeneralErrorUnauthenticated(err) {
		t.Fatalf("disabled tenant: want unauthenticated, got %v", err)
	}
	if _, ok := rp.vals[sessionPrefix+token]; !ok {
		t.Fatal("disabled tenant: the gate must not revoke — the key stays for resurrection")
	}

	tn.row.Enabled = true
	if _, err := uc.AuthenticateSession(context.Background(), token); err != nil {
		t.Fatalf("re-enabled tenant: session must resurrect, got %v", err)
	}

	tn.row = nil
	if _, err := uc.AuthenticateSession(context.Background(), token); !errorspb.IsGeneralErrorUnauthenticated(err) {
		t.Fatalf("missing registry row: want unauthenticated same shape, got %v", err)
	}

	tn.err = errorspb.ErrorInfraErrorNetworkConnection("")
	if _, err := uc.AuthenticateSession(context.Background(), token); !errorspb.IsInfraErrorNetworkConnection(err) {
		t.Fatalf("generic registry error: want passthrough, got %v", err)
	}
}

// A disabled identity dies at verification the same way its tenant does —
// a check, not a revocation: re-enabling resurrects stored sessions, a
// missing row answers in the same shape, and the read backstops the
// disable path's best-effort marker write.
func TestAuthenticateSessionUserGate(t *testing.T) {
	rp := &fakeTokenRP{vals: map[string][]byte{}, ttls: map[string]time.Duration{}}
	token := mintSession(t, rp, &Session{UserID: "u1", TenantID: "t1", Email: "a@x.co", IssuedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour)})
	row := seedRow(t, "a@x.co", "right")
	uf := &fakeUserRP{row: row}
	uc := newAuthUCTenant(uf, &fakeAuthzRP{}, rp, &fakeTenantRP{row: &tenant.Tenant{Code: "t1", Enabled: true}})

	row.Enabled = false
	if _, err := uc.AuthenticateSession(context.Background(), token); !errorspb.IsGeneralErrorUnauthenticated(err) {
		t.Fatalf("disabled identity: want unauthenticated, got %v", err)
	}
	if _, ok := rp.vals[sessionPrefix+token]; !ok {
		t.Fatal("disabled identity: the gate must not revoke — the key stays for resurrection")
	}

	row.Enabled = true
	if _, err := uc.AuthenticateSession(context.Background(), token); err != nil {
		t.Fatalf("re-enabled identity: session must resurrect, got %v", err)
	}

	uf.row = nil
	if _, err := uc.AuthenticateSession(context.Background(), token); !errorspb.IsGeneralErrorUnauthenticated(err) {
		t.Fatalf("missing identity row: want unauthenticated same shape, got %v", err)
	}

	uf.findIDErr = errorspb.ErrorInfraErrorNetworkConnection("")
	if _, err := uc.AuthenticateSession(context.Background(), token); !errorspb.IsInfraErrorNetworkConnection(err) {
		t.Fatalf("generic identity error: want passthrough, got %v", err)
	}
}

// Membership face ----------------------------------------------------------------------------------------------------

// The me face carries the caller's seats — tenancy lives on seats, never
// on the identity row.
func TestGetCurrentUserCarriesSeats(t *testing.T) {
	subject := &security.Subject{UserID: "u1", TenantID: "t1", SessionID: "s1"}
	uf := &fakeUserRP{seats: []*user.Seat{
		{UserID: "u1", TenantID: "default", Population: "user"},
	}}
	uc := newAuthUCTenant(uf, &fakeAuthzRP{grants: []authz.GrantView{{Pattern: "/svc/*"}}},
		&fakeTokenRP{vals: map[string][]byte{}, ttls: map[string]time.Duration{}},
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

// Login anatomy ------------------------------------------------------------------------------------------------------

// The session's tenant and population come from the resolved seat, never
// from the user row's columns — the user row may carry a tenant column for
// row scoping, but the seat is what the login occupies.
func TestLoginMintsFromSeat(t *testing.T) {
	rp := &fakeTokenRP{vals: map[string][]byte{}, ttls: map[string]time.Duration{}}
	uc := newAuthUCTenant(&fakeUserRP{row: seedRow(t, "a@x.co", "right"),
		seats: []*user.Seat{{UserID: "u1", TenantID: "t9", Population: "user"}}}, &fakeAuthzRP{}, rp,
		&fakeTenantRP{row: &tenant.Tenant{Code: "t9", Enabled: true}})

	token, _, _, err := uc.Login(context.Background(), "a@x.co", "right")
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
		&fakeTokenRP{vals: map[string][]byte{}, ttls: map[string]time.Duration{}},
		&fakeTenantRP{row: &tenant.Tenant{Code: "t1", Enabled: true}})

	if _, _, _, err := uc.Login(context.Background(), "a@x.co", "right"); !errorspb.IsGeneralErrorInternal(err) {
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
	uc := newAuthUCTenant(uf, &fakeAuthzRP{},
		&fakeTokenRP{vals: map[string][]byte{}, ttls: map[string]time.Duration{}},
		&fakeTenantRP{row: &tenant.Tenant{Code: "t1", Enabled: true}})

	if _, _, _, err := uc.Login(context.Background(), "a@x.co", "right"); !systempb.IsSystemLoginSeatRequired(err) {
		t.Fatalf("multi-seat: want SEAT_REQUIRED, got %v", err)
	}
}

// The verified subject carries the seat's population downstream.
func TestAuthenticateSessionCarriesPopulation(t *testing.T) {
	rp := &fakeTokenRP{vals: map[string][]byte{}, ttls: map[string]time.Duration{}}
	token := mintSession(t, rp, &Session{UserID: "u1", TenantID: "t1", Population: "user", Email: "a@x.co", IssuedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour)})
	uc := newAuthUCTenant(&fakeUserRP{row: seedRow(t, "a@x.co", "right")}, &fakeAuthzRP{}, rp,
		&fakeTenantRP{row: &tenant.Tenant{Code: "t1", Enabled: true}})

	sub, err := uc.AuthenticateSession(context.Background(), token)
	if err != nil || sub.Population != "user" {
		t.Fatalf("verify: sub=%+v err=%v", sub, err)
	}
}
