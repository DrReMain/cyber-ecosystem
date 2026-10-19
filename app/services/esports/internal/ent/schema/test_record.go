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

type TestRecord struct {
	ent.Schema
}

func (TestRecord) Fields() []ent.Field {
	return []ent.Field{
		field.String("test_code").NotEmpty().MaxRuneLen(50).Immutable().
			Comment("the attempted test (join into test.code)"),
		field.String("user_id").NotEmpty().MaxRuneLen(20).Immutable().
			Comment("owning member identity (system user id), logical reference — same key law as profile"),
		field.Int("attempt_no").Positive().
			Comment("attempt ordinal inside the match session; uniqueness is the E4 session law, not a DB constraint yet"),
		field.String("session_id").Optional().Nillable().MaxRuneLen(64).
			Comment("match session reference — null until the E4 session aggregate lands, then backfilled"),
		field.Time("started_at").Immutable().
			Comment("attempt start (wall clock, server-stamped from the submitted anchor)"),
		field.Time("ended_at").Immutable().
			Comment("attempt end"),
		field.JSON("backgrounds", []types.BackgroundSpan{}).Optional().
			Comment("engine-marked backgrounding spans; validity is judged against test_config.background_invalid_seconds"),
		field.JSON("raw_detail", map[string]any{}).
			Comment("the oneof fact stream verbatim (e.g. {'red_dot': {'events': [...]}}) — facts only, every aggregate is recomputed from it; the detail kind is the map's single key"),
		field.Float("raw_score").Optional().Nillable().
			Comment("linear uncapped track — null on invalid attempts (nothing scores)"),
		field.Float("authoritative_score").Optional().Nillable().
			Comment("percentile-tiered authoritative track — null on invalid attempts"),
		field.Enum("grade").Optional().Nillable().Values("S", "A", "B", "C").
			Comment("percentile tier — top 5 / 15 / 40 percent / rest; enum-bounded at the ent layer so drift fails the write, not the report"),
		field.Bool("valid").Default(true).
			Comment("server-side validity verdict — background over budget or incomplete detail invalidates the attempt"),
		field.String("invalid_reason").Optional().Nillable().MaxRuneLen(64).
			Comment("ops vocabulary code when valid=false (user copy resolves client-side)"),
	}
}

func (TestRecord) Edges() []ent.Edge {
	return nil
}

func (TestRecord) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
		local_mixins.TenantMixin{},
		local_mixins.SoftDeleteMixin{},
	}
}

func (TestRecord) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "created_at"),
		index.Fields("user_id", "test_code", "created_at"),
		index.Fields("test_code", "created_at"),
	}
}

func (TestRecord) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "test_record"},
	}
}
