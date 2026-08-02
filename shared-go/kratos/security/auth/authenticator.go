package auth

import (
	"context"

	"github.com/go-kratos/kratos/v3/errors"

	"cyber-ecosystem/shared-go/kratos/security"
)

var ErrInvalidToken = errors.Unauthorized("INVALID_TOKEN", "")

type Authenticator interface {
	Authenticate(ctx context.Context, token string) (*security.Subject, error)
}
