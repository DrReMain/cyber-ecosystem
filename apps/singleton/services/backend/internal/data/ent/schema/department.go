package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"
)

type Department struct {
	ent.Schema
}

func (Department) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().MaxLen(100),
		field.String("code").NotEmpty().MaxLen(50),
		field.String("parent_id").Optional().Nillable().MaxLen(20).Comment("parent department id, null = root"),
		field.String("path").NotEmpty().MaxLen(500).Comment("materialized path, e.g. /d1/d3/d7/"),
		field.Uint8("status").Default(1).Comment("1=enabled, 0=disabled"),
	}
}

func (Department) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("users", User.Type),
	}
}

func (Department) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
		mixins.SortMixin{},
	}
}

func (Department) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code").Unique(),
		index.Fields("parent_id"),
		index.Fields("path"),
	}
}

func (Department) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "department"},
	}
}
