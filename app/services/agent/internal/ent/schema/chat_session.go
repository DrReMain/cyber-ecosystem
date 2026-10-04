package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"

	"cyber-ecosystem/app/services/agent/internal/ent/schema/local_mixins"
)

type ChatSession struct {
	ent.Schema
}

func (ChatSession) Fields() []ent.Field {
	return []ent.Field{
		field.String("owner_id").NotEmpty().MaxRuneLen(20).
			Comment("owning user; me-face — every query filters owner_id"),
		field.String("title").Default("").MaxRuneLen(64).
			Comment("first user message clipped to 20 runes + ellipsis; written once at creation"),
	}
}

func (ChatSession) Edges() []ent.Edge {
	return nil
}

func (ChatSession) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
		local_mixins.TenantMixin{},
		local_mixins.SoftDeleteMixin{},
	}
}

func (ChatSession) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("owner_id", "updated_at").
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),
	}
}

func (ChatSession) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "chat_session"},
	}
}
