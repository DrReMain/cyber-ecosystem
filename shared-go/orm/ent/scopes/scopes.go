// Package scopes compiles authz scope descriptors into ent SQL predicates —
// the execution half of datascope narrowing, shared by every service whose
// entities declare scope dimensions. The judgment half (which scopes a
// subject holds) stays with the IAM engine, local or remote; this package
// turns the descriptors into row filters where the rows live.
//
// The compilation law is fail-closed: a descriptor whose kind is unknown, or
// whose dimension the entity does not carry, contributes nothing to the
// union, and a descriptor set with nothing honorable yields Closed — the
// read returns the empty set, never falls back to unfiltered. The two
// sanctioned lifts are structural: a tenant-wide descriptor (kind ""/"all")
// dominates the union by design, and the empty set (the builtin baseline)
// narrows nothing.
package scopes

import (
	"entgo.io/ent/dialect/sql"

	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"
)

// Fields names the column each scope kind narrows on; an empty field means
// the kind is not expressible on the entity and its descriptors contribute
// nothing to the union.
type Fields struct {
	User string // owner column for "self" scopes
	Dept string // org column for "dept_tree" scopes
}

// Result is the outcome of compiling one descriptor set — exactly one of
// three states: Filtered (Predicate is the OR-union narrowing), Unscoped
// (narrowing lifted, apply no filter), Closed (Predicate is always false —
// narrowing was promised but nothing honorable remained).
type Result struct {
	Predicate func(*sql.Selector)
	Unscoped  bool
	Closed    bool
}

// Compile turns a decision's scopes into the narrowing predicate for one
// entity shape. userID feeds "self" descriptors; the dept id list rides the
// descriptor params (hydrated at decide time by the engine).
func Compile(descriptors []kauthz.ScopeDescriptor, f Fields, userID string) Result {
	if len(descriptors) == 0 {
		return Result{Unscoped: true}
	}
	preds := make([]*sql.Predicate, 0, len(descriptors))
	for _, d := range descriptors {
		switch d.Kind {
		case "", kauthz.ScopeKindAll:
			return Result{Unscoped: true}
		case kauthz.ScopeKindSelf:
			if f.User != "" {
				preds = append(preds, sql.EQ(f.User, userID))
			}
		case kauthz.ScopeKindDeptTree:
			if f.Dept != "" {
				// Zero ids (deptless subject) keep the zero-arg In form,
				// which entgo compiles to the FALSE constant: the grant
				// yields no rows rather than no filtering.
				preds = append(preds, sql.In(f.Dept, deptIDs(d)...))
			}
		}
	}
	if len(preds) == 0 {
		return Result{Closed: true, Predicate: func(s *sql.Selector) { s.Where(sql.ExprP("false")) }}
	}
	or := sql.Or(preds...)
	return Result{Predicate: func(s *sql.Selector) { s.Where(or) }}
}

// deptIDs extracts the hydrated dept id list; a malformed value compiles to
// the zero-arg (FALSE) form along with it.
func deptIDs(d kauthz.ScopeDescriptor) []any {
	raw, _ := d.Params[kauthz.ScopeParamDeptIDs].([]string)
	args := make([]any, len(raw))
	for i, id := range raw {
		args[i] = id
	}
	return args
}
