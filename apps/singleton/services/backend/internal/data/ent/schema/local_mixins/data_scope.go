package local_mixins

import (
	"context"
	"errors"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/mixin"

	gen "cyber-ecosystem/apps/singleton/services/backend/internal/data/ent"
	"cyber-ecosystem/apps/singleton/services/backend/internal/data/ent/hook"
	"cyber-ecosystem/apps/singleton/services/backend/internal/data/ent/intercept"
	"cyber-ecosystem/apps/singleton/services/backend/internal/pkg/datascope"
	"cyber-ecosystem/apps/singleton/services/backend/internal/pkg/security"
)

// DataScopeMixin provides automatic data permission filtering for Ent schemas.
// Configure with DeptField and OwnerField to declare which column names the mixin creates.
// The mixin auto-injects the fields, indexes, interceptors (read filter), and hooks (write filter).
type DataScopeMixin struct {
	mixin.Schema
	DeptField  string
	OwnerField string
}

// Fields delegates to the shared datascope mixin.
func (d DataScopeMixin) Fields() []ent.Field {
	return datascope.DataScopeMixin{DeptField: d.DeptField, OwnerField: d.OwnerField}.Fields()
}

// Indexes delegates to the shared datascope mixin.
func (d DataScopeMixin) Indexes() []ent.Index {
	return datascope.DataScopeMixin{DeptField: d.DeptField, OwnerField: d.OwnerField}.Indexes()
}

// Interceptors of the DataScopeMixin.
func (d DataScopeMixin) Interceptors() []gen.Interceptor {
	return []gen.Interceptor{
		intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
			if datascope.SkipDataScopeFromContext(ctx) {
				return nil
			}
			resolver, ok := datascope.ScopeResolverFromContext(ctx)
			if !ok || resolver == nil {
				return nil
			}
			userID, ok := datascope.ScopeUserIDFromContext(ctx)
			if !ok || userID == "" {
				return nil
			}
			operation, ok := security.OperationFromContext(ctx)
			if !ok || operation == "" {
				return errors.New("missing security operation for data scope resolution")
			}

			scope, err := resolver(ctx, userID, operation)
			if err != nil {
				return err
			}

			d.applyScope(q, scope, userID)
			return nil
		}),
	}
}

// Hooks of the DataScopeMixin.
func (d DataScopeMixin) Hooks() []gen.Hook {
	return []gen.Hook{
		hook.On(
			func(next gen.Mutator) gen.Mutator {
				return gen.MutateFunc(func(ctx context.Context, m gen.Mutation) (gen.Value, error) {
					if datascope.SkipDataScopeFromContext(ctx) {
						return next.Mutate(ctx, m)
					}
					if d.OwnerField != "" && m.Op() == gen.OpCreate {
						userID, ok := datascope.ScopeUserIDFromContext(ctx)
						if ok && userID != "" {
							if _, exists := m.Field(d.OwnerField); !exists {
								m.SetField(d.OwnerField, userID)
							}
						}
					}
					return next.Mutate(ctx, m)
				})
			},
			gen.OpCreate,
		),
	}
}

func (d DataScopeMixin) applyScope(w interface{ WhereP(...func(*sql.Selector)) }, scope *datascope.EffectiveScope, userID string) {
	if scope.IsAll {
		return
	}

	var orPreds []func(*sql.Selector)

	if scope.SelfFilter && d.OwnerField != "" {
		orPreds = append(orPreds, sql.FieldEQ(d.OwnerField, userID))
	}

	if scope.DeptFilter && d.DeptField != "" && len(scope.DeptIDs) > 0 {
		orPreds = append(orPreds, sql.FieldIn(d.DeptField, toAnySlice(scope.DeptIDs)...))
	}

	if scope.AttributeFilter && len(scope.Rules) > 0 {
		attrPreds := datascope.BuildRulePredicates(scope.Rules)
		if len(attrPreds) > 0 {
			if scope.Logic == "or" {
				orPreds = append(orPreds, sql.OrPredicates(attrPreds...))
			} else {
				andGroup := func(s *sql.Selector) {
					for _, p := range attrPreds {
						p(s)
					}
				}
				orPreds = append(orPreds, andGroup)
			}
		}
	}

	if len(scope.ExtraPredicates) > 0 {
		orPreds = append(orPreds, scope.ExtraPredicates...)
	}

	if len(orPreds) == 1 {
		w.WhereP(orPreds[0])
	} else if len(orPreds) > 1 {
		w.WhereP(sql.OrPredicates(orPreds...))
	}
}

func toAnySlice(s []string) []any {
	result := make([]any, len(s))
	for i, v := range s {
		result[i] = v
	}
	return result
}
