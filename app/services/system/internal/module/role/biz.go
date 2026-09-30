package role

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"cyber-ecosystem/shared-go/kratos/security"
	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"

	commonpb "cyber-ecosystem/gen/go/cyber/shared/common/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/module/authz"
	"cyber-ecosystem/app/services/system/internal/module/policy"
	"cyber-ecosystem/app/services/system/internal/module/resource"
	"cyber-ecosystem/app/services/system/internal/shared"
)

// DO ------------------------------------------------------------------------------------------------------------------

type Role struct {
	ID            string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	TenantID      string
	Code          string
	Name          string
	Enabled       bool
	Remark        *string
	GrantCount    int64
	BoundCount    int64
	GlobalGrant   bool
	WildcardCount int64
	NarrowedCount int64
	PolicyCount   int64
	Grants        []*Permission
}

type Permission struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	TenantID  string
	RoleID    string
	Operation string
	Effect    string
	ScopeKind string
	PolicyIDs []string
}

type RoleMember struct {
	PrincipalType string
	PrincipalID   string
	CreatedAt     time.Time
}

type RoleListIn struct {
	*commonpb.PageRequest
	OrderBy []string
	Code    *string
	Name    *string
}

type RoleListOut struct {
	*commonpb.PageResponse
	List []*Role
}

type MemberListIn struct {
	RoleID string
	*commonpb.PageRequest
	PrincipalType *string
}

type MemberListOut struct {
	*commonpb.PageResponse
	List []*RoleMember
}

// Port ----------------------------------------------------------------------------------------------------------------

type RoleRP interface {
	Create(ctx context.Context, r *Role) (*Role, error)
	Update(ctx context.Context, fieldsMask []string, r *Role) (*Role, error)
	UpdateStatus(ctx context.Context, id string, enabled bool) (*Role, error)
	Delete(ctx context.Context, id string) (string, error)
	List(ctx context.Context, in *RoleListIn) (*RoleListOut, error)
	FindByID(ctx context.Context, id string) (*Role, error)
	ExistsBindings(ctx context.Context, roleID string) (bool, error)
	CountBindings(ctx context.Context, roleID string) (int64, error)
	ListByPrincipal(ctx context.Context, principalType, principalID string) ([]*Role, error)
	ReplaceGrants(ctx context.Context, roleID string, grants []*Permission) error
	ListGrants(ctx context.Context, roleID string) ([]*Permission, error)
	RemoveMember(ctx context.Context, roleID, principalType, principalID string) error
	ListMembers(ctx context.Context, in *MemberListIn) (*MemberListOut, error)
	NotifyChanged(ctx context.Context)
}

// UC ------------------------------------------------------------------------------------------------------------------

type RoleUC struct {
	shared.UC
	roleRP     RoleRP
	resourceRP resource.ResourceRP
	authzRP    authz.AuthzRP
	policyRP   policy.PolicyRP
}

func NewRoleUC(logger *slog.Logger, tm shared.Transaction, roleRP RoleRP, resourceRP resource.ResourceRP, authzRP authz.AuthzRP, policyRP policy.PolicyRP) *RoleUC {
	return &RoleUC{
		UC:         shared.NewUC(logger.With("module", "module/role"), tm),
		roleRP:     roleRP,
		resourceRP: resourceRP,
		authzRP:    authzRP,
		policyRP:   policyRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *RoleUC) Create(ctx context.Context, r *Role, grants []*Permission) (out *Role, err error) {
	if err = uc.validateGrants(ctx, grants); err != nil {
		return nil, err
	}
	err = uc.Tm.InTx(ctx, func(ctx context.Context) error {
		if out, err = uc.roleRP.Create(ctx, r); err != nil {
			return err
		}
		return uc.roleRP.ReplaceGrants(ctx, out.ID, grants)
	})
	if err == nil {
		uc.roleRP.NotifyChanged(ctx)
	}
	return
}

func (uc *RoleUC) Update(ctx context.Context, fieldsMask []string, r *Role, grants []*Permission) (out *Role, err error) {
	if err = uc.ensureRole(ctx, r.ID); err != nil {
		return nil, err
	}
	replaceGrants := slices.Contains(fieldsMask, "grants")
	if replaceGrants {
		if err = uc.validateGrants(ctx, grants); err != nil {
			return nil, err
		}
	}
	err = uc.Tm.InTx(ctx, func(ctx context.Context) error {
		if out, err = uc.roleRP.Update(ctx, fieldsMask, r); err != nil {
			return err
		}
		if replaceGrants {
			return uc.roleRP.ReplaceGrants(ctx, r.ID, grants)
		}
		return nil
	})
	if err == nil && replaceGrants {
		// Only grants feed the compiled snapshot (code/enabled aside);
		// cosmetic edits skip the rebuild broadcast.
		uc.roleRP.NotifyChanged(ctx)
	}
	return
}

func (uc *RoleUC) UpdateStatus(ctx context.Context, id string, enabled bool) (out *Role, err error) {
	if err = uc.ensureRole(ctx, id); err != nil {
		return nil, err
	}
	err = uc.Tm.InTx(ctx, func(ctx context.Context) error {
		out, err = uc.roleRP.UpdateStatus(ctx, id, enabled)
		return err
	})
	if err != nil {
		return nil, err
	}
	// A disabled role loses its policy rows on the next compile; every member
	// is denied on their next request after replicas rebuild.
	uc.roleRP.NotifyChanged(ctx)
	return out, nil
}

func (uc *RoleUC) Delete(ctx context.Context, id string) (string, error) {
	var out string
	err := uc.Tm.InTx(ctx, func(ctx context.Context) error {
		if exists, err := uc.roleRP.ExistsBindings(ctx, id); err != nil {
			return err
		} else if exists {
			return systempb.ErrorSystemRoleHasBindings("")
		}
		deleted, derr := uc.roleRP.Delete(ctx, id)
		out = deleted
		return derr
	})
	if err != nil {
		return "", err
	}
	uc.roleRP.NotifyChanged(ctx)
	return out, nil
}

func (uc *RoleUC) List(ctx context.Context, in *RoleListIn) (*RoleListOut, error) {
	return uc.roleRP.List(ctx, in)
}

func (uc *RoleUC) Get(ctx context.Context, id string) (*Role, error) {
	out, err := uc.roleRP.FindByID(ctx, id)
	if err != nil {
		if errorspb.IsInfraErrorDbNotFound(err) {
			return nil, systempb.ErrorSystemRoleNotFound("")
		}
		return nil, err
	}
	if out.Grants, err = uc.roleRP.ListGrants(ctx, id); err != nil {
		return nil, err
	}
	// Same light material as the list row, so detail and list never disagree.
	applyGrantCounts(out, out.Grants)
	if out.BoundCount, err = uc.roleRP.CountBindings(ctx, id); err != nil {
		return nil, err
	}
	return out, nil
}

func (uc *RoleUC) ListByPrincipal(ctx context.Context, principalType, principalID string) ([]*Role, error) {
	return uc.roleRP.ListByPrincipal(ctx, principalType, principalID)
}

func (uc *RoleUC) RemoveMember(ctx context.Context, roleID, principalType, principalID string) error {
	if err := uc.ensureRole(ctx, roleID); err != nil {
		return err
	}
	err := uc.Tm.InTx(ctx, func(ctx context.Context) error {
		return uc.roleRP.RemoveMember(ctx, roleID, principalType, principalID)
	})
	if err == nil {
		uc.roleRP.NotifyChanged(ctx)
	}
	return err
}

func (uc *RoleUC) ListMembers(ctx context.Context, in *MemberListIn) (*MemberListOut, error) {
	if err := uc.ensureRole(ctx, in.RoleID); err != nil {
		return nil, err
	}
	return uc.roleRP.ListMembers(ctx, in)
}

func (uc *RoleUC) ExplainOperation(ctx context.Context, principalType, principalID, operation string) (*authz.ExplainResult, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("no subject in context"))
	}
	// The caller scopes the diagnosis to their own tenant.
	out, err := uc.authzRP.ExplainGrants(ctx, principalType, principalID, subject.TenantID, operation)
	if err != nil {
		// Engine failure is infra, not a diagnosis outcome — mirror Decide's
		// fuzzy identity while the cause stays in-process.
		return nil, errorspb.ErrorGeneralErrorInternal("").WithCause(err)
	}
	return out, nil
}

func (uc *RoleUC) PreviewGrants(ctx context.Context, roleCodes []string) ([]string, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("no subject in context"))
	}
	ops, err := uc.authzRP.PreviewGrants(ctx, subject.TenantID, roleCodes)
	if err != nil {
		return nil, errorspb.ErrorGeneralErrorInternal("").WithCause(err)
	}
	return ops, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func (uc *RoleUC) ensureRole(ctx context.Context, id string) error {
	if _, err := uc.roleRP.FindByID(ctx, id); err != nil {
		if errorspb.IsInfraErrorDbNotFound(err) {
			return systempb.ErrorSystemRoleNotFound("")
		}
		return err
	}
	return nil
}

func (uc *RoleUC) validateGrants(ctx context.Context, grants []*Permission) error {
	seen := make(map[string]struct{}, len(grants))
	policyIDs := make(map[string]struct{})
	for _, g := range grants {
		if _, dup := seen[g.Operation]; dup {
			return systempb.ErrorSystemPermissionInvalidOperation("").WithCause(fmt.Errorf("duplicate operation %q in grants", g.Operation))
		}
		seen[g.Operation] = struct{}{}
		if err := uc.validateOperation(ctx, g.Operation); err != nil {
			return err
		}
		for _, pid := range g.PolicyIDs {
			policyIDs[pid] = struct{}{}
		}
	}
	return uc.ensurePolicies(ctx, policyIDs)
}

func (uc *RoleUC) ensurePolicies(ctx context.Context, ids map[string]struct{}) error {
	if len(ids) == 0 {
		return nil
	}
	want := make([]string, 0, len(ids))
	for pid := range ids {
		want = append(want, pid)
	}
	found, err := uc.policyRP.FindByIDs(ctx, want)
	if err != nil {
		return err
	}
	if len(found) == len(want) {
		return nil
	}
	got := make(map[string]struct{}, len(found))
	for _, p := range found {
		got[p.ID] = struct{}{}
	}
	missing := make([]string, 0, len(want)-len(found))
	for _, pid := range want {
		if _, ok := got[pid]; !ok {
			missing = append(missing, pid)
		}
	}
	// Absence covers both deleted rows and other-tenant rows — the mixin
	// scopes the lookup, and callers get one honest not-found either way.
	return systempb.ErrorSystemPolicyNotFound("").WithCause(
		fmt.Errorf("policies not found: %v", missing))
}

func (uc *RoleUC) validateOperation(ctx context.Context, pattern string) error {
	svc, method, exact, ok := parseOperationPattern(pattern)
	if !ok {
		return systempb.ErrorSystemPermissionInvalidOperation("").WithCause(fmt.Errorf("malformed pattern %q", pattern))
	}
	if svc == "" {
		return nil // the global "/*" bypasses the catalog by definition
	}
	catalog, err := uc.resourceRP.ListResource(ctx)
	if err != nil {
		return err
	}
	for _, s := range catalog {
		if s.FullName != svc {
			continue
		}
		if !exact {
			return nil
		}
		for _, m := range s.Methods {
			if m.Name == method {
				return nil
			}
		}
		return systempb.ErrorSystemPermissionInvalidOperation("").WithCause(fmt.Errorf("method %q not found on %s", method, svc))
	}
	return systempb.ErrorSystemPermissionInvalidOperation("").WithCause(fmt.Errorf("service %q not in resource catalog", svc))
}

func parseOperationPattern(pattern string) (svc, method string, exact, ok bool) {
	if pattern == "/*" {
		return "", "", false, true
	}
	if !strings.HasPrefix(pattern, "/") {
		return "", "", false, false
	}
	body := pattern[1:]
	switch {
	case strings.HasSuffix(body, "/*"):
		svc = strings.TrimSuffix(body, "/*")
		if svc == "" || !strings.Contains(svc, ".") || strings.ContainsAny(svc, "*/") {
			return "", "", false, false
		}
		return svc, "", false, true
	case strings.Contains(body, "*"):
		return "", "", false, false // a "*" that is not a clean tail is not a wildcard
	default:
		svc, method, found := strings.Cut(body, "/")
		if !found || method == "" || !strings.Contains(svc, ".") || strings.ContainsAny(svc+" "+method, "*/") {
			return "", "", false, false
		}
		return svc, method, true, true
	}
}

func isGlobalPattern(op string) bool { return op == "/*" }

func isWildcardPattern(op string) bool { return strings.HasSuffix(op, "/*") }

func isNarrowingScope(kind string) bool {
	return kind == kauthz.ScopeKindSelf || kind == kauthz.ScopeKindDeptTree
}

func applyGrantCounts(r *Role, grants []*Permission) {
	r.GrantCount = int64(len(grants))
	for _, g := range grants {
		switch {
		case isGlobalPattern(g.Operation):
			r.GlobalGrant = true
		case isWildcardPattern(g.Operation):
			r.WildcardCount++
		}
		if isNarrowingScope(g.ScopeKind) {
			r.NarrowedCount++
		}
		if len(g.PolicyIDs) > 0 {
			r.PolicyCount++
		}
	}
}
