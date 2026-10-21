package local_mixins

import (
	"context"
	"fmt"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/mixin"

	"cyber-ecosystem/shared-go/kratos/security"
	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"
	"cyber-ecosystem/shared-go/orm/ent/scopes"

	gen "cyber-ecosystem/app/services/system/internal/ent"
	"cyber-ecosystem/app/services/system/internal/ent/hook"
	"cyber-ecosystem/app/services/system/internal/ent/intercept"
)

// DatascopeMixin narrows rows to the ctx Decision's scopes; kinds OR-union
// into one predicate, and nothing expressible fails closed — never unfiltered.
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

// scopePredicate compiles the decision's scopes via the shared compiler, or
// returns nil when this read is unscoped (lifted, or no decision in ctx).
func (d DatascopeMixin) scopePredicate(ctx context.Context) func(*sql.Selector) {
	if _, lifted := ctx.Value(unscopedKey{}).(bool); lifted {
		return nil
	}
	if security.IsRegistryRead(ctx) {
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
	res := scopes.Compile(dn.Scopes, scopes.Fields{User: d.UserField, Dept: d.DeptField}, subject.UserID)
	if res.Unscoped {
		return nil
	}
	return res.Predicate
}

func applyScope(w interface{ WhereP(...func(*sql.Selector)) }, p func(*sql.Selector)) {
	w.WhereP(p)
}
