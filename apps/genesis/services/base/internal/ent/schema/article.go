package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"

	"cyber-ecosystem/apps/genesis/services/base/internal/ent/schema/local_mixins"
)

type Article struct {
	ent.Schema
}

func (Article) Fields() []ent.Field {
	return []ent.Field{
		field.String("title").NotEmpty().MaxLen(64),
		field.Text("content").Default("").MaxLen(1024),
		field.String("status").Default("draft").MaxLen(10).Comment("draft/published/archived"),
	}
}

func (Article) Edges() []ent.Edge {
	return []ent.Edge{}
}

func (Article) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
		local_mixins.SortMixin{SoftDelete: true},
		local_mixins.SoftDeleteMixin{},
	}
}

func (Article) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("title").Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),
	}
}

func (Article) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "article"},
	}
}
