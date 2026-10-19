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

type TestConfig struct {
	ent.Schema
}

func (TestConfig) Fields() []ent.Field {
	return []ent.Field{
		field.String("test_code").NotEmpty().MaxRuneLen(50).Immutable().
			Comment("join key into test.code — same aggregate, one row per test, joined plain by code (the contract's join-key law), not an edge: the pair is born together but read and tuned on different cadences"),
		field.Int("duration_seconds").Default(30).
			Comment("mirror of test.duration_seconds — the engine boot bag is self-sufficient behind one config call"),
		field.Int("background_invalid_seconds").Default(10).
			Comment("client-side mechanics threshold: backgrounding longer than this invalidates the run; the engine only marks spans, validity is judged at submit"),
		field.JSON("engine_params", map[string]string{}).Optional().
			Comment("engine tuning bag — spawn intervals, speeds, sizes; keys are content owned by the engine registry (metric spec sheet), the KV shape keeps tuning ops-only"),
	}
}

func (TestConfig) Edges() []ent.Edge {
	return nil
}

func (TestConfig) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
		local_mixins.TenantMixin{},
		local_mixins.SoftDeleteMixin{},
	}
}

func (TestConfig) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "created_at"),
		index.Fields("test_code").Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),
	}
}

func (TestConfig) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "test_config"},
	}
}
