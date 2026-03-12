package schema

import (
	"github.com/DrReMain/cyber-ecosystem/go-shared/orm/ent/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("username").NotEmpty(),
		field.String("email").Optional().Nillable(),
		field.Int32("age").Optional().Nillable(),
		field.String("password").NotEmpty(),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{}
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("username").Unique(),
		index.Fields("email").Unique(),
	}
}

func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "user"},
	}
}
