package auth

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/rs/xid"

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

const (
	sessionPrefix       = "auth:web:session:"
	sessionIdle         = 7 * 24 * time.Hour
	identityStatePrefix = "auth:web:oauth-state:"
	identityStateTTL    = 10 * time.Minute
)

var dummyDigest, _ = utils.Hash("timing-shape placeholder")

// DO ------------------------------------------------------------------------------------------------------------------

type Session struct {
	UserID     string
	TenantID   string
	Population string
	ExpiresAt  time.Time
	IssuedAt   time.Time
	Account    string
	// LivenessStamp is the engine version this payload last verified
	// against; -1 means it was minted under a failed stamp read.
	LivenessStamp int64
}

type CurrentUser struct {
	ID          string
	Account     string
	Avatar      *string
	Permissions []string
	Seats       []*user.Seat
}

// Port ----------------------------------------------------------------------------------------------------------------

type TokenRP interface {
	Get(ctx context.Context, key string) ([]byte, error)
	GetDel(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
	Del(ctx context.Context, key string) error
	Expire(ctx context.Context, key string, ttl time.Duration) error
}

// UC ------------------------------------------------------------------------------------------------------------------

type AuthUC struct {
	shared.UC
	auth         *conf.Auth
	providers    map[string]IdentityProvider
	userRP       user.UserRP
	authzRP      authz.AuthzRP
	tenantRP     tenant.TenantRP
	populationRP population.PopulationRP
	tokenRP      TokenRP
}

func NewAuthUC(logger *slog.Logger, tm shared.Transaction, auth *conf.Auth, providers map[string]IdentityProvider, userRP user.UserRP, authzRP authz.AuthzRP, tenantRP tenant.TenantRP, populationRP population.PopulationRP, tokenRP TokenRP) (*AuthUC, error) {
	if auth != nil {
		for _, row := range auth.GetProviders() {
			if _, ok := providers[row.GetKind()]; !ok {
				return nil, fmt.Errorf("auth: provider %q references unregistered kind %q", row.GetKey(), row.GetKind())
			}
		}
	}
	return &AuthUC{
		UC:           shared.NewUC(logger.With("module", "module/auth"), tm),
		auth:         auth,
		providers:    providers,
		userRP:       userRP,
		authzRP:      authzRP,
		tenantRP:     tenantRP,
		populationRP: populationRP,
		tokenRP:      tokenRP,
	}, nil
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *AuthUC) Login(ctx context.Context, account, password, seatPopulation string) (token string, issuedAt, expiresAt time.Time, err error) {
	u, err := uc.exchangePassword(ctx, account, password)
	if err != nil {
		return "", time.Time{}, time.Time{}, err
	}
	seat, err := uc.resolveSeat(ctx, u, seatPopulation)
	if err != nil {
		return "", time.Time{}, time.Time{}, err
	}
	return uc.mintSession(ctx, u, seat)
}

func (uc *AuthUC) LoginByIdentity(ctx context.Context, providerKey, state string, payload identityPayload) (token string, issuedAt, expiresAt time.Time, err error) {
	row, p, err := uc.providerRow(providerKey)
	if err != nil {
		return "", time.Time{}, time.Time{}, err
	}
	birthPopulation, err := uc.landingPopulation(ctx, row)
	if err != nil {
		return "", time.Time{}, time.Time{}, err
	}
	if err = uc.cashState(ctx, providerKey, state); err != nil {
		return "", time.Time{}, time.Time{}, err
	}
	id, err := p.Exchange(ctx, row, payload)
	if err != nil {
		return "", time.Time{}, time.Time{}, err
	}
	c, err := uc.userRP.FindCredential(ctx, id.Kind, id.Identifier)
	if err != nil {
		if errorspb.IsInfraErrorDbNotFound(err) {
			if row.AutoRegister {
				u, aerr := uc.autoRegister(ctx, id)
				if aerr != nil {
					return "", time.Time{}, time.Time{}, aerr
				}
				seat, serr := uc.resolveProviderSeat(ctx, u, birthPopulation)
				if serr != nil {
					return "", time.Time{}, time.Time{}, serr
				}
				return uc.mintSession(ctx, u, seat)
			}
			return "", time.Time{}, time.Time{}, systempb.ErrorSystemIdentityExchangeFailed("").WithCause(
				fmt.Errorf("identity %s %s has no binding", id.Kind, id.Identifier))
		}
		return "", time.Time{}, time.Time{}, err
	}
	u, err := uc.userRP.FindByID(ctx, c.UserID)
	if err != nil {
		return "", time.Time{}, time.Time{}, err
	}
	seat, err := uc.resolveProviderSeat(ctx, u, birthPopulation)
	if err != nil {
		return "", time.Time{}, time.Time{}, err
	}
	return uc.mintSession(ctx, u, seat)
}

func (uc *AuthUC) BeginIdentityLogin(ctx context.Context, providerKey, redirectURI string) (string, error) {
	row, p, err := uc.providerRow(providerKey)
	if err != nil {
		return "", err
	}
	// Only jump-capable kinds have an authorize URL; a direct-code kind
	// (no redirect) refuses the begin flow — its entry is SendSmsCode.
	rp, ok := p.(redirectProvider)
	if !ok {
		return "", systempb.ErrorSystemIdentityProviderUnknown("").WithCause(
			fmt.Errorf("provider %q has no redirect flow", providerKey))
	}
	state, err := mintIdentityState(ctx, uc.tokenRP, providerKey)
	if err != nil {
		return "", err
	}
	return rp.AuthorizeURL(row, redirectURI, state), nil
}

func (uc *AuthUC) SendSmsCode(ctx context.Context, providerKey, phone string) (string, error) {
	row, p, err := uc.providerRow(providerKey)
	if err != nil {
		return "", err
	}
	ci, ok := p.(codeIssuer)
	if !ok {
		return "", systempb.ErrorSystemIdentityProviderUnknown("").WithCause(
			fmt.Errorf("provider %q issues no codes", providerKey))
	}
	return ci.IssueCode(ctx, row, phone)
}

func (uc *AuthUC) autoRegister(ctx context.Context, id *exchangedIdentity) (*user.User, error) {
	account := xid.New().String()
	// An empty hash permanently closes the password plane: Verify against
	// "" can never succeed, so this identity logs in only through its
	// provider credential — that is the auto-registration contract.
	hash := ""
	var u *user.User
	err := uc.Tm.InTx(ctx, func(ctx context.Context) error {
		created, cerr := uc.userRP.Create(ctx, &user.User{
			Account:      &account,
			PasswordHash: &hash,
			Enabled:      true,
		})
		if cerr != nil {
			return cerr
		}
		u = created
		return uc.userRP.CreateCredential(ctx, &user.Credential{
			Kind:       id.Kind,
			Identifier: id.Identifier,
			UserID:     u.ID,
		})
	})
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (uc *AuthUC) BindCredential(ctx context.Context, providerKey, state string, payload identityPayload) error {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok {
		return errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("binding requires an authenticated caller"))
	}
	row, p, err := uc.providerRow(providerKey)
	if err != nil {
		return err
	}
	if err = uc.cashState(ctx, providerKey, state); err != nil {
		return err
	}
	id, err := p.Exchange(ctx, row, payload)
	if err != nil {
		return err
	}
	c, err := uc.userRP.FindCredential(ctx, id.Kind, id.Identifier)
	if err != nil && !errorspb.IsInfraErrorDbNotFound(err) {
		return err
	}
	if err == nil {
		// Already bound to this identity: cosmetic. Bound to another: the
		// identifier belongs to someone else.
		if c.UserID == subject.UserID {
			return nil
		}
		return systempb.ErrorSystemIdentityAlreadyBound("").WithCause(
			fmt.Errorf("identity %s %s already bound", id.Kind, id.Identifier))
	}
	if err = uc.userRP.CreateCredential(ctx, &user.Credential{Kind: id.Kind, Identifier: id.Identifier, UserID: subject.UserID}); err != nil {
		// The unique is the race backstop — same shape as the pre-check.
		if errorspb.IsInfraErrorDbConstraint(err) {
			return systempb.ErrorSystemIdentityAlreadyBound("").WithCause(err)
		}
		return err
	}
	return nil
}

func (uc *AuthUC) Logout(ctx context.Context) error {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok {
		return errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("logout requires an authenticated caller"))
	}
	// The token is the session's only key, so it doubles as the deletion
	// handle.
	return uc.tokenRP.Del(ctx, sessionPrefix+subject.SessionID)
}

func (uc *AuthUC) AuthenticateSession(ctx context.Context, token string) (*security.Subject, error) {
	// Freshness rides the liveness stamp: the payload carries the engine
	// version it last verified against, and that version moves inside the
	// very transaction that flips a gate, so a disable still lands on the
	// very next request — the row reads are paid only when it moved.
	val, err := uc.tokenRP.Get(ctx, sessionPrefix+token)
	if err != nil {
		if errorspb.IsInfraErrorCacheMiss(err) {
			return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("session credential not recognized"))
		}
		return nil, err
	}
	sess, err := utils.Unmarshal[Session](val)
	if err != nil {
		return nil, err
	}
	if time.Now().After(sess.ExpiresAt) {
		if err := uc.tokenRP.Del(ctx, sessionPrefix+token); err != nil {
			return nil, err
		}
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("session passed its absolute cap"))
	}
	if val, err := uc.tokenRP.Get(ctx, shared.RevokedMarkerPrefix+sess.UserID); err != nil {
		if !errorspb.IsInfraErrorCacheMiss(err) {
			return nil, err
		}
	} else {
		revokedAt, uerr := utils.Unmarshal[time.Time](val)
		if uerr != nil {
			return nil, uerr
		}
		// Sessions issued at or before the marker die; later logins survive —
		// the semantics are a kick, not a ban. Sessions created before this
		// field existed unmarshal to the zero time, which any marker kills.
		if !sess.IssuedAt.After(revokedAt) {
			if derr := uc.tokenRP.Del(ctx, sessionPrefix+token); derr != nil {
				return nil, derr
			}
			return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("session revoked by administration"))
		}
	}
	ver, serr := uc.authzRP.StampVersion(ctx)
	if serr != nil {
		uc.Log.Warn("auth: liveness stamp read failed; verifying deep", "error", serr)
	}
	if serr != nil || sess.LivenessStamp != ver {
		// A disabled tenant kills its inhabitants' sessions at verification — a
		// check, not a revocation: re-enabling resurrects stored sessions.
		live, err := uc.tenantLive(ctx, sess.TenantID)
		if err != nil {
			return nil, err
		}
		if !live {
			return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("tenant disabled: %s", sess.TenantID))
		}
		u, err := uc.userRP.FindByID(ctx, sess.UserID)
		if err != nil {
			if errorspb.IsInfraErrorDbNotFound(err) {
				return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("identity missing: %s", sess.UserID))
			}
			return nil, err
		}
		if !u.Enabled {
			return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("identity disabled: %s", sess.UserID))
		}
		if serr == nil {
			// The survivor carries the current version so its next request
			// rides the comparison; a failed write-back only costs the next
			// request its deep reads, never its correctness.
			sess.LivenessStamp = ver
			if werr := uc.tokenRP.Set(ctx, sessionPrefix+token, utils.MustMarshal(sess), sessionIdle); werr != nil {
				uc.Log.Warn("auth: liveness stamp write-back failed; next request verifies deep", "error", werr)
			}
		}
	}
	slide := min(sessionIdle, time.Until(sess.ExpiresAt))
	if err := uc.tokenRP.Expire(ctx, sessionPrefix+token, slide); err != nil {
		return nil, err
	}
	plane, err := uc.authzRP.PlaneOf(sess.Population)
	if err != nil {
		return nil, err
	}
	return &security.Subject{UserID: sess.UserID, TenantID: sess.TenantID, Population: sess.Population, Plane: plane, SessionID: token, Extra: map[string]any{shared.ExtraAccount: sess.Account}}, nil
}

func (uc *AuthUC) GetCurrentUser(ctx context.Context) (*CurrentUser, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("no subject in context"))
	}
	account, _ := subject.Extra[shared.ExtraAccount].(string)
	grants, err := uc.authzRP.EffectiveGrants(ctx, subject)
	if err != nil {
		// A session without a resolved grant set would silently masquerade
		// as "no permissions"; engine failure is infra, not an empty grant.
		return nil, errorspb.ErrorGeneralErrorInternal("").WithCause(fmt.Errorf("resolve grants: %w", err))
	}
	perms := make([]string, 0, len(grants))
	for _, g := range grants {
		perms = append(perms, g.Pattern)
	}
	var avatar *string
	if row, err := uc.userRP.FindByID(ctx, subject.UserID); err == nil {
		avatar = row.Avatar
	} else if !errorspb.IsInfraErrorDbNotFound(err) {
		uc.Log.Warn("avatar hydration failed", "user_id", subject.UserID, "error", err)
	}
	seats, err := uc.userRP.ListSeatsByUser(ctx, subject.UserID)
	if err != nil {
		return nil, err
	}
	return &CurrentUser{ID: subject.UserID, Account: account, Avatar: avatar, Permissions: perms, Seats: seats}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func (uc *AuthUC) providerRow(providerKey string) (*conf.Auth_Provider, IdentityProvider, error) {
	for _, row := range uc.auth.GetProviders() {
		if row.GetKey() == providerKey {
			if !row.GetEnabled() {
				break
			}
			return row, uc.providers[row.GetKind()], nil
		}
	}
	return nil, nil, systempb.ErrorSystemIdentityProviderUnknown("").WithCause(fmt.Errorf("identity provider %q", providerKey))
}

func (uc *AuthUC) tenantLive(ctx context.Context, code string) (bool, error) {
	tn, err := uc.tenantRP.FindByCode(ctx, code)
	if err != nil && !errorspb.IsInfraErrorDbNotFound(err) {
		return false, err
	}
	return err == nil && tn.Enabled, nil
}

func (uc *AuthUC) cashState(ctx context.Context, providerKey, state string) error {
	stored, err := uc.tokenRP.GetDel(ctx, identityStatePrefix+state)
	if err != nil {
		if errorspb.IsInfraErrorCacheMiss(err) {
			return systempb.ErrorSystemIdentityExchangeFailed("").WithCause(fmt.Errorf("state not recognized"))
		}
		return err
	}
	if string(stored) != providerKey {
		return systempb.ErrorSystemIdentityExchangeFailed("").WithCause(fmt.Errorf("state not recognized"))
	}
	return nil
}

func (uc *AuthUC) exchangePassword(ctx context.Context, account, password string) (*user.User, error) {
	u, err := uc.userRP.FindByAccount(ctx, account)
	if err != nil {
		if errorspb.IsInfraErrorDbNotFound(err) {
			utils.Verify(password, dummyDigest)
			return nil, systempb.ErrorSystemLoginFailed("").WithCause(fmt.Errorf("account not found: %s", account))
		}
		return nil, err
	}
	if !utils.Verify(password, *u.PasswordHash) {
		return nil, systempb.ErrorSystemLoginFailed("").WithCause(fmt.Errorf("password mismatch: %s", account))
	}
	// Lazy digest upgrade: a legacy row re-hashes as argon2id on the login
	// that verified it; a failed upgrade warns and never blocks the login.
	if utils.NeedsRehash(*u.PasswordHash) {
		if hash, herr := utils.Hash(password); herr == nil {
			if _, uerr := uc.userRP.Update(ctx, []string{"password"}, &user.User{ID: u.ID, PasswordHash: &hash}); uerr != nil {
				uc.Log.Warn("digest upgrade failed", "user_id", u.ID, "error", uerr)
			}
		} else {
			uc.Log.Warn("digest upgrade failed", "user_id", u.ID, "error", herr)
		}
	}
	return u, nil
}

func (uc *AuthUC) resolveSeat(ctx context.Context, u *user.User, hint string) (*user.Seat, error) {
	if !u.Enabled {
		return nil, systempb.ErrorSystemUserDisabled("").WithCause(fmt.Errorf("user disabled: %s", utils.Deref(u.Account, "")))
	}
	seats, err := uc.userRP.ListSeatsByUser(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	if len(seats) == 0 {
		return nil, errorspb.ErrorGeneralErrorInternal("").WithCause(fmt.Errorf("identity without a seat: %s", u.ID))
	}
	if hint != "" {
		for _, seat := range seats {
			if seat.Population == hint {
				return uc.liveSeat(ctx, seat)
			}
		}
		return nil, systempb.ErrorSystemLoginSeatRequired("").WithCause(fmt.Errorf("identity holds no seat in population %q", hint))
	}
	if len(seats) > 1 {
		return nil, systempb.ErrorSystemLoginSeatRequired("").WithCause(fmt.Errorf("identity holds %d seats; a choice is required", len(seats)))
	}
	return uc.liveSeat(ctx, seats[0])
}

func (uc *AuthUC) resolveProviderSeat(ctx context.Context, u *user.User, population string) (*user.Seat, error) {
	if !u.Enabled {
		return nil, systempb.ErrorSystemUserDisabled("").WithCause(fmt.Errorf("user disabled: %s", utils.Deref(u.Account, "")))
	}
	seat, err := uc.seatInPopulation(ctx, u.ID, population)
	if err != nil {
		return nil, err
	}
	if seat != nil {
		return uc.liveSeat(ctx, seat)
	}
	// A bound identity enters the provider's population here — entering the
	// face is the provider's doing, never a prior administrative act. The
	// unique index makes the ensure idempotent.
	if serr := uc.userRP.EnsureSeat(ctx, &user.Seat{UserID: u.ID, TenantID: security.DefaultTenant, Population: population}); serr != nil {
		return nil, serr
	}
	seat, err = uc.seatInPopulation(ctx, u.ID, population)
	if err != nil {
		return nil, err
	}
	if seat == nil {
		return nil, errorspb.ErrorGeneralErrorInternal("").WithCause(fmt.Errorf("seat ensure left no seat: %s %s", u.ID, population))
	}
	return uc.liveSeat(ctx, seat)
}

func (uc *AuthUC) landingPopulation(ctx context.Context, row *conf.Auth_Provider) (string, error) {
	code := row.GetPopulation()
	if code == "" {
		return security.DefaultPopulation, nil
	}
	if _, err := uc.populationRP.FindByCode(ctx, code); err != nil {
		if errorspb.IsInfraErrorDbNotFound(err) {
			return "", systempb.ErrorSystemPopulationNotFound("").WithCause(fmt.Errorf("provider %q references population %q", row.GetKey(), code))
		}
		return "", err
	}
	return code, nil
}

func (uc *AuthUC) seatInPopulation(ctx context.Context, userID, population string) (*user.Seat, error) {
	seats, err := uc.userRP.ListSeatsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, seat := range seats {
		if seat.Population == population {
			return seat, nil
		}
	}
	return nil, nil
}

func (uc *AuthUC) liveSeat(ctx context.Context, seat *user.Seat) (*user.Seat, error) {
	live, err := uc.tenantLive(ctx, seat.TenantID)
	if err != nil {
		return nil, err
	}
	if !live {
		return nil, systempb.ErrorSystemTenantDisabled("").WithCause(fmt.Errorf("tenant disabled: %s", seat.TenantID))
	}
	return seat, nil
}

func (uc *AuthUC) mintSession(ctx context.Context, u *user.User, seat *user.Seat) (token string, issuedAt, expiresAt time.Time, err error) {
	token, err = utils.RandomToken()
	if err != nil {
		return "", time.Time{}, time.Time{}, err
	}
	issuedAt = time.Now()
	expiresAt = issuedAt.Add(shared.SessionMax)
	// A failed stamp read mints -1 — a version no bump produces — so the
	// session simply verifies deep until its first write-back.
	ver, verr := uc.authzRP.StampVersion(ctx)
	if verr != nil {
		uc.Log.Warn("auth: liveness stamp read failed at mint; session verifies deep", "error", verr)
		ver = -1
	}
	sess := &Session{UserID: u.ID, TenantID: seat.TenantID, Population: seat.Population, ExpiresAt: expiresAt, IssuedAt: issuedAt, Account: utils.Deref(u.Account, ""), LivenessStamp: ver}
	if err = uc.tokenRP.Set(ctx, sessionPrefix+token, utils.MustMarshal(sess), sessionIdle); err != nil {
		return "", time.Time{}, time.Time{}, err
	}
	return token, issuedAt, expiresAt, nil
}
