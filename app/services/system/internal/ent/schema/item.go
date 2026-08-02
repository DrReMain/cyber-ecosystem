package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"

	"cyber-ecosystem/app/services/system/internal/ent/schema/local_mixins"
)

type Item struct {
	ent.Schema
}

func (Item) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().MaxLen(64),
		field.Text("description").Default("").MaxLen(1024),
		field.String("status").Default("active").MaxLen(10).Comment("active/inactive"),
	}
}

func (Item) Edges() []ent.Edge {
	return nil
}

func (Item) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
		local_mixins.SoftDeleteMixin{},
		// Table must match the entsql table annotation below; SortMixin queries
		// max(sort) by table name to auto-assign a key on create.
		local_mixins.SortMixin{SoftDelete: true, Table: "item"},
	}
}

func (Item) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name").Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),
	}
}

func (Item) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "item"},
	}
}
