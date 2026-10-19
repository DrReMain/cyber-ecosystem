package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"
)

type Credential struct {
	ent.Schema
}

func (Credential) Fields() []ent.Field {
	return []ent.Field{
		field.String("kind").NotEmpty().MaxRuneLen(16).Immutable().
			Comment("code-registered exchange kind (provider_<kind>.go)"),
		field.String("identifier").NotEmpty().MaxRuneLen(191).Immutable().
			Comment("the external identifier the kind exchanged (e.g. wechat openid)"),
		field.String("user_id").NotEmpty().MaxRuneLen(20).
			Comment("logical reference to user.id"),
	}
}

func (Credential) Edges() []ent.Edge {
	return nil
}

// Identity-plane registry data: one external identifier resolves to one
// global identity — no tenant, no soft delete (a deleted row must not hold
// the unique hostage).
func (Credential) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
	}
}

func (Credential) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("kind", "identifier").Unique(),
		index.Fields("user_id"),
	}
}

func (Credential) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "credential"},
	}
}
