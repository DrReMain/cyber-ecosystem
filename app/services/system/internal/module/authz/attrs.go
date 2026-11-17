package authz

import (
	"context"

	"cyber-ecosystem/shared-go/kratos/security"
	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"

	"cyber-ecosystem/app/services/system/internal/shared"
)

type timeSource struct {
	clock shared.Clock // evaluation-time seam: the owning UC wires time.Now, a test freezes it
}

func (timeSource) Keys() []string {
	return []string{kauthz.AttrTimeNow}
}

func (ts timeSource) Resolve(_ context.Context, _ *security.Subject) (map[string]any, error) {
	return map[string]any{kauthz.AttrTimeNow: ts.clock()}, nil
}
