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

type MatchConfig struct {
	ent.Schema
}

func (MatchConfig) Fields() []ent.Field {
	return []ent.Field{
		field.String("genre").NotEmpty().MaxRuneLen(16).Immutable().
			Comment("type-axis vocabulary fps|moba|... — the scenario axis (channel batches) joins the same column later; open word set, admin-managed, same law as test.category"),
		field.String("ability").NotEmpty().MaxRuneLen(32).Immutable().
			Comment("six-word vocabulary reactionSpeed|aimingAccuracy|... — the ability this slot draws a test for"),
		field.Int("render_order").Default(1).
			Comment("orchestration position inside the genre; steps emit in this order, so retuning reorders lineups without a contract change"),
	}
}

func (MatchConfig) Edges() []ent.Edge {
	return nil
}

func (MatchConfig) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
		local_mixins.TenantMixin{},
		local_mixins.SoftDeleteMixin{},
	}
}

func (MatchConfig) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "created_at"),
		index.Fields("genre", "render_order"),
		index.Fields("genre", "ability").Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),
	}
}

func (MatchConfig) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "match_config"},
	}
}
