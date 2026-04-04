package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"
)

// Reading holds the schema definition for the Reading entity
type Reading struct {
	ent.Schema
}

// Fields of the Reading
func (Reading) Fields() []ent.Field {
	return []ent.Field{
		field.String("blog_id").NotEmpty(),
		field.Int64("reading_count"),
	}
}

// Edges of the Reading
func (Reading) Edges() []ent.Edge {
	return []ent.Edge{}
}

func (Reading) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
	}
}

func (Reading) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("blog_id").Unique(),
	}
}

func (Reading) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "reading"},
	}
}
