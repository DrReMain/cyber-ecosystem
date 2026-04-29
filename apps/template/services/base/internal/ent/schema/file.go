package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"
)

type File struct {
	ent.Schema
}

func (File) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").MaxLen(255),
		field.String("content_type").MaxLen(128),
		field.Int64("size").NonNegative(),
		field.String("status").Default("pending").MaxLen(20).Comment("pending / attached"),
	}
}

func (File) Edges() []ent.Edge {
	return []ent.Edge{}
}

func (File) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
	}
}

func (File) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
	}
}

func (File) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "file"},
	}
}
