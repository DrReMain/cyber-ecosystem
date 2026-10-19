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

// Typed ABAC constraint; kind must match a registered policy plugin. Go
// name avoids ent's predeclared "Policy"; the table stays "policy".
type AuthzPolicy struct {
	ent.Schema
}

func (AuthzPolicy) Fields() []ent.Field {
	return []ent.Field{
		field.String("kind").NotEmpty().MaxRuneLen(32).Immutable().
			Comment("policy type; must match a registered engine plugin"),
		field.String("name").NotEmpty().MaxRuneLen(128),
		field.JSON("params", map[string]any{}).
			Comment("typed plugin payload as jsonb; read whole, never queried internally"),
		field.Bool("enabled").Default(true).
			Comment("false unlinks the constraint at compile (relax direction)"),
	}
}

func (AuthzPolicy) Edges() []ent.Edge {
	return nil
}

func (AuthzPolicy) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
		local_mixins.TenantMixin{},
	}
}

func (AuthzPolicy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "kind", "name").Unique(),
	}
}

func (AuthzPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "policy"},
	}
}
