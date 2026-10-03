package authz

import (
	"context"
	"log/slog"
	"slices"

	"cyber-ecosystem/shared-go/kratos/security"
	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"

	"cyber-ecosystem/app/services/agent/internal/shared"
)

// Port ----------------------------------------------------------------------------------------------------------------

type AuthzRP interface {
	CheckGrants(ctx context.Context, subject *security.Subject, operation string) (*kauthz.Decision, error)
}

// UC ------------------------------------------------------------------------------------------------------------------

type AuthzUC struct {
	shared.UC
	authzRP AuthzRP
}

func NewAuthzUC(logger *slog.Logger, tm shared.Transaction, authzRP AuthzRP) *AuthzUC {
	return &AuthzUC{
		UC:      shared.NewUC(logger.With("module", "module/authz"), tm),
		authzRP: authzRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *AuthzUC) Decide(ctx context.Context, subject *security.Subject, operation string) (*kauthz.Decision, error) {
	if slices.Contains(builtinOperations, operation) {
		return &kauthz.Decision{Allowed: true, Reason: kauthz.AllowBuiltin}, nil
	}
	return uc.authzRP.CheckGrants(ctx, subject, operation)
}
