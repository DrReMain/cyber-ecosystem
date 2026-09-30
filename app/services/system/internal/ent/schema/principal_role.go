package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"

	"cyber-ecosystem/app/services/system/internal/ent/schema/local_mixins"
)

// Binds any principal kind to a role; principal_type is a
// plain string so new kinds need no migration.
type PrincipalRole struct {
	ent.Schema
}

func (PrincipalRole) Fields() []ent.Field {
	return []ent.Field{
		field.String("principal_type").NotEmpty().MaxRuneLen(16).Immutable().
			Comment("user; plain string so new kinds need no migration"),
		field.String("principal_id").NotEmpty().MaxRuneLen(20).Immutable(),
		field.String("role_id").NotEmpty().MaxRuneLen(20).Immutable(),
	}
}

func (PrincipalRole) Edges() []ent.Edge {
	return nil
}

func (PrincipalRole) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
		local_mixins.TenantMixin{},
	}
}

func (PrincipalRole) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("principal_type", "principal_id", "role_id").Unique(),
	}
}

func (PrincipalRole) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "principal_role"},
	}
}
