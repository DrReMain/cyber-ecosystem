package auth

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"cyber-ecosystem/shared-go/kratos/security"
	"cyber-ecosystem/shared-go/utils"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/module/authz"
	"cyber-ecosystem/app/services/system/internal/module/user"
	"cyber-ecosystem/app/services/system/internal/shared"
)

const (
	sessionPrefix = "auth:web:session:"
	sessionIdle   = 7 * 24 * time.Hour
	sessionMax    = 30 * 24 * time.Hour
)

// DO ------------------------------------------------------------------------------------------------------------------

type Session struct {
	UserID    string
	TenantID  string
	ExpiresAt time.Time
	IssuedAt  time.Time
	Email     string
}

type CurrentUser struct {
	ID          string
	Email       string
	Avatar      *string
	Permissions []string
}

// Port ----------------------------------------------------------------------------------------------------------------

type TokenRP interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
	Del(ctx context.Context, key string) error
	Expire(ctx context.Context, key string, ttl time.Duration) error
}

// UC ------------------------------------------------------------------------------------------------------------------

type AuthUC struct {
	shared.UC
	userRP  user.UserRP
	authzRP authz.AuthzRP
	tokenRP TokenRP
}

func NewAuthUC(logger *slog.Logger, tm shared.Transaction, userRP user.UserRP, authzRP authz.AuthzRP, tokenRP TokenRP) *AuthUC {
	return &AuthUC{
		UC:      shared.NewUC(logger.With("module", "module/auth"), tm),
		userRP:  userRP,
		authzRP: authzRP,
		tokenRP: tokenRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *AuthUC) Login(ctx context.Context, email, password string) (token string, issuedAt, expiresAt time.Time, err error) {
	u, err := uc.userRP.FindByEmail(ctx, email)
	if err != nil {
		if errorspb.IsInfraErrorDbNotFound(err) {
			return "", time.Time{}, time.Time{}, systempb.ErrorSystemLoginFailed("").WithCause(fmt.Errorf("user not found: %s", email))
		}
		return "", time.Time{}, time.Time{}, err
	}
	if !utils.Verify(password, *u.PasswordHash) {
		return "", time.Time{}, time.Time{}, systempb.ErrorSystemLoginFailed("").WithCause(fmt.Errorf("password mismatch: %s", email))
	}
	if !u.Enabled {
		return "", time.Time{}, time.Time{}, systempb.ErrorSystemUserDisabled("").WithCause(fmt.Errorf("user disabled: %s", email))
	}
	token, err = utils.RandomToken()
	if err != nil {
		return "", time.Time{}, time.Time{}, err
	}
	issuedAt = time.Now()
	expiresAt = issuedAt.Add(sessionMax)
	sess := &Session{UserID: u.ID, TenantID: u.TenantID, ExpiresAt: expiresAt, IssuedAt: issuedAt, Email: utils.Deref(u.Email, "")}
	if err = uc.tokenRP.Set(ctx, sessionPrefix+token, utils.MustMarshal(sess), sessionIdle); err != nil {
		return "", time.Time{}, time.Time{}, err
	}
	return token, issuedAt, expiresAt, nil
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
	slide := min(sessionIdle, time.Until(sess.ExpiresAt))
	if err := uc.tokenRP.Expire(ctx, sessionPrefix+token, slide); err != nil {
		return nil, err
	}
	return &security.Subject{UserID: sess.UserID, TenantID: sess.TenantID, SessionID: token, Extra: map[string]any{shared.ExtraEmail: sess.Email}}, nil
}

func (uc *AuthUC) GetCurrentUser(ctx context.Context) (*CurrentUser, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("no subject in context"))
	}
	email, _ := subject.Extra[shared.ExtraEmail].(string)
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
	return &CurrentUser{ID: subject.UserID, Email: email, Avatar: avatar, Permissions: perms}, nil
}
