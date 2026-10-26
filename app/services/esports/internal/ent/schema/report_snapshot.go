package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"

	"cyber-ecosystem/app/services/esports/internal/ent/schema/local_mixins"
	"cyber-ecosystem/app/services/esports/internal/ent/schema/types"
)

type ReportSnapshot struct {
	ent.Schema
}

func (ReportSnapshot) Fields() []ent.Field {
	return []ent.Field{
		field.String("record_id").NotEmpty().MaxRuneLen(20).Immutable().
			Comment("owning attempt (join into test_record.id) — plain id, not an edge: the loose join keeps soft-delete cascade and future aggregation out of this table's way, the ability_score precedent"),
		field.String("user_id").NotEmpty().MaxRuneLen(20).Immutable().
			Comment("denormalized owner, the ability_score precedent — cohort reads never join through test_record"),
		field.Enum("tier").Values("S", "A", "B", "C", "unjudged").
			Comment("the pinned tier badge, judged once at unlock time — the ruling the roadmap froze: an unlocked report never moves with the rubric; unjudged pins the honest no-verdict (every cohort too thin for a percentile), never a made-up letter"),
		field.Float("composite_percentile").Optional().Nillable().
			Comment("the pinned composite (equal-weight mean, missing abilities never zero) — null when no measured ability carried a percentile at pin time"),
		field.JSON("abilities", []types.AbilitySnapshot{}).
			Comment("the pinned radar board: one entry per measured ability (ability, game_code, score, percentile, confidence, sample_count, peer_average) — the exact shape the report face returns"),
		field.String("cohort_basis").NotEmpty().MaxRuneLen(32).
			Comment("the cohort rubric this snapshot was computed under (same-age-same-sex exact, or band-approximate when only the band is known) — the audit stamp that makes drift detectable"),
	}
}

func (ReportSnapshot) Edges() []ent.Edge {
	return nil
}

func (ReportSnapshot) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
		local_mixins.TenantMixin{},
		local_mixins.SoftDeleteMixin{},
	}
}

func (ReportSnapshot) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "record_id").Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("tenant_id", "user_id", "created_at"),
	}
}

func (ReportSnapshot) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "report_snapshot"},
	}
}
