package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"

	"cyber-ecosystem/app/services/core/internal/ent/schema/local_mixins"
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("nickname").Optional().MaxLen(64),
		field.String("avatar").Optional().MaxLen(512),
		field.String("phone").NotEmpty().MaxLen(20).Comment("login account"),
		field.String("status").Default("enabled").MaxLen(16).Comment("enabled|disabled|restricted"),
		field.String("password_hash").Sensitive().NotEmpty().MaxLen(128),
	}
}

func (User) Edges() []ent.Edge {
	return nil
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
		local_mixins.SoftDeleteMixin{},
		local_mixins.SortMixin{SoftDelete: true, Table: "users"},
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("phone").Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),
	}
}

func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "users"},
	}
}
