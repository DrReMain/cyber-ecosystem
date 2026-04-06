package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"
)

type UserAttribute struct {
	ent.Schema
}

func (UserAttribute) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id").NotEmpty().MaxLen(20),
		field.String("key").NotEmpty().MaxLen(50),
		field.String("value").NotEmpty().MaxLen(200),
	}
}

func (UserAttribute) Edges() []ent.Edge {
	return []ent.Edge{}
}

func (UserAttribute) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
	}
}

func (UserAttribute) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "key").Unique(),
	}
}

func (UserAttribute) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "user_attribute"},
	}
}
