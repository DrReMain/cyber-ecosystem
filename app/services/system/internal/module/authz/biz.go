package authz

import (
	"context"
	"log/slog"
	"slices"
	"strings"

	"cyber-ecosystem/shared-go/kratos/security"
	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"

	"cyber-ecosystem/app/services/system/internal/shared"
)

// DO ------------------------------------------------------------------------------------------------------------------

type GrantView struct {
	Pattern     string
	ScopeKind   string // "" = tenant-wide (no narrowing)
	ScopeParams map[string]any
	Policies    []PolicyView
}

type PolicyView struct {
	Kind   string
	Name   string
	Params map[string]any
}

type ExplainResult struct {
	Allowed   bool
	Builtin   bool
	RoleCodes []string
	Hits      []GrantHit
}

type GrantHit struct {
	Pattern   string
	RoleCode  string
	ScopeKind string
	Policies  []PolicyState
}

type PolicyState struct {
	Kind  string
	Name  string
	State string // passed | failed; indeterminate reserved for source-type attrs
}

// Port ----------------------------------------------------------------------------------------------------------------

type AuthzRP interface {
	ImplicitGrants(ctx context.Context, userID, tenant string) ([]GrantView, error)
	EffectiveGrants(ctx context.Context, subject *security.Subject) ([]GrantView, error)
	ExplainGrants(ctx context.Context, principalType, principalID, tenant, operation string) (*ExplainResult, error)
	PreviewGrants(ctx context.Context, tenant string, roleCodes []string) ([]string, error)
	StartWatch(ctx context.Context) error
	StopWatch(ctx context.Context) error
}

// UC ------------------------------------------------------------------------------------------------------------------

type AuthzUC struct {
	shared.UC
	authzRP AuthzRP
}

func NewAuthzUC(logger *slog.Logger, tm shared.Transaction, authzRP AuthzRP, lc shared.HookRegistry) *AuthzUC {
	lc.OnStart(authzRP.StartWatch)
	lc.OnStop(authzRP.StopWatch)
	return &AuthzUC{UC: shared.NewUC(logger.With("module", "module/authz"), tm), authzRP: authzRP}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *AuthzUC) Decide(ctx context.Context, subject *security.Subject, operation string) (*kauthz.Decision, error) {
	if slices.Contains(builtinOperations, operation) {
		return &kauthz.Decision{Allowed: true, Reason: kauthz.AllowBuiltin}, nil
	}
	grants, err := uc.authzRP.ImplicitGrants(ctx, subject.UserID, subject.TenantID)
	if err != nil {
		// Engine failure is infra, not a domain rule — fuzzy identity keeps
		// the telemetry bucket distinct while the cause stays in-process.
		return nil, errorspb.ErrorGeneralErrorInternal("").WithCause(err)
	}
	matched := make([]GrantView, 0, len(grants))
	for _, g := range grants {
		if matchOperation(operation, g.Pattern) {
			matched = append(matched, g)
		}
	}
	if len(matched) == 0 {
		return &kauthz.Decision{Reason: kauthz.DenyNoGrant}, nil
	}
	surviving := survivingGrants(ctx, uc.Log, subject, matched)
	if len(surviving) == 0 {
		return &kauthz.Decision{Reason: kauthz.DenyAbacConstraint}, nil
	}
	scopes := make([]kauthz.ScopeDescriptor, 0, len(surviving))
	for _, g := range surviving {
		scopes = append(scopes, kauthz.ScopeDescriptor{Kind: g.ScopeKind, Params: g.ScopeParams})
	}
	return &kauthz.Decision{Allowed: true, Scopes: scopes}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func matchOperation(op, pattern string) bool {
	// Mirrors the client-side permission semantics verbatim; a non-suffix
	// "*" is not a wildcard (save-time validation rejects it, this stays
	// honest if one slips through).
	if !strings.HasSuffix(pattern, "/*") {
		return op == pattern
	}
	return strings.HasPrefix(op, strings.TrimSuffix(pattern, "*"))
}
