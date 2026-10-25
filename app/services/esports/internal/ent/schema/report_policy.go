package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"
)

// ReportPolicy is the operator's access policy for one report family —
// one row per family, upserted in place; an absent row means the shipped
// default, so the table only ever carries overrides. A redeemed code
// outranks whatever the row says. The policy is global — one operator
// plane governs every tenant, so the row carries no tenant column and no
// tenant filtering: a per-tenant index would let one tenant's upsert
// silently rewrite the row another tenant reads.
type ReportPolicy struct {
	ent.Schema
}

func (ReportPolicy) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("report_type").Values("cognitive", "career", "school").Immutable().
			Comment("the family this row governs — the upsert key, one row per family"),
		field.Enum("policy").Values("open", "membership", "tenant", "closed").
			Comment("who may enter: everyone, active members, the tenant's own (awaiting the tenant plane), or only explicit grants"),
		field.String("updated_by").NotEmpty().MaxRuneLen(20).
			Comment("admin identity that last set the policy (audit join)"),
	}
}

func (ReportPolicy) Edges() []ent.Edge {
	return nil
}

func (ReportPolicy) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
	}
}

func (ReportPolicy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("report_type").Unique(),
	}
}

func (ReportPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "report_policy"},
	}
}
