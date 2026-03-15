package schema

import (
	"github.com/DrReMain/cyber-ecosystem/examples/template1/internal/data/ent/schema/local_mixins"

	"github.com/DrReMain/cyber-ecosystem/shared-go/orm/ent/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/matoous/go-nanoid/v2"
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
		field.String("title").NotEmpty().DefaultFunc(BlogDefaultTitle).SchemaType(map[string]string{dialect.Postgres: "varchar(64)"}).Comment("标题"),
		field.Text("content").DefaultFunc(BlogDefaultContent).Comment("内容"),
		field.Time("published_at").Optional().Nillable().Comment("发布时间"),
	}
}

// Edges of the Blog.
func (Blog) Edges() []ent.Edge {
	return []ent.Edge{}
}

func (Blog) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		local_mixins.SoftDeleteMixin{},
	}
}

func (Blog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("title").Unique(),
	}
}

func (Blog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "blog"},
	}
}
