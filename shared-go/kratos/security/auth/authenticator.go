package auth

import (
	"context"

	"cyber-ecosystem/shared-go/kratos/security"
)

type Authenticator interface {
	Authenticate(ctx context.Context, token string) (*security.Subject, error)
	CheckSession(ctx context.Context, sid string) error
}
