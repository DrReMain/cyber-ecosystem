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

// ActivationBatch is one generation call's mint — the unit an operator
// disables, exports, and reads burn-down by. The row carries only
// batch-level truth; count facts are aggregated over the code rows at
// read time, so the two can never drift.
type ActivationBatch struct {
	ent.Schema
}

func (ActivationBatch) Fields() []ent.Field {
	return []ent.Field{
		field.String("no").NotEmpty().MaxLen(32).Immutable().
			Comment("human-shape batch number (B-YYMM-XXXX) — the string support quotes"),
		field.Enum("report_type").Values("cognitive", "career", "school").Immutable().
			Comment("the report family every code in the batch opens"),
		field.Time("expires_at").
			Comment("code validity deadline shared by the whole batch"),
		field.String("created_by").NotEmpty().MaxRuneLen(20).Immutable().
			Comment("admin identity that minted the batch (audit join)"),
		field.Time("disabled_at").Optional().Nillable().
			Comment("null = live batch; set = the moment the operator disabled it — unredeemed codes void on the spot, landed grants stay landed"),
	}
}

func (ActivationBatch) Edges() []ent.Edge {
	return nil
}

func (ActivationBatch) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
		local_mixins.TenantMixin{},
	}
}

func (ActivationBatch) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("no").Unique(),
		index.Fields("report_type", "created_at"),
	}
}

func (ActivationBatch) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "activation_batch"},
	}
}
