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

type Sku struct {
	ent.Schema
}

func (Sku) Fields() []ent.Field {
	return []ent.Field{
		field.String("code").NotEmpty().MaxRuneLen(50).Immutable().
			Comment("catalog code join key (e.g. member_month); orders snapshot it, the row ULID never crosses a boundary"),
		field.Enum("kind").Values("trial", "membership").Immutable().
			Comment("what the sku sells: one trial's report permanently vs a membership window"),
		field.String("name").NotEmpty().MaxRuneLen(100).
			Comment("display title (pay sheet line); also the channel order description source"),
		field.String("description").Optional().Nillable().MaxRuneLen(200).
			Comment("pay-sheet blurb under the title"),
		field.Int64("price_fen").Positive().
			Comment("selling price in fen (1/100 yuan) — money stays integer fen end to end"),
		field.Int64("list_price_fen").Positive().
			Comment("list (pre-discount) price; equals price_fen while no discount is shown, carried so discounting stays a value change"),
		field.Int("period_days").Optional().Nillable().Positive().
			Comment("membership window in days (31: the 32nd day expires); null on trial skus"),
		field.Bool("on_shelf").Default(true).
			Comment("member-plane visibility switch; off-shelf rows stay for order-history joins"),
		field.String("updated_by").Optional().Nillable().MaxRuneLen(20).
			Comment("operator id of the last catalog write (audit join); null on seed rows"),
	}
}

func (Sku) Edges() []ent.Edge {
	return nil
}

func (Sku) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
		local_mixins.TenantMixin{},
		local_mixins.SoftDeleteMixin{},
	}
}

func (Sku) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code").Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),
	}
}

func (Sku) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "sku"},
	}
}
