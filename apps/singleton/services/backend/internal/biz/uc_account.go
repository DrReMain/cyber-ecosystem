package biz

import (
	"context"
	"errors"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/rs/xid"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/shared-go/utils"

	singletonV1 "cyber-ecosystem/apps/singleton/gen/go/v1"
	"cyber-ecosystem/apps/singleton/services/backend/internal/conf"
	"cyber-ecosystem/apps/singleton/services/backend/internal/pkg/auth"
)

var (
	ErrPasswordWrong = singletonV1.ErrorErrorReasonPasswordWrong("")
	ErrTokenInvalid  = singletonV1.ErrorErrorReasonUnauthorized("")
	ErrSignFailed    = singletonV1.ErrorErrorReasonUnspecified("")
)

// Model ---------------------------------------------------------------------------------------------------------------

type TokenSubject struct {
	Sub string
}

type Token struct {
	Value    string
	ExpireAt time.Time
}

type TokenPair struct {
	AccessToken  Token
	RefreshToken Token
}

type AccountProfile struct {
	ID    *string
	Email *string
}

// Port ----------------------------------------------------------------------------------------------------------------

type SessionRP interface {
	RevokeSession(ctx context.Context, sid string, ttl time.Duration) error
	IsSessionRevoked(ctx context.Context, sid string) (bool, error)
}

// UC ------------------------------------------------------------------------------------------------------------------

type AccountUC struct {
	UC
	signingMethod jwtv5.SigningMethod
	ca            *conf.Auth
	sessionRP     SessionRP
	userRP        UserRP
}

func NewAccountUC(logger log.Logger, tm Transaction, ca *conf.Auth, sessionRP SessionRP, userRP UserRP) *AccountUC {
	return &AccountUC{
		UC: UC{
			log: log.NewHelper(log.With(logger, "module", "biz/uc_account")),
			tm:  tm,
		},
		signingMethod: auth.SigningMethod,
		ca:            ca,
		sessionRP:     sessionRP,
		userRP:        userRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *AccountUC) SigningMethod() jwtv5.SigningMethod {
	return uc.signingMethod
}

func (uc *AccountUC) LoginPassword(ctx context.Context, email, password string) (*TokenPair, error) {
	u, err := uc.userRP.FindByEmail(ctx, email)
	if err != nil {
		if singletonV1.IsErrorReasonEntNotFound(err) {
			return nil, ErrPasswordWrong.WithCause(err)
		}
		return nil, err
	}
	if u.PasswordCipher == nil || !utils.EncryptCheck(password, *u.PasswordCipher) {
		return nil, ErrPasswordWrong.WithCause(errors.New("password check failed"))
	}

	tp, err := uc.generateTokenPair(TokenSubject{Sub: *u.ID})
	if err != nil {
		return nil, err
	}
	return tp, nil
}

func (uc *AccountUC) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	claims, err := uc.parseToken(refreshToken)
	if err != nil {
		return nil, err
	}

	revoked, err := uc.sessionRP.IsSessionRevoked(ctx, claims.Sid)
	if err != nil {
		return nil, auth.ErrSessionUnavailable.WithCause(err)
	}
	if revoked {
		return nil, auth.ErrSessionRevoked.WithCause(errors.New("session is revoked"))
	}

	remaining := time.Until(claims.ExpiresAt.Time)
	if remaining > 0 {
		if err := uc.sessionRP.RevokeSession(ctx, claims.Sid, remaining); err != nil {
			return nil, auth.ErrSessionUnavailable.WithCause(err)
		}
	}

	return uc.generateTokenPair(TokenSubject{Sub: claims.Subject})
}

func (uc *AccountUC) Logout(ctx context.Context, claims *auth.Identity) error {
	refreshTTL := uc.ca.GetToken().GetRefreshTtl().AsDuration()
	sessionExpiry := claims.IssuedAt.Time.Add(refreshTTL)
	ttl := time.Until(sessionExpiry)
	if ttl > 0 {
		return uc.sessionRP.RevokeSession(ctx, claims.Sid, ttl)
	}
	return nil
}

func (uc *AccountUC) GetProfile(ctx context.Context, userID string) (*AccountProfile, error) {
	u, err := uc.userRP.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &AccountProfile{
		ID:    u.ID,
		Email: u.Email,
	}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func (uc *AccountUC) generateTokenPair(sub TokenSubject) (*TokenPair, error) {
	now := time.Now()
	sid := xid.New().String()

	var generateToken = func(ttl time.Duration) (*Token, error) {
		claims := &auth.Identity{
			RegisteredClaims: jwtv5.RegisteredClaims{
				ID:        xid.New().String(),
				Subject:   sub.Sub,
				IssuedAt:  jwtv5.NewNumericDate(now),
				ExpiresAt: jwtv5.NewNumericDate(now.Add(ttl)),
			},
			Sid: sid,
		}
		token := jwtv5.NewWithClaims(uc.signingMethod, claims)
		signed, err := token.SignedString([]byte(uc.ca.GetSecret()))
		if err != nil {
			return nil, ErrSignFailed.WithCause(err)
		}
		return &Token{
			Value:    signed,
			ExpireAt: now.Add(ttl),
		}, nil
	}

	accessToken, err := generateToken(uc.ca.GetToken().GetAccessTtl().AsDuration())
	if err != nil {
		return nil, err
	}

	refreshToken, err := generateToken(uc.ca.GetToken().GetRefreshTtl().AsDuration())
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  *accessToken,
		RefreshToken: *refreshToken,
	}, nil
}

func (uc *AccountUC) parseToken(raw string) (*auth.Identity, error) {
	token, err := jwtv5.ParseWithClaims(raw, &auth.Identity{}, func(t *jwtv5.Token) (any, error) {
		if _, ok := t.Method.(*jwtv5.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalid.WithCause(errors.New("unexpected signing method"))
		}
		return []byte(uc.ca.GetSecret()), nil
	})
	if err != nil {
		return nil, ErrTokenInvalid.WithCause(err)
	}
	claims, ok := token.Claims.(*auth.Identity)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid.WithCause(errors.New("invalid token claims"))
	}
	return claims, nil
}
