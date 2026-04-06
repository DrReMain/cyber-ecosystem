package auth

import (
	"context"
	"errors"

	jwtv5 "github.com/golang-jwt/jwt/v5"

	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"

	singletonV1 "cyber-ecosystem/apps/singleton/gen/go/v1"
)

var (
	ErrSessionRevoked     = singletonV1.ErrorErrorReasonSessionRevoked("")
	ErrSessionUnavailable = singletonV1.ErrorErrorReasonUnspecified("")
	ErrGetClaim           = singletonV1.ErrorErrorReasonUnspecified("")
)

type Identity struct {
	jwtv5.RegisteredClaims
	Sid string `json:"sid,omitempty"`
}

func IdentityFromContext(ctx context.Context) (*Identity, error) {
	raw, ok := jwt.FromContext(ctx)
	if !ok {
		return nil, ErrGetClaim.WithCause(errors.New("missing auth claims"))
	}
	claims, ok := raw.(*Identity)
	if !ok {
		return nil, ErrGetClaim.WithCause(errors.New("invalid claims type"))
	}
	return claims, nil
}
