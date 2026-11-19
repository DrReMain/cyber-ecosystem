package auth

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"unicode/utf8"

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
	// wechat's own state cap; a folded relay must fit or it is dropped.
	identityStateMaxLen = 128
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

type SecurityCredential struct {
	Kind       string
	Identifier string // display-masked; the raw value never crosses the boundary
}

type SecurityStatus struct {
	HasPassword bool
	Credentials []*SecurityCredential
}

// Port ----------------------------------------------------------------------------------------------------------------

type TokenRP interface {
	Get(ctx context.Context, key string) ([]byte, error)
	GetDel(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
	// CompareAndSet swaps val in only when the key still holds expected —
	// the session write-back's guard against resurrecting a key deleted
	// by logout or revoke mid-request.
	CompareAndSet(ctx context.Context, key string, expected, val []byte, ttl time.Duration) (bool, error)
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
	clock        shared.Clock // injected time source — session expiry and minting are test-deterministic
}

func NewAuthUC(logger *slog.Logger, tm shared.Transaction, auth *conf.Auth, providers map[string]IdentityProvider, userRP user.UserRP, authzRP authz.AuthzRP, tenantRP tenant.TenantRP, populationRP population.PopulationRP, tokenRP TokenRP) (*AuthUC, error) {
	if auth != nil {
		for _, row := range auth.GetProviders() {
			if _, ok := providers[row.GetKind()]; !ok {
				return nil, fmt.Errorf("auth: provider %q references unregistered kind %q", row.GetKey(), row.GetKind())
			}
			// The boot-time resolution banner: one line per configured row
			// naming the shape the service actually runs — the endpoint
			// verdict is the wechat family's own (empty = the official
			// bases, the e2e seam overrides), so an operator reading the
			// boot log sees the channel topology without reading assembly
			// code.
			endpoint := "kind-default"
			if row.GetEndpoint() != "" {
				endpoint = "override:" + row.GetEndpoint()
			}
			logger.Info("auth provider resolved", "key", row.GetKey(), "kind", row.GetKind(), "enabled", row.GetEnabled(), "auto_register", row.GetAutoRegister(), "endpoint", endpoint)
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
		clock:        time.Now,
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

func (uc *AuthUC) LoginByIdentity(ctx context.Context, providerKey, state string, payload identityPayload) (token string, issuedAt, expiresAt time.Time, profile *wechatProfile, err error) {
	row, p, err := uc.providerRow(providerKey)
	if err != nil {
		return "", time.Time{}, time.Time{}, nil, err
	}
	birthPopulation, err := uc.landingPopulation(ctx, row)
	if err != nil {
		return "", time.Time{}, time.Time{}, nil, err
	}
	if err = uc.cashState(ctx, providerKey, state); err != nil {
		return "", time.Time{}, time.Time{}, nil, err
	}
	id, err := p.Exchange(ctx, row, payload)
	if err != nil {
		return "", time.Time{}, time.Time{}, nil, err
	}
	c, err := uc.userRP.FindCredential(ctx, id.Kind, id.Identifier)
	if err != nil {
		if errorspb.IsInfraErrorDbNotFound(err) {
			if row.AutoRegister {
				u, aerr := uc.autoRegister(ctx, id)
				if aerr != nil {
					return "", time.Time{}, time.Time{}, nil, aerr
				}
				seat, serr := uc.resolveProviderSeat(ctx, u, birthPopulation)
				if serr != nil {
					return "", time.Time{}, time.Time{}, nil, serr
				}
				token, issuedAt, expiresAt, err = uc.mintSession(ctx, u, seat)
				return token, issuedAt, expiresAt, id.Profile, err
			}
			return "", time.Time{}, time.Time{}, nil, systempb.ErrorSystemIdentityExchangeFailed("").WithCause(
				fmt.Errorf("identity %s %s has no binding", id.Kind, id.Identifier))
		}
		return "", time.Time{}, time.Time{}, nil, err
	}
	u, err := uc.userRP.FindByID(ctx, c.UserID)
	if err != nil {
		return "", time.Time{}, time.Time{}, nil, err
	}
	seat, err := uc.resolveProviderSeat(ctx, u, birthPopulation)
	if err != nil {
		return "", time.Time{}, time.Time{}, nil, err
	}
	token, issuedAt, expiresAt, err = uc.mintSession(ctx, u, seat)
	return token, issuedAt, expiresAt, id.Profile, err
}

func (uc *AuthUC) BeginIdentityLogin(ctx context.Context, providerKey, redirectURI, next string) (string, error) {
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
	// next rides inside the state the mint returns — the redirect_uri
	// cannot carry it (wechat rejects embedded queries).
	state, err := mintIdentityState(ctx, uc.tokenRP, providerKey, next)
	if err != nil {
		return "", err
	}
	return rp.AuthorizeURL(row, redirectURI, state), nil
}

func (uc *AuthUC) SendSmsCode(ctx context.Context, providerKey, phone, source string) (string, error) {
	row, p, err := uc.providerRow(providerKey)
	if err != nil {
		return "", err
	}
	ci, ok := p.(codeIssuer)
	if !ok {
		return "", systempb.ErrorSystemIdentityProviderUnknown("").WithCause(
			fmt.Errorf("provider %q issues no codes", providerKey))
	}
	return ci.IssueCode(ctx, row, phone, source)
}

func (uc *AuthUC) LoginConfig(ctx context.Context, plane string) ([]*systempb.LoginProvider, error) {
	// The anonymous discovery face: the login page renders exactly what the
	// deployment wired — keys and kinds only, no secrets, no endpoint
	// topology. An empty answer is a closed deployment, a shape the page
	// must own instead of a dead button. A plane ask narrows the answer to
	// providers whose landing population seats that plane — the population
	// row owns the plane, and a provider naming a population the table does
	// not hold stays unlisted rather than guessed at: a login that can only
	// mint a session the plane guard would refuse is not an entry point.
	out := make([]*systempb.LoginProvider, 0)
	if uc.auth == nil {
		return out, nil
	}
	planes := make(map[string]string)
	for _, row := range uc.auth.GetProviders() {
		if !row.GetEnabled() {
			continue
		}
		if plane == "" {
			out = append(out, &systempb.LoginProvider{Key: row.GetKey(), Kind: row.GetKind()})
			continue
		}
		p, ok := planes[row.GetPopulation()]
		if !ok {
			pop, err := uc.populationRP.FindByCode(ctx, row.GetPopulation())
			if err != nil {
				if errorspb.IsInfraErrorDbNotFound(err) {
					continue
				}
				return nil, err
			}
			p = pop.Plane
			planes[row.GetPopulation()] = p
		}
		if p == plane {
			out = append(out, &systempb.LoginProvider{Key: row.GetKey(), Kind: row.GetKind()})
		}
	}
	return out, nil
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

type ImportAccountsIn struct {
	// Phones as submitted — the row's own spelling rides to the verdict so
	// the operator's list maps back line by line.
	Rows []string
}

type AccountImportOutcome struct {
	RowIndex int32 // 1-based, the request's own row
	Phone    string
	// Empty unless the row created — the new account's hydration join key.
	UserID string
	Status systempb.AccountImportStatus
	Reason systempb.AccountImportReason
}

type ImportAccountsOut struct {
	Created int32
	Skipped int32
	Invalid int32
	Results []*AccountImportOutcome
}

func (uc *AuthUC) ImportAccounts(ctx context.Context, in *ImportAccountsIn) (*ImportAccountsOut, error) {
	// The operator's bulk form of the sms auto-registration: the same
	// writes LoginByIdentity performs, minus the code proof (the numbers
	// were collected off-line). Rows are independent — a real failure
	// stops the batch with the landed rows still landed.
	row, _, err := uc.providerRowOfKind(smsKind)
	if err != nil {
		return nil, err
	}
	// The seat lands where the sms provider's own logins land, so a later
	// real login finds the seat instead of growing a second one.
	population, err := uc.landingPopulation(ctx, row)
	if err != nil {
		return nil, err
	}
	out := &ImportAccountsOut{Results: make([]*AccountImportOutcome, 0, len(in.Rows))}
	claimed := make(map[string]struct{}, len(in.Rows))
	for i, phone := range in.Rows {
		idx := int32(i + 1)
		canon := normalizePhone(phone)
		if canon == "" {
			// The contract checked shape; the mainland rule is this
			// channel's own law, the same refusal IssueCode gives.
			out.Invalid++
			out.Results = append(out.Results, &AccountImportOutcome{RowIndex: idx, Phone: phone,
				Status: systempb.AccountImportStatus_ACCOUNT_IMPORT_STATUS_INVALID, Reason: systempb.AccountImportReason_ACCOUNT_IMPORT_REASON_NOT_MAINLAND_MOBILE})
			continue
		}
		if _, dup := claimed[canon]; dup {
			out.Invalid++
			out.Results = append(out.Results, &AccountImportOutcome{RowIndex: idx, Phone: phone,
				Status: systempb.AccountImportStatus_ACCOUNT_IMPORT_STATUS_INVALID, Reason: systempb.AccountImportReason_ACCOUNT_IMPORT_REASON_DUPLICATE_IN_FILE})
			continue
		}
		claimed[canon] = struct{}{}
		existing, cerr := uc.userRP.FindCredential(ctx, smsKind, canon)
		switch {
		case cerr == nil:
			// An existing account owns the phone — its history is its own;
			// the row never touches it. The id rides along so a caller
			// with a pre-fill to land can key on it without a second read.
			out.Skipped++
			out.Results = append(out.Results, &AccountImportOutcome{RowIndex: idx, Phone: phone, UserID: existing.UserID,
				Status: systempb.AccountImportStatus_ACCOUNT_IMPORT_STATUS_SKIPPED_DUPLICATE,
				Reason: systempb.AccountImportReason_ACCOUNT_IMPORT_REASON_PHONE_TAKEN})
			continue
		case !errorspb.IsInfraErrorDbNotFound(cerr):
			return nil, cerr
		}
		account := xid.New().String()
		// The auto-registration contract: an empty hash permanently closes
		// the password plane, the phone is the only way in.
		hash := ""
		var userID string
		txErr := uc.Tm.InTx(ctx, func(ctx context.Context) error {
			u, cerr := uc.userRP.Create(ctx, &user.User{
				Account:      &account,
				PasswordHash: &hash,
				Enabled:      true,
			})
			if cerr != nil {
				return cerr
			}
			userID = u.ID
			if cerr = uc.userRP.CreateCredential(ctx, &user.Credential{
				Kind:       smsKind,
				Identifier: canon,
				UserID:     u.ID,
			}); cerr != nil {
				return cerr
			}
			return uc.userRP.EnsureSeat(ctx, &user.Seat{
				UserID:     u.ID,
				TenantID:   security.DefaultTenant,
				Population: population,
			})
		})
		if txErr != nil {
			return nil, txErr
		}
		out.Created++
		out.Results = append(out.Results, &AccountImportOutcome{RowIndex: idx, Phone: phone, UserID: userID,
			Status: systempb.AccountImportStatus_ACCOUNT_IMPORT_STATUS_CREATED})
	}
	return out, nil
}

func (uc *AuthUC) providerRowOfKind(kind string) (*conf.Auth_Provider, IdentityProvider, error) {
	// Anchors on the code-side kind vocabulary (the stable word), not the
	// deploy-side provider key.
	for _, row := range uc.auth.GetProviders() {
		if row.GetKind() == kind && row.GetEnabled() {
			return row, uc.providers[row.GetKind()], nil
		}
	}
	return nil, nil, systempb.ErrorSystemIdentityProviderUnknown("").WithCause(
		fmt.Errorf("identity provider kind %q", kind))
}

func (uc *AuthUC) BindCredential(ctx context.Context, providerKey, state string, payload identityPayload) (*wechatProfile, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("binding requires an authenticated caller"))
	}
	row, p, err := uc.providerRow(providerKey)
	if err != nil {
		return nil, err
	}
	if err = uc.cashBoundState(ctx, providerKey, state, subject.UserID); err != nil {
		return nil, err
	}
	id, err := p.Exchange(ctx, row, payload)
	if err != nil {
		return nil, err
	}
	c, err := uc.userRP.FindCredential(ctx, id.Kind, id.Identifier)
	if err != nil && !errorspb.IsInfraErrorDbNotFound(err) {
		return nil, err
	}
	if err == nil {
		// Already bound to this identity: cosmetic. Bound to another: the
		// identifier belongs to someone else.
		if c.UserID == subject.UserID {
			return id.Profile, nil
		}
		return nil, systempb.ErrorSystemIdentityAlreadyBound("").WithCause(
			fmt.Errorf("identity %s %s already bound", id.Kind, id.Identifier))
	}
	if err = uc.userRP.CreateCredential(ctx, &user.Credential{Kind: id.Kind, Identifier: id.Identifier, UserID: subject.UserID}); err != nil {
		// The unique is the race backstop — same shape as the pre-check.
		if errorspb.IsInfraErrorDbConstraint(err) {
			return nil, systempb.ErrorSystemIdentityAlreadyBound("").WithCause(err)
		}
		return nil, err
	}
	return id.Profile, nil
}

func (uc *AuthUC) BeginBindCredential(ctx context.Context, providerKey, redirectURI string) (string, error) {
	if _, ok := security.SubjectFromCtx(ctx); !ok {
		return "", errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("binding requires an authenticated caller"))
	}
	row, p, err := uc.providerRow(providerKey)
	if err != nil {
		return "", err
	}
	// Only jump-capable kinds have an authorize URL; a direct-code kind
	// (no redirect) refuses the begin flow — its entry is SendBindSmsCode.
	rp, ok := p.(redirectProvider)
	if !ok {
		return "", systempb.ErrorSystemIdentityProviderUnknown("").WithCause(
			fmt.Errorf("provider %q has no redirect flow", providerKey))
	}
	// The authenticated rung puts the caller in ctx, so this mint binds
	// the state to the session — the cash at BindCredential enforces it.
	// Bind flows relay nothing: their landing is the caller's own page.
	state, err := mintIdentityState(ctx, uc.tokenRP, providerKey, "")
	if err != nil {
		return "", err
	}
	return rp.AuthorizeURL(row, redirectURI, state), nil
}

func (uc *AuthUC) UnbindCredential(ctx context.Context, providerKey string) error {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok {
		return errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("unbinding requires an authenticated caller"))
	}
	// Addressing rides the provider registry like every bind leg: the row
	// names the kind whose bound credential this removes.
	row, _, err := uc.providerRow(providerKey)
	if err != nil {
		return err
	}
	creds, err := uc.userRP.ListCredentialsByUser(ctx, subject.UserID)
	if err != nil {
		return err
	}
	var target *user.Credential
	for _, c := range creds {
		if c.Kind == row.GetKind() {
			target = c
			break
		}
	}
	if target == nil {
		return systempb.ErrorSystemIdentityNotBound("").WithCause(
			fmt.Errorf("no %s credential on account", row.GetKind()))
	}
	// The sole-path guard: removal must leave at least one login path —
	// another credential or a set password. The machine owns this gate;
	// a face that forgets it only loses the nicer refusal copy.
	u, err := uc.userRP.FindByID(ctx, subject.UserID)
	if err != nil {
		return err
	}
	if len(creds) == 1 && utils.Deref(u.PasswordHash, "") == "" {
		return systempb.ErrorSystemLastLoginMethod("").WithCause(
			fmt.Errorf("unbinding %s would leave no login path", row.GetKind()))
	}
	// Sessions stay: unbind is not logout — the token chain carries no
	// credential link, and re-binding demands its own fresh proof anyway.
	return uc.userRP.DeleteCredential(ctx, target.Kind, target.Identifier, subject.UserID)
}

func (uc *AuthUC) AdminUnbindCredential(ctx context.Context, userID, providerKey string) error {
	// The operator's twin of the member leg above: the same machine with
	// the target named on the wire instead of read from the session. The
	// access annotation upstream is this rung's whole gate.
	row, _, err := uc.providerRow(providerKey)
	if err != nil {
		return err
	}
	u, err := uc.userRP.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	creds, err := uc.userRP.ListCredentialsByUser(ctx, userID)
	if err != nil {
		return err
	}
	var target *user.Credential
	for _, c := range creds {
		if c.Kind == row.GetKind() {
			target = c
			break
		}
	}
	if target == nil {
		return systempb.ErrorSystemIdentityNotBound("").WithCause(
			fmt.Errorf("no %s credential on account", row.GetKind()))
	}
	// The sole-path guard rides identically — an operator cannot brick an
	// account either; containment for a compromised sole credential is the
	// disable toggle, not removal.
	if len(creds) == 1 && utils.Deref(u.PasswordHash, "") == "" {
		return systempb.ErrorSystemLastLoginMethod("").WithCause(
			fmt.Errorf("unbinding %s would leave user %s no login path", row.GetKind(), userID))
	}
	return uc.userRP.DeleteCredential(ctx, target.Kind, target.Identifier, userID)
}

func (uc *AuthUC) ResetPassword(ctx context.Context, providerKey, state, newPassword string, payload smsPayload) error {
	// One setter for both password-less legs: forgot (login page) and
	// first-set (account security). The sms code to the account's own phone
	// is the whole anti-takeover gate — the code just proved the phone, and
	// the phone names its account.
	if err := uc.cashState(ctx, providerKey, state); err != nil {
		return err
	}
	row, p, err := uc.providerRow(providerKey)
	if err != nil {
		return err
	}
	id, err := p.Exchange(ctx, row, payload)
	if err != nil {
		return err
	}
	c, err := uc.userRP.FindCredential(ctx, id.Kind, id.Identifier)
	if err != nil {
		if errorspb.IsInfraErrorDbNotFound(err) {
			return systempb.ErrorSystemUserNotFound("").WithCause(
				fmt.Errorf("no account holds %s %s", id.Kind, maskPhone(id.Identifier)))
		}
		return err
	}
	if err := uc.PasswordPolicyError(newPassword); err != nil {
		return err
	}
	hash, herr := utils.Hash(newPassword)
	if herr != nil {
		return errorspb.ErrorGeneralErrorInvalidArgument("").WithCause(fmt.Errorf("password rejected: %w", herr))
	}
	// Sessions stay standing, same ruling as ChangePassword: revocation
	// belongs to disable, not to a credential rotation.
	if _, err = uc.userRP.Update(ctx, []string{"password"}, &user.User{ID: c.UserID, PasswordHash: &hash}); err != nil {
		return err
	}
	return nil
}

func (uc *AuthUC) ChangePassword(ctx context.Context, oldPassword, newPassword string) error {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok {
		return errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("change password requires an authenticated caller"))
	}
	u, err := uc.userRP.FindByID(ctx, subject.UserID)
	if err != nil {
		if errorspb.IsInfraErrorDbNotFound(err) {
			return systempb.ErrorSystemUserNotFound("")
		}
		return err
	}
	if !utils.Verify(oldPassword, *u.PasswordHash) {
		return systempb.ErrorSystemPasswordMismatch("").WithCause(fmt.Errorf("old password mismatch: user %s", subject.UserID))
	}
	if err := uc.PasswordPolicyError(newPassword); err != nil {
		return err
	}
	hash, err := utils.Hash(newPassword)
	if err != nil {
		return errorspb.ErrorGeneralErrorInvalidArgument("").WithCause(fmt.Errorf("password rejected: %w", err))
	}
	if _, err = uc.userRP.Update(ctx, []string{"password"}, &user.User{ID: subject.UserID, PasswordHash: &hash}); err != nil {
		return err
	}
	// Existing sessions survive a self-service password change: the caller
	// proved the old password, so this is the same person rotating a
	// credential, not an account takeover — unlike disable/delete, which do
	// revoke (their actor is someone else).
	return nil
}

func (uc *AuthUC) PasswordPolicyError(candidate string) error {
	hasLetter, hasDigit := false, false
	for _, r := range candidate {
		switch {
		case ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z'):
			hasLetter = true
		case '0' <= r && r <= '9':
			hasDigit = true
		}
	}
	// The prototype's own rule enforced server-side: at least six runes
	// carrying both an ASCII letter and an ASCII digit — one law for every
	// seam that accepts a freshly chosen password, and the face the shared
	// PasswordPolicy contract serves for the other modules' birth flows.
	// The hash-only callers (login rehash, auto-register's closed plane,
	// the boot seed) stay outside on purpose: none of them carries a
	// member's fresh choice.
	if utf8.RuneCountInString(candidate) < 6 || !hasLetter || !hasDigit {
		return systempb.ErrorSystemPasswordPolicy("")
	}
	return nil
}

func (uc *AuthUC) GetSecurityStatus(ctx context.Context) (*SecurityStatus, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("security status requires an authenticated caller"))
	}
	u, err := uc.userRP.FindByID(ctx, subject.UserID)
	if err != nil {
		return nil, err
	}
	creds, err := uc.userRP.ListCredentialsByUser(ctx, subject.UserID)
	if err != nil {
		return nil, err
	}
	return &SecurityStatus{
		HasPassword: utils.Deref(u.PasswordHash, "") != "",
		Credentials: utils.SliceMap(creds, func(c *user.Credential) *SecurityCredential {
			return &SecurityCredential{Kind: c.Kind, Identifier: maskIdentifier(c.Kind, c.Identifier)}
		}),
	}, nil
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
	if uc.clock().After(sess.ExpiresAt) {
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
			// request its deep reads, never its correctness. The swap is
			// compare-and-set against the bytes this request read: a logout
			// or revoke deleting the key mid-flight must not be undone by
			// the write-back resurrecting it — a lost race just skips the
			// stamp bump and lets the next request verify deep again.
			sess.LivenessStamp = ver
			swapped, werr := uc.tokenRP.CompareAndSet(ctx, sessionPrefix+token, val, utils.MustMarshal(sess), sessionIdle)
			if werr != nil {
				uc.Log.Warn("auth: liveness stamp write-back failed; next request verifies deep", "error", werr)
			} else if !swapped {
				uc.Log.Debug("auth: liveness stamp write-back lost its race; session changed or died")
			}
		}
	}
	slide := min(sessionIdle, sess.ExpiresAt.Sub(uc.clock()))
	if err := uc.tokenRP.Expire(ctx, sessionPrefix+token, slide); err != nil {
		return nil, err
	}
	plane, err := uc.authzRP.PlaneOf(sess.Population)
	if err != nil {
		return nil, err
	}
	return &security.Subject{UserID: sess.UserID, TenantID: sess.TenantID, Population: sess.Population, Plane: plane, SessionID: token, Extra: map[string]any{security.ExtraAccount: sess.Account}}, nil
}

func (uc *AuthUC) GetCurrentUser(ctx context.Context) (*CurrentUser, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("no subject in context"))
	}
	account, _ := subject.Extra[security.ExtraAccount].(string)
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

func maskIdentifier(kind, identifier string) string {
	// The one masking home: phones keep the 3-****-4 channel form, anything
	// else (the openid family) keeps head and tail around a fixed star
	// block.
	if kind == smsKind {
		return maskPhone(identifier)
	}
	// The shared mask's own shape — head/tail around a fixed star block,
	// full stars under its length floor (7; the openid family is far
	// longer, the one-step difference never shows).
	return utils.MaskPhone(identifier)
}

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
	// Login faces take any live state of their provider: the exchanged
	// code is the proof, and a state minted on a bind rung still belongs
	// to the same person's flow.
	_, err := loadIdentityState(ctx, uc.tokenRP, providerKey, state)
	return err
}

func (uc *AuthUC) cashBoundState(ctx context.Context, providerKey, state, userID string) error {
	val, err := loadIdentityState(ctx, uc.tokenRP, providerKey, state)
	if err != nil {
		return err
	}
	if val.BindUser != userID {
		return systempb.ErrorSystemIdentityExchangeFailed("").WithCause(fmt.Errorf("state not minted for this session"))
	}
	return nil
}

func (uc *AuthUC) resolveByCredential(ctx context.Context, account string) (*user.User, error) {
	// Answers whether a login spelling that missed the account handle still
	// names an account: any bound credential identifier resolves
	// kind-agnostically. Nil, nil = no credential row knows the spelling
	// (or its user vanished — a settled miss); a non-nil error is
	// infrastructure and stays loud.
	cred, err := uc.userRP.FindCredentialByIdentifier(ctx, account)
	if errorspb.IsInfraErrorDbNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	found, ferr := uc.userRP.FindByID(ctx, cred.UserID)
	if errorspb.IsInfraErrorDbNotFound(ferr) {
		return nil, nil
	}
	if ferr != nil {
		return nil, ferr
	}
	return found, nil
}

func (uc *AuthUC) exchangePassword(ctx context.Context, account, password string) (*user.User, error) {
	u, err := uc.userRP.FindByAccount(ctx, account)
	if errorspb.IsInfraErrorDbNotFound(err) {
		// Handle first, credential second: the account handle is not the
		// only spelling that names an account — any bound credential
		// identifier names its own, the same resolution the sms proof
		// already makes ("the phone names its account"). Kind-agnostic on
		// purpose: a credential kind yet to be invented needs no new
		// branch here. An identifier no credential row knows keeps the
		// timing-honest refusal below.
		found, rerr := uc.resolveByCredential(ctx, account)
		if rerr != nil {
			return nil, rerr
		}
		if found != nil {
			u, err = found, nil
		}
		if err != nil {
			utils.Verify(password, dummyDigest)
			return nil, systempb.ErrorSystemLoginFailed("").WithCause(fmt.Errorf("account not found: %s", account))
		}
	} else if err != nil {
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
		return nil, systempb.ErrorSystemLoginSeatChoiceRequired("").WithCause(fmt.Errorf("identity holds %d seats; a choice is required", len(seats)))
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
	issuedAt = uc.clock()
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
