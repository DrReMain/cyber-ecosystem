package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"
)

type DataScope struct {
	ent.Schema
}

func (DataScope) Fields() []ent.Field {
	return []ent.Field{
		field.String("role_code").NotEmpty().MaxLen(50),
		field.String("scope_type").NotEmpty().MaxLen(20).Comment("all, self, dept, custom"),
		field.Text("scope_config").Optional().Nillable().Comment("JSON config"),
		field.String("target_resource").Optional().MaxLen(200).Comment("target operation pattern, empty = all"),
	}
}

func (DataScope) Edges() []ent.Edge {
	return []ent.Edge{}
}

func (DataScope) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
	}
}

func (DataScope) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("role_code"),
	}
}

func (DataScope) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "data_scope"},
	}
}
