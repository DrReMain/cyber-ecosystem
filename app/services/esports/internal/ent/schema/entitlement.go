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

// Entitlement is the grant ledger, written only inside the settle
// transaction: a paid order inserts exactly its one grant row. Trial rows
// are permanent (an unlock never lapses); membership rows carry their
// window and stack across renewals. The unique key lands the ledger law:
// a trial unlocks once regardless of who or what paid for it, while
// membership rows hold a null ref (NULLS DISTINCT) so each purchase is
// its own row.
type Entitlement struct {
	ent.Schema
}

func (Entitlement) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id").NotEmpty().MaxRuneLen(20).Immutable().
			Comment("holding member identity (system user id), logical reference — same key law as profile"),
		field.Enum("scope_kind").Values("trial", "membership", "report_type").Immutable().
			Comment("grant shape: trial points at one record permanently, membership opens a window, report_type opens a report family via a type code"),
		field.String("scope_ref").Optional().Nillable().MaxRuneLen(32).Immutable().
			Comment("trial: the unlocked test record id; report_type: the family name (dedup guard); membership: null — renewal rows stack instead of dedup"),
		field.Time("expires_at").Optional().Nillable().Immutable().
			Comment("membership window end (paid_at + period days); null on trial rows — permanence is the absence of a deadline"),
		field.String("source").NotEmpty().MaxRuneLen(64).Immutable().
			Comment("granting artifact id — the order number today; activation codes join this column when they land"),
	}
}

func (Entitlement) Edges() []ent.Edge {
	return nil
}

func (Entitlement) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
		local_mixins.TenantMixin{},
	}
}

func (Entitlement) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "user_id", "scope_kind", "scope_ref").Unique(),
	}
}

func (Entitlement) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "entitlement"},
	}
}
