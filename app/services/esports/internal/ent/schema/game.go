package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"

	"cyber-ecosystem/app/services/esports/internal/ent/schema/local_mixins"
)

type Game struct {
	ent.Schema
}

func (Game) Fields() []ent.Field {
	return []ent.Field{
		field.String("code").NotEmpty().MaxRuneLen(50).Immutable().
			Comment("immutable catalog code join key (e.g. lol); born once, referenced forever"),
		field.String("name").NotEmpty().MaxRuneLen(100).
			Comment("display title, legacy experience data"),
		field.String("category").Optional().Nillable().MaxRuneLen(16).
			Comment("vocabulary fps|moba|... — open word set, admin-managed; new categories classify in without a contract change"),
		field.String("icon_url").Optional().Nillable().MaxRuneLen(500).
			Comment("icon URL; hydrated from file service at render"),
		field.String("description").Optional().Nillable().MaxRuneLen(500).
			Comment("short catalog blurb"),
	}
}

func (Game) Edges() []ent.Edge {
	return nil
}

func (Game) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
		local_mixins.TenantMixin{},
		local_mixins.SoftDeleteMixin{},
	}
}

func (Game) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "created_at"),
		index.Fields("code").Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),
	}
}

func (Game) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "game"},
	}
}
