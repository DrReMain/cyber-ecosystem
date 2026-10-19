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

type QuestionBank struct {
	ent.Schema
}

func (QuestionBank) Fields() []ent.Field {
	return []ent.Field{
		field.String("test_code").NotEmpty().MaxRuneLen(50).Immutable().
			Comment("owning test (join into test.code) — the scenario answer key lives with the test it grades"),
		field.String("scenario_code").NotEmpty().MaxRuneLen(64).
			Comment("scenario join key — the identifier attempts echo back in their decision facts"),
		field.JSON("payload", map[string]any{}).Optional().
			Comment("question face — scenario content for rendering; shape is test-owned content, not contract"),
		field.JSON("answer", map[string]any{}).
			Comment("answer key — correctness of submitted choices is judged against this server-side, never asserted by the client"),
	}
}

func (QuestionBank) Edges() []ent.Edge {
	return nil
}

func (QuestionBank) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
		local_mixins.TenantMixin{},
		local_mixins.SoftDeleteMixin{},
	}
}

func (QuestionBank) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "created_at"),
		index.Fields("test_code", "scenario_code").Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),
	}
}

func (QuestionBank) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "question_bank"},
	}
}
