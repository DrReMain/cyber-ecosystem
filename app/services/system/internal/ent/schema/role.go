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

type Role struct {
	ent.Schema
}

func (Role) Fields() []ent.Field {
	return []ent.Field{
		field.String("code").NotEmpty().MaxRuneLen(64).Immutable().
			Comment("stable identifier; user role bindings reference codes, not ids"),
		field.String("name").NotEmpty().MaxRuneLen(128),
		field.Bool("enabled").Default(true).
			Comment("false drops the role's grants at engine compile (tighten direction)"),
		field.String("remark").Default("").MaxRuneLen(255),
	}
}

func (Role) Edges() []ent.Edge {
	return nil
}

func (Role) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
		local_mixins.TenantMixin{},
	}
}

func (Role) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "code").Unique(),
	}
}

func (Role) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "role"},
	}
}
