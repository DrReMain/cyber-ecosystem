package authz

import (
	"context"
	"time"

	"cyber-ecosystem/shared-go/kratos/security"
	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"
)

type timeSource struct{}

func (timeSource) Keys() []string {
	return []string{kauthz.AttrTimeNow}
}

func (timeSource) Resolve(_ context.Context, _ *security.Subject) (map[string]any, error) {
	return map[string]any{kauthz.AttrTimeNow: time.Now()}, nil
}
