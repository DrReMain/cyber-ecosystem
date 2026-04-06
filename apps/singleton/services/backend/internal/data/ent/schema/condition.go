package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"
)

type Condition struct {
	ent.Schema
}

func (Condition) Fields() []ent.Field {
	return []ent.Field{
		field.String("role_code").NotEmpty().MaxLen(50),
		field.String("operation").NotEmpty().MaxLen(200),
		field.String("condition_type").NotEmpty().MaxLen(50),
		field.Text("config").Optional().Nillable(),
		field.String("group_id").Optional().Nillable().MaxLen(50).Comment("optional group for OR logic between groups"),
	}
}

func (Condition) Edges() []ent.Edge {
	return []ent.Edge{}
}

func (Condition) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
	}
}

func (Condition) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("role_code"),
		index.Fields("role_code", "operation"),
		index.Fields("role_code", "group_id"),
	}
}

func (Condition) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "condition"},
	}
}
