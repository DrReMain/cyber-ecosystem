package authz

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"sync/atomic"
	"time"

	"cyber-ecosystem/shared-go/capability/cache"
	"cyber-ecosystem/shared-go/kratos/security"

	"cyber-ecosystem/app/services/system/internal/ent/authzpolicy"
	"cyber-ecosystem/app/services/system/internal/ent/user"
	"cyber-ecosystem/app/services/system/internal/platform"
	"cyber-ecosystem/app/services/system/internal/shared"
)

const reconcileEvery = 5 * time.Minute

// Repo ----------------------------------------------------------------------------------------------------------------

type authzRP struct {
	shared.RP
	snap atomic.Pointer[snapshot]
}

func NewAuthzRP(logger *slog.Logger, p *platform.Platform) (AuthzRP, func(), error) {
	r := &authzRP{RP: shared.NewRP(logger.With("module", "module/authz"), p)}
	ctx := context.Background()
	snap, err := r.compileFromDB(ctx, r.remoteVersionSafe(ctx))
	if err != nil {
		return nil, nil, fmt.Errorf("authz rp: initial compile: %w", err)
	}
	r.snap.Store(snap)
	return r, r.watch(), nil
}

// Method --------------------------------------------------------------------------------------------------------------

func (r *authzRP) ImplicitGrants(ctx context.Context, userID, tenant string) ([]GrantView, error) {
	snap := r.snap.Load()
	if snap == nil {
		return nil, fmt.Errorf("authz rp: snapshot not compiled")
	}
	// The Management API filters policy rows by domain on its own, so the
	// tenant closure of this path does not depend on the Enforce matcher.
	rows, err := snap.enf.GetImplicitPermissionsForUser(principalSubject(shared.PrincipalTypeUser, userID), tenant)
	if err != nil {
		return nil, fmt.Errorf("authz rp: resolve grants for %q: %w", userID, err)
	}
	views := make([]GrantView, 0, len(rows)+len(builtinOperations))
	for _, op := range builtinOperations {
		views = append(views, GrantView{Pattern: op})
	}
	for _, row := range rows {
		if len(row) < 3 {
			continue
		}
		meta := snap.grants[grantKey{tenant, row[0], row[2]}]
		views = append(views, GrantView{
			Pattern:     row[2],
			ScopeKind:   meta.scopeKind,
			ScopeParams: meta.scopeParams,
			Policies:    policyViews(meta.policies),
		})
	}
	return r.hydrateScopes(ctx, userID, views)
}

func (r *authzRP) EffectiveGrants(ctx context.Context, subject *security.Subject) ([]GrantView, error) {
	views, err := r.ImplicitGrants(ctx, subject.UserID, subject.TenantID)
	if err != nil {
		return nil, err
	}
	return survivingGrants(ctx, r.Log, subject, views), nil
}

func (r *authzRP) ExplainGrants(ctx context.Context, principalType, principalID, tenant, operation string) (*ExplainResult, error) {
	snap := r.snap.Load()
	if snap == nil {
		return nil, fmt.Errorf("authz rp: snapshot not compiled")
	}
	out := &ExplainResult{Builtin: slices.Contains(builtinOperations, operation)}
	subj := principalSubject(principalType, principalID)
	// Grouping rows survive a role's disable, so the held-bindings list stays
	// honest about what the principal holds even when grants are gone.
	roles, err := snap.enf.GetRolesForUser(subj, tenant)
	if err != nil {
		return nil, fmt.Errorf("authz rp: resolve roles for %q: %w", principalID, err)
	}
	for _, rs := range roles {
		out.RoleCodes = append(out.RoleCodes, strings.TrimPrefix(rs, roleSubjectPrefix))
	}
	rows, err := snap.enf.GetImplicitPermissionsForUser(subj, tenant)
	if err != nil {
		return nil, fmt.Errorf("authz rp: resolve grants for %q: %w", principalID, err)
	}
	var views []GrantView
	var codes []string
	for _, row := range rows {
		if len(row) < 3 || !matchOperation(operation, row[2]) {
			continue
		}
		meta := snap.grants[grantKey{tenant, row[0], row[2]}]
		views = append(views, GrantView{
			Pattern:   row[2],
			ScopeKind: meta.scopeKind,
			Policies:  policyViews(meta.policies),
		})
		codes = append(codes, strings.TrimPrefix(row[0], roleSubjectPrefix))
	}
	// Explain evaluates at explain time with the same kernel as Decide, so
	// the verdict and the per-policy states agree with what enforcement
	// would do this instant.
	subject := &security.Subject{UserID: principalID, TenantID: tenant}
	var attrs map[string]any
	if anyLinkedPolicy(views) {
		if attrs, err = resolveAttributes(ctx, subject, neededAttrKeys(views)); err != nil {
			// Decide's posture: unresolvable attributes fail every constraint
			// closed; nil attrs make each plugin fail below, states show it.
			r.Log.Error("authz: attribute resolution failed; constraints fail closed", "error", err)
			attrs = nil
		}
	}
	survivedAny := false
	for i, v := range views {
		passed := true
		var states []PolicyState
		if len(v.Policies) > 0 {
			passed, states = evalGrantPolicies(ctx, r.Log, attrs, v.Policies)
		}
		survivedAny = survivedAny || passed
		out.Hits = append(out.Hits, GrantHit{
			Pattern:   v.Pattern,
			RoleCode:  codes[i],
			ScopeKind: v.ScopeKind,
			Policies:  states,
		})
	}
	out.Allowed = out.Builtin || survivedAny
	return out, nil
}

func (r *authzRP) PreviewGrants(ctx context.Context, tenant string, roleCodes []string) ([]string, error) {
	snap := r.snap.Load()
	if snap == nil {
		return nil, fmt.Errorf("authz rp: snapshot not compiled")
	}
	seen := make(map[string]struct{}, 2*len(roleCodes))
	ops := make([]string, 0, len(roleCodes))
	for _, code := range roleCodes {
		// The role's own policy rows: a disabled (or grantless) role simply
		// contributes nothing — the same outcome enforcement would produce.
		rows, err := snap.enf.GetPermissionsForUser(roleSubject(code), tenant)
		if err != nil {
			return nil, fmt.Errorf("authz rp: preview role %q: %w", code, err)
		}
		for _, row := range rows {
			if len(row) < 3 {
				continue
			}
			if _, ok := seen[row[2]]; ok {
				continue
			}
			seen[row[2]] = struct{}{}
			ops = append(ops, row[2])
		}
	}
	return ops, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func (r *authzRP) reconcile() {
	ctx := context.Background()
	ver, err := r.remoteVersion(ctx)
	if err != nil {
		r.Log.Warn("authz: version read failed; keeping snapshot", "version", r.version(), "error", err)
		return
	}
	// >= so a stamp that ever moves backwards never pins an old snapshot.
	if r.version() >= ver {
		return
	}
	if err := r.rebuild(ctx, ver); err != nil {
		r.Log.Warn("authz: rebuild failed; keeping snapshot", "version", r.version(), "error", err)
	}
}

func (r *authzRP) version() int64 {
	if s := r.snap.Load(); s != nil {
		return s.version
	}
	return -1
}

func (r *authzRP) remoteVersion(ctx context.Context) (int64, error) {
	ver, err := r.Platform.GetCache().Counter.Get(ctx, shared.PolicyVersionKey)
	if errors.Is(err, cache.ErrKeyNotFound) {
		return 0, nil // never bumped yet: version zero
	}
	return ver, err
}

func (r *authzRP) remoteVersionSafe(ctx context.Context) int64 {
	ver, err := r.remoteVersion(ctx)
	if err != nil {
		// A Redis outage must not block startup; the first reconcile after
		// recovery realigns the version.
		r.Log.Warn("authz: initial version read failed; starting unversioned", "error", err)
		return -1
	}
	return ver
}

func (r *authzRP) rebuild(ctx context.Context, version int64) error {
	snap, err := r.compileFromDB(ctx, version)
	if err != nil {
		return err
	}
	r.snap.Store(snap)
	r.Log.Info("authz: snapshot rebuilt", "version", version)
	return nil
}

func (r *authzRP) watch() func() {
	var sub cache.Subscription
	if s, err := r.Platform.GetCache().PubSub.Subscribe(context.Background(), shared.PolicyChangedChannel); err != nil {
		r.Log.Warn("authz: pubsub subscribe failed; periodic reconcile only", "error", err)
	} else {
		sub = s
	}
	done := make(chan struct{})
	go func() {
		if sub == nil {
			return
		}
		for range sub.Channel() {
			r.reconcile()
		}
	}()
	go func() {
		ticker := time.NewTicker(reconcileEvery)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				r.reconcile()
			}
		}
	}()
	return func() {
		close(done)
		if sub != nil {
			_ = sub.Close()
		}
	}
}

func (r *authzRP) compileFromDB(ctx context.Context, version int64) (*snapshot, error) {
	client := r.Platform.GetClient(ctx)
	roles, err := client.Role.Query().All(ctx)
	if err != nil {
		return nil, r.Platform.HandleEntError(fmt.Errorf("authz compile: load roles: %w", err))
	}
	perms, err := client.Permission.Query().All(ctx)
	if err != nil {
		return nil, r.Platform.HandleEntError(fmt.Errorf("authz compile: load permissions: %w", err))
	}
	polRows, err := client.AuthzPolicy.Query().Where(authzpolicy.EnabledEQ(true)).All(ctx)
	if err != nil {
		return nil, r.Platform.HandleEntError(fmt.Errorf("authz compile: load policies: %w", err))
	}
	linkRows, err := client.PermissionPolicy.Query().All(ctx)
	if err != nil {
		return nil, r.Platform.HandleEntError(fmt.Errorf("authz compile: load policy links: %w", err))
	}
	bindings, err := client.PrincipalRole.Query().All(ctx)
	if err != nil {
		return nil, r.Platform.HandleEntError(fmt.Errorf("authz compile: load bindings: %w", err))
	}
	disabledUsers, err := client.User.Query().Where(user.EnabledEQ(false)).All(ctx)
	if err != nil {
		return nil, r.Platform.HandleEntError(fmt.Errorf("authz compile: load disabled users: %w", err))
	}
	disabled := make(map[string]struct{}, len(disabledUsers))
	for _, du := range disabledUsers {
		disabled[du.ID] = struct{}{}
	}

	type roleInfo struct {
		code    string
		enabled bool
	}
	byID := make(map[string]roleInfo, len(roles))
	for _, ro := range roles {
		byID[ro.ID] = roleInfo{code: ro.Code, enabled: ro.Enabled}
	}
	type policyInfo struct {
		meta   policyMeta
		tenant string
	}
	policies := make(map[string]policyInfo, len(polRows))
	for _, po := range polRows {
		policies[po.ID] = policyInfo{
			meta:   policyMeta{kind: po.Kind, name: po.Name, params: po.Params},
			tenant: po.TenantID,
		}
	}
	linksByPerm := make(map[string][]string, len(linkRows))
	for _, l := range linkRows {
		linksByPerm[l.PermissionID] = append(linksByPerm[l.PermissionID], l.PolicyID)
	}
	grants := make([]grantRow, 0, len(perms))
	for _, p := range perms {
		ri, ok := byID[p.RoleID]
		if !ok {
			continue // orphaned grant row; hard delete removes it next rebuild
		}
		var polys []policyMeta
		for _, pid := range linksByPerm[p.ID] {
			pi, ok := policies[pid]
			if !ok || pi.tenant != p.TenantID {
				continue
			}
			polys = append(polys, pi.meta)
		}
		grants = append(grants, grantRow{
			RoleCode:    ri.code,
			TenantID:    p.TenantID,
			Pattern:     p.Operation,
			Enabled:     ri.enabled,
			ScopeKind:   p.ScopeKind,
			ScopeParams: p.ScopeParams,
			Policies:    polys,
		})
	}
	rows := make([]bindingRow, 0, len(bindings))
	for _, b := range bindings {
		ri, ok := byID[b.RoleID]
		if !ok {
			continue
		}
		enabled := true
		if b.PrincipalType == shared.PrincipalTypeUser {
			_, enabled = disabled[b.PrincipalID]
			enabled = !enabled
		}
		rows = append(rows, bindingRow{
			PrincipalType: b.PrincipalType,
			PrincipalID:   b.PrincipalID,
			RoleCode:      ri.code,
			TenantID:      b.TenantID,
			Enabled:       enabled,
		})
	}
	return compile(version, grants, rows)
}

func policyViews(polys []policyMeta) []PolicyView {
	if len(polys) == 0 {
		return nil
	}
	views := make([]PolicyView, 0, len(polys))
	for _, p := range polys {
		views = append(views, PolicyView{Kind: p.kind, Name: p.name, Params: p.params})
	}
	return views
}
