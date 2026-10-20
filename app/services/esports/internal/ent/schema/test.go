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

type Test struct {
	ent.Schema
}

func (Test) Fields() []ent.Field {
	return []ent.Field{
		field.String("code").NotEmpty().MaxRuneLen(50).Immutable().
			Comment("immutable catalog code join key (e.g. red_dot); engines, seeds, and cross-domain references key off the code, same law as game codes"),
		field.String("name").NotEmpty().MaxRuneLen(100).
			Comment("display title; seeds carry the placeholder code until the metric sheet calibrates"),
		field.String("category").Optional().Nillable().MaxRuneLen(16).
			Comment("vocabulary fps|moba|universal — universal feeds every genre pool; open word set, admin-managed, classification is content not contract"),
		field.String("ability").Optional().Nillable().MaxRuneLen(32).
			Comment("primary ability, six-word vocabulary reactionSpeed|aimingAccuracy|trackingStability|skillReleaseAccuracy|resourceAllocation|decisionEfficiency; attribution is metric-sheet data pending calibration"),
		field.String("description").Optional().Nillable().MaxRuneLen(500).
			Comment("short catalog blurb"),
		field.String("unit_score_label").Optional().Nillable().MaxRuneLen(50).
			Comment("display form of the unit score (rendering copy only; the numeric itself never crosses the contract)"),
		field.Int("duration_seconds").Default(30).
			Comment("canonical test length — the single source TestConfig mirrors so engine boot is one call; 30 is the placeholder until the metric sheet calibrates"),
		field.Text("introduction").Optional().Nillable().
			Comment("long-form how-to copy, per test; copy changes are content, never contract"),
	}
}

func (Test) Edges() []ent.Edge {
	return nil
}

func (Test) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
		local_mixins.TenantMixin{},
		local_mixins.SoftDeleteMixin{},
	}
}

func (Test) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "created_at"),
		index.Fields("code").Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),
	}
}

func (Test) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "test"},
	}
}
