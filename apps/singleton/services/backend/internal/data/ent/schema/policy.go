package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"
)

type PolicyRule struct {
	ent.Schema
}

func (PolicyRule) Fields() []ent.Field {
	return []ent.Field{
		field.String("ptype").NotEmpty().MaxLen(10).Comment("policy type: p, g"),
		field.String("v0").Optional().MaxLen(200),
		field.String("v1").Optional().MaxLen(200),
		field.String("v2").Optional().MaxLen(200),
		field.String("v3").Optional().MaxLen(200),
		field.String("v4").Optional().MaxLen(200),
		field.String("v5").Optional().MaxLen(200),
	}
}

func (PolicyRule) Edges() []ent.Edge {
	return []ent.Edge{}
}

func (PolicyRule) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
	}
}

func (PolicyRule) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("ptype", "v0", "v1", "v2", "v3").Unique(),
	}
}

func (PolicyRule) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "policy"},
	}
}
