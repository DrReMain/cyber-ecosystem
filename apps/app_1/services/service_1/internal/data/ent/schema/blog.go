package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/matoous/go-nanoid/v2"

	"cyber-ecosystem/shared-go/orm/ent/mixins"

	"cyber-ecosystem/apps/app_1/services/service_1/internal/data/ent/schema/local_mixins"
)

var (
	BlogDefaultTitle   = func() string { return gonanoid.MustGenerate("ABCDEFGHIJKLMNOPQRSTUVWXYZ", 32) }
	BlogDefaultContent = func() string { return "" }
)

// Blog holds the schema definition for the Blog entity.
type Blog struct {
	ent.Schema
}

// Fields of the Blog.
func (Blog) Fields() []ent.Field {
	return []ent.Field{
		field.String("title").NotEmpty().DefaultFunc(BlogDefaultTitle).SchemaType(map[string]string{dialect.Postgres: "varchar(64)"}).Comment("Title"),
		field.Text("content").DefaultFunc(BlogDefaultContent).Comment("Content"),
		field.Time("published_at").Optional().Nillable().Comment("Publish Time"),
	}
}

// Edges of the Blog.
func (Blog) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("authors", Author.Type).StorageKey(edge.Table("blog_author")),
	}
}

func (Blog) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
		local_mixins.SoftDeleteMixin{},
	}
}

func (Blog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("title").
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("id").
			StorageKey("blog_id_published_null").
			Annotations(entsql.IndexWhere("published_at IS NULL")),
	}
}

func (Blog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "blog"},
	}
}
