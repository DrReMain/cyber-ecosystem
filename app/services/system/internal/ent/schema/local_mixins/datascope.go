package local_mixins

import (
	"context"
	"fmt"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/mixin"

	"cyber-ecosystem/shared-go/kratos/security"
	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"

	gen "cyber-ecosystem/app/services/system/internal/ent"
	"cyber-ecosystem/app/services/system/internal/ent/hook"
	"cyber-ecosystem/app/services/system/internal/ent/intercept"
)

// DatascopeMixin narrows queries and mutations to the rows permitted by the
// ctx Decision's scopes. Descriptors match by kind expressibility:
// self needs UserField, dept_tree needs DeptField; an empty field means the
// kind is not expressible on this entity and its descriptors are skipped
// ("no matching scope → no filter"). Narrowing is the OR-union of expressible
// scopes; an all/"" descriptor dominates and lifts filtering entirely.
// Paths without a Decision (seed, login, engine compilation, builtin
// baseline) skip filtering, mirroring TenantMixin's no-subject skip.
type DatascopeMixin struct {
	mixin.Schema
	UserField string // owner column for self scopes ("id" on user); empty = self not expressible
	DeptField string // org column for dept_tree scopes ("dept_id" on user, "id" on dept); empty = dept_tree not expressible
}

type unscopedKey struct{}

// Unscoped lifts datascope narrowing for the reads it wraps. Reserved for the
// reference plane — resolving a bound reference (avatar → file id → key)
// whose authorization rides the referring row, not the target's owner scope.
func Unscoped(ctx context.Context) context.Context {
	return context.WithValue(ctx, unscopedKey{}, true)
}

func (DatascopeMixin) Fields() []ent.Field { return nil }

func (DatascopeMixin) Indexes() []ent.Index { return nil }

func (d DatascopeMixin) Interceptors() []gen.Interceptor {
	return []gen.Interceptor{
		intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
			if p := d.scopePredicate(ctx); p != nil {
				applyScope(q, p)
			}
			return nil
		}),
	}
}

// Hooks guard update mutations with the same predicate — user/dept deletes are
// rewritten to updates by SoftDeleteMixin and re-enter the hook chain, so
// OpUpdate|OpUpdateOne covers writes without double application. An
// out-of-scope row is silently untouched (0 rows); the UC-level FindByID
// ensure surfaces that as NotFound, leaking neither existence nor scope.
func (d DatascopeMixin) Hooks() []gen.Hook {
	return []gen.Hook{
		hook.On(
			func(next gen.Mutator) gen.Mutator {
				return gen.MutateFunc(func(ctx context.Context, m gen.Mutation) (ent.Value, error) {
					p := d.scopePredicate(ctx)
					if p == nil {
						return next.Mutate(ctx, m)
					}
					mx, ok := m.(interface{ WhereP(...func(*sql.Selector)) })
					if !ok {
						return nil, fmt.Errorf("datascope mixin: unexpected mutation type %T", m)
					}
					applyScope(mx, p)
					return next.Mutate(ctx, m)
				})
			},
			gen.OpUpdate|gen.OpUpdateOne,
		),
	}
}

// scopePredicate compiles the decision's scopes into one OR predicate, or nil
// when this entity is unscoped by them.
func (d DatascopeMixin) scopePredicate(ctx context.Context) func(*sql.Selector) {
	if _, lifted := ctx.Value(unscopedKey{}).(bool); lifted {
		return nil
	}
	dn, ok := kauthz.DecisionFromCtx(ctx)
	if !ok {
		return nil
	}
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok {
		return nil
	}
	var preds []*sql.Predicate
	for _, sc := range dn.Scopes {
		switch sc.Kind {
		case "", kauthz.ScopeKindAll:
			return nil // a tenant-wide grant dominates the union
		case kauthz.ScopeKindSelf:
			if d.UserField != "" {
				preds = append(preds, sql.EQ(d.UserField, subject.UserID))
			}
		case kauthz.ScopeKindDeptTree:
			if d.DeptField != "" {
				// Empty subtree (deptless subject) or malformed Params both
				// compile to FALSE via In's zero-arg form — the grant yields
				// no rows rather than no filtering.
				preds = append(preds, sql.In(d.DeptField, deptIDArgs(sc)...))
			}
		}
	}
	if len(preds) == 0 {
		return nil
	}
	or := sql.Or(preds...)
	return func(s *sql.Selector) { s.Where(or) }
}

func applyScope(w interface{ WhereP(...func(*sql.Selector)) }, p func(*sql.Selector)) {
	w.WhereP(p)
}

func deptIDArgs(sc kauthz.ScopeDescriptor) []any {
	raw, _ := sc.Params[kauthz.ScopeParamDeptIDs].([]string)
	args := make([]any, len(raw))
	for i, id := range raw {
		args[i] = id
	}
	return args
}
