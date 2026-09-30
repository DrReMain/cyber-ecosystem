package authz

import (
	"fmt"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
)

const (
	// RBAC with domains, no act dimension: subjects are "user:<id>" /
	// "role:<code>", domain is the tenant.
	modelText = `
[request_definition]
r = sub, dom, obj

[policy_definition]
p = sub, dom, obj

[role_definition]
g = _, _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom && matchOp(r.obj, p.obj)
`
)

const roleSubjectPrefix = "role:"

type snapshot struct {
	version int64
	enf     *casbin.Enforcer
	grants  map[grantKey]grantMeta
}

type grantRow struct {
	RoleCode    string
	TenantID    string
	Pattern     string
	Enabled     bool
	ScopeKind   string
	ScopeParams map[string]any
	Policies    []policyMeta
}

type bindingRow struct {
	PrincipalType string
	PrincipalID   string
	RoleCode      string
	TenantID      string
	Enabled       bool
}

type grantKey struct {
	tenant  string
	roleSub string
	pattern string
}

type policyMeta struct {
	kind   string
	name   string
	params map[string]any
}

type grantMeta struct {
	scopeKind   string
	scopeParams map[string]any
	policies    []policyMeta
}

func roleSubject(code string) string { return roleSubjectPrefix + code }

func principalSubject(ptype, id string) string { return ptype + ":" + id }

func matchOpFunc(args ...any) (any, error) {
	op, ok1 := args[0].(string)
	pattern, ok2 := args[1].(string)
	if !ok1 || !ok2 {
		return false, fmt.Errorf("matchOp: want string arguments, got %T, %T", args[0], args[1])
	}
	return matchOperation(op, pattern), nil
}

func compile(version int64, grants []grantRow, bindings []bindingRow) (*snapshot, error) {
	// matchOp is registered as the matcher function so casbin's Enforce and
	// the biz-side matching share one implementation.
	m, err := model.NewModelFromString(modelText)
	if err != nil {
		return nil, fmt.Errorf("authz compile: load model: %w", err)
	}
	enf, err := casbin.NewEnforcer(m)
	if err != nil {
		return nil, fmt.Errorf("authz compile: new enforcer: %w", err)
	}
	enf.AddFunction("matchOp", matchOpFunc)

	var ps [][]string
	meta := make(map[grantKey]grantMeta, len(grants))
	for _, g := range grants {
		// A disabled role keeps its grouping rows but loses its policy rows
		// (members fall through to NO_GRANT); re-enabling restores on the
		// next rebuild.
		if !g.Enabled {
			continue
		}
		rs := roleSubject(g.RoleCode)
		ps = append(ps, []string{rs, g.TenantID, g.Pattern})
		meta[grantKey{g.TenantID, rs, g.Pattern}] = grantMeta{scopeKind: g.ScopeKind, scopeParams: g.ScopeParams, policies: g.Policies}
	}
	var gs [][]string
	for _, b := range bindings {
		if !b.Enabled {
			continue
		}
		gs = append(gs, []string{principalSubject(b.PrincipalType, b.PrincipalID), roleSubject(b.RoleCode), b.TenantID})
	}
	if len(ps) > 0 {
		if _, err := enf.AddPolicies(ps); err != nil {
			return nil, fmt.Errorf("authz compile: add policies: %w", err)
		}
	}
	if len(gs) > 0 {
		if _, err := enf.AddGroupingPolicies(gs); err != nil {
			return nil, fmt.Errorf("authz compile: add groupings: %w", err)
		}
	}
	return &snapshot{version: version, enf: enf, grants: meta}, nil
}
