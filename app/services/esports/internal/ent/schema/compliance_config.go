package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"
)

// ComplianceConfig is the singleton row behind the operator's compliance
// knobs — birth-date collection, the pay age line, and the two pay-wall
// prompt overrides. The slot key pins exactly one row; an absent row means
// every knob answers its shipped default (the column defaults here, the
// fold constants in module/compliance), so the table only ever carries
// what an operator actually set. Like report_policy the knobs are global —
// one operator plane governs every tenant — so the row carries no tenant
// column and no tenant filtering.
type ComplianceConfig struct {
	ent.Schema
}

func (ComplianceConfig) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("slot").Values("singleton").Immutable().
			Comment("the upsert pin — this table's one and only row"),
		field.Bool("birth_collect").Default(true).
			Comment("collect the member's birth date at registration; off = age band only (the pay wall then guides the backfill)"),
		field.Int("pay_age_line").Default(18).
			Comment("members younger than this cannot pay; age reckons on the maximum boundary — born Jan 1 means one year older only on Jan 2"),
		field.String("minor_prompt").Default("").
			Comment("pay-wall copy for an under-age member; empty = each face renders its own localized default"),
		field.String("missing_birth_prompt").Default("").
			Comment("pay-wall copy when the birth date is missing; empty = localized default"),
		field.String("updated_by").NotEmpty().MaxRuneLen(20).
			Comment("admin identity that last set any knob (audit join)"),
	}
}

func (ComplianceConfig) Edges() []ent.Edge {
	return nil
}

func (ComplianceConfig) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
	}
}

func (ComplianceConfig) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("slot").Unique(),
	}
}

func (ComplianceConfig) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "compliance_config"},
	}
}
