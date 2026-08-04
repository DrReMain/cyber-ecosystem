package biz

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"cyber-ecosystem/shared-go/kratos/security"
	"cyber-ecosystem/shared-go/kratos/security/auth"
	"cyber-ecosystem/shared-go/utils"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
)

const (
	accessPrefix  = "auth:access:"
	refreshPrefix = "auth:refresh:"
	sessionPrefix = "auth:session:"

	accessTTL  = 15 * time.Minute
	refreshTTL = 7 * 24 * time.Hour
	sessionTTL = refreshTTL // a session dies with its longest-lived token (the refresh)
)

// DO ------------------------------------------------------------------------------------------------------------------

type Session struct {
	UserID   string
	TenantID string
}

type TokenPair struct {
	Access  string
	Refresh string
}

// Port ----------------------------------------------------------------------------------------------------------------

type TokenRP interface {
	Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
	Del(ctx context.Context, key string) error
}

// UC ------------------------------------------------------------------------------------------------------------------

type AuthUC struct {
	UC
	userRP  UserRP
	tokenRP TokenRP
}

func NewAuthUC(logger *slog.Logger, tm Transaction, userRP UserRP, tokenRP TokenRP) *AuthUC {
	return &AuthUC{
		UC:      UC{log: logger.With("module", "biz/auth"), tm: tm},
		userRP:  userRP,
		tokenRP: tokenRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *AuthUC) Login(ctx context.Context, email, password string) (*TokenPair, error) {
	u, err := uc.userRP.FindByEmail(ctx, email)
	if err != nil {
		if errorspb.IsInfraErrorDbNotFound(err) {
			return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(errors.New("user not exist"))
		}
		return nil, err
	}
	if !utils.Verify(password, u.PasswordHash) {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(errors.New("password is wrong"))
	}

	sid, err := auth.GenerateToken()
	if err != nil {
		return nil, err
	}
	refresh, err := auth.GenerateToken()
	if err != nil {
		return nil, err
	}
	if err := uc.tokenRP.Set(ctx, sessionPrefix+sid, utils.MustMarshal(&Session{UserID: u.ID, TenantID: u.TenantID}), sessionTTL); err != nil {
		return nil, err
	}
	access, err := auth.GenerateToken()
	if err != nil {
		return nil, err
	}
	if err := uc.tokenRP.Set(ctx, accessPrefix+access, utils.MustMarshal(&security.Subject{UserID: u.ID, TenantID: u.TenantID, SessionID: sid}), accessTTL); err != nil {
		return nil, err
	}
	if err := uc.tokenRP.Set(ctx, refreshPrefix+refresh, []byte(sid), refreshTTL); err != nil {
		return nil, err
	}
	return &TokenPair{Access: access, Refresh: refresh}, nil
}

func (uc *AuthUC) Authenticate(ctx context.Context, token string) (*security.Subject, error) {
	data, err := uc.tokenRP.Get(ctx, accessPrefix+token)
	if err != nil {
		return nil, err
	}
	s, err := utils.Unmarshal[security.Subject](data)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (uc *AuthUC) CheckSession(ctx context.Context, sid string) error {
	_, err := uc.tokenRP.Get(ctx, sessionPrefix+sid)
	return err
}

func (uc *AuthUC) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	sidBytes, err := uc.tokenRP.Get(ctx, refreshPrefix+refreshToken)
	if err != nil {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(errors.New("invalid refresh token"))
	}
	sid := string(sidBytes)

	sessBytes, err := uc.tokenRP.Get(ctx, sessionPrefix+sid)
	if err != nil {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(errors.New("session expired"))
	}
	sess, err := utils.Unmarshal[Session](sessBytes)
	if err != nil {
		return nil, err
	}
	// rotation: issue a new access + refresh bound to the same session, then invalidate the old refresh
	access, err := auth.GenerateToken()
	if err != nil {
		return nil, err
	}
	newRefresh, err := auth.GenerateToken()
	if err != nil {
		return nil, err
	}
	if err := uc.tokenRP.Set(ctx, accessPrefix+access, utils.MustMarshal(&security.Subject{UserID: sess.UserID, TenantID: sess.TenantID, SessionID: sid}), accessTTL); err != nil {
		return nil, err
	}
	if err := uc.tokenRP.Set(ctx, refreshPrefix+newRefresh, []byte(sid), refreshTTL); err != nil {
		return nil, err
	}
	if err := uc.tokenRP.Del(ctx, refreshPrefix+refreshToken); err != nil {
		return nil, err
	}
	return &TokenPair{Access: access, Refresh: newRefresh}, nil
}

func (uc *AuthUC) Logout(ctx context.Context) error {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok {
		return errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(errors.New("no subject in context"))
	}
	return uc.tokenRP.Del(ctx, sessionPrefix+subject.SessionID)
}

// Private --------------------------------------------------------------------------------------------------------------
