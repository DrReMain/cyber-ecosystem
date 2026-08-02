package biz

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"cyber-ecosystem/shared-go/kratos/security"
	"cyber-ecosystem/shared-go/kratos/security/auth"
	"cyber-ecosystem/shared-go/utils"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
)

const (
	prefix = "auth:token:"
	ttl    = 15 * time.Minute
)

// DO ------------------------------------------------------------------------------------------------------------------

// Port ----------------------------------------------------------------------------------------------------------------

type TokenRP interface {
	Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
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

func (uc *AuthUC) Login(ctx context.Context, email, password string) (string, error) {
	u, err := uc.userRP.FindByEmail(ctx, email)
	if err != nil {
		if errorspb.IsInfraErrorDbNotFound(err) {
			return "", errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(errors.New("user not exist"))
		}
		return "", err
	}
	if !utils.Verify(password, u.PasswordHash) {
		return "", errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(errors.New("password is wrong"))
	}
	return uc.issueToken(ctx, u.ID, u.TenantID)
}

func (uc *AuthUC) Authenticate(ctx context.Context, token string) (*security.Subject, error) {
	data, err := uc.tokenRP.Get(ctx, prefix+token)
	if err != nil {
		return nil, err
	}
	var s security.Subject
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func (uc *AuthUC) issueToken(ctx context.Context, userID, tenantID string) (string, error) {
	tok, err := auth.GenerateToken()
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(&security.Subject{UserID: userID, TenantID: tenantID})
	if err != nil {
		return "", err
	}
	if err := uc.tokenRP.Set(ctx, prefix+tok, data, ttl); err != nil {
		return "", err
	}
	return tok, nil
}
