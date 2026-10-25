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

// ActivationCode is a type code: redeeming opens one report family for
// the redeemer. Codes are single-use (generation mints quantity=1) and
// the single use is taken by the conditional UPDATE in the redeem path —
// used_count < quantity is the whole arbitration, the same shape the
// settle path gives money.
type ActivationCode struct {
	ent.Schema
}

func (ActivationCode) Fields() []ent.Field {
	return []ent.Field{
		field.String("code").NotEmpty().MaxLen(64).Immutable().
			Comment("operator-handout code string (prefix + secure random); stored clear because support re-copies it"),
		field.Enum("report_type").Values("cognitive", "career", "school").Immutable().
			Comment("the report family this code opens — the pricing ruling retired single-record codes"),
		field.Int("quantity").Default(1).Immutable().
			Comment("redeems allowed; generation mints single-use codes, the field keeps the redeem guard general"),
		field.Int("used_count").Default(0).
			Comment("redeems taken; the conditional UPDATE increments under the quantity guard"),
		field.Time("expires_at").
			Comment("code validity deadline; a redeem after it refuses even with uses left"),
		field.String("created_by").NotEmpty().MaxRuneLen(20).Immutable().
			Comment("admin identity that minted the batch (audit join)"),
	}
}

func (ActivationCode) Edges() []ent.Edge {
	return nil
}

func (ActivationCode) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
		local_mixins.TenantMixin{},
	}
}

func (ActivationCode) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code").Unique(),
		index.Fields("report_type", "created_at"),
	}
}

func (ActivationCode) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "activation_code"},
	}
}
