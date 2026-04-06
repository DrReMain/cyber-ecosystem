package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"

	"cyber-ecosystem/apps/singleton/services/backend/internal/data/ent/schema/local_mixins"
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("password").NotEmpty().MaxLen(256),
		field.String("email").NotEmpty().MaxLen(256),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("departments", Department.Type).
			Ref("users"),
	}
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
		local_mixins.SoftDeleteMixin{},
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email").Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),
	}
}

func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "user"},
	}
}
