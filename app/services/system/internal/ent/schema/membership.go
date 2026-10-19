package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"
)

type Membership struct {
	ent.Schema
}

func (Membership) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id").NotEmpty().MaxRuneLen(20).
			Comment("logical reference to user.id"),
		field.String("tenant_id").NotEmpty().MaxRuneLen(20).
			Comment("tenant code — the join key, mirroring tenant-scoped rows everywhere"),
		field.String("population").NotEmpty().MaxRuneLen(20).
			Comment("population code — the archetype instance this seat carries"),
	}
}

func (Membership) Edges() []ent.Edge {
	return nil
}

func (Membership) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
	}
}

func (Membership) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "tenant_id", "population").Unique(),
		index.Fields("user_id"),
	}
}

func (Membership) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "membership"},
	}
}
