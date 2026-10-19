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

type AbilityScore struct {
	ent.Schema
}

func (AbilityScore) Fields() []ent.Field {
	return []ent.Field{
		field.String("record_id").NotEmpty().MaxRuneLen(20).Immutable().
			Comment("owning attempt (join into test_record.id) — plain id, not an edge: soft-delete cascade and future cross-service aggregation both favor the loose join"),
		field.String("user_id").NotEmpty().MaxRuneLen(20).Immutable().
			Comment("denormalized owner so cross-attempt aggregation (historical best, same-age percentile) never joins through test_record"),
		field.String("ability").NotEmpty().MaxRuneLen(32).
			Comment("six-word vocabulary reactionSpeed|aimingAccuracy|trackingStability|skillReleaseAccuracy|resourceAllocation|decisionEfficiency — one row per ability an attempt feeds"),
		field.Float("score").
			Comment("the ability contribution this attempt scored — written by the E3 scoring engine, never the client"),
		field.Float("percentile").Optional().Nillable().
			Comment("same-age percentile once the E5 cohort math lands; null until then"),
	}
}

func (AbilityScore) Edges() []ent.Edge {
	return nil
}

func (AbilityScore) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
		local_mixins.TenantMixin{},
		local_mixins.SoftDeleteMixin{},
	}
}

func (AbilityScore) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "created_at"),
		index.Fields("record_id", "ability").Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("user_id", "ability", "created_at"),
	}
}

func (AbilityScore) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "ability_score"},
	}
}
