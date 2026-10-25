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

// PayOrder is the local projection of a channel payment. Money rows are
// permanent audit records: no soft delete, and the state machine's only
// arbiter is the conditional UPDATE in the settle path — pending → paid,
// pending → closed, and the truth correction closed → paid (a converger
// close racing a late payment loses; money taken must be granted).
type PayOrder struct {
	ent.Schema
}

func (PayOrder) Fields() []ent.Field {
	return []ent.Field{
		field.String("order_no").NotEmpty().MaxLen(32).Immutable().
			Comment("business order number ('P'+ULID) — the channel idempotency key (out_trade_no) and the audit join key into bills"),
		field.String("user_id").NotEmpty().MaxRuneLen(20).Immutable().
			Comment("buying member identity (system user id), logical reference — same key law as profile"),
		field.String("sku_code").NotEmpty().MaxRuneLen(50).Immutable().
			Comment("sku catalog code at purchase time (snapshot; later sku edits never rewrite history)"),
		field.Enum("scope_kind").Values("trial", "period").Immutable().
			Comment("grant shape: trial unlocks one test record (scope_ref), period opens the membership window"),
		field.String("scope_ref").Optional().Nillable().MaxRuneLen(32).Immutable().
			Comment("trial: the test record id being unlocked; period: null — the window shape comes from the sku"),
		field.String("provider").NotEmpty().MaxRuneLen(16).Immutable().
			Comment("payment provider name (wechat, mock) — the key the platform's container map routes query/close by"),
		field.String("channel").NotEmpty().MaxRuneLen(32).Immutable().
			Comment("access-context channel value (provider vocabulary, e.g. wechat_jsapi); fixes how the credential is consumed"),
		field.Int64("amount_fen").Positive().Immutable().
			Comment("agreed amount in fen, frozen at creation; settlement verifies the channel-reported amount against it"),
		field.Enum("state").Values("pending", "paid", "closed").Default("pending").
			Comment("channel-state projection; paid is terminal — refunds are derived facts, not states"),
		field.String("credential_kind").NotEmpty().MaxRuneLen(16).Immutable().
			Comment("how the page consumes the credential (invoke|jump_url|qr_code), snapshotted with the value"),
		field.String("credential_value").NotEmpty().MaxRuneLen(1024).Immutable().
			Comment("channel credential at creation; same-scope reuse replays it with zero channel round trips"),
		field.String("transaction_id").Optional().Nillable().MaxRuneLen(64).
			Comment("channel transaction id, backfilled at settlement — the bill-reconciliation join key"),
		field.Time("expires_at").Immutable().
			Comment("payment deadline carried from creation (channel default two hours); expiry closes pending orders"),
		field.Time("paid_at").Optional().Nillable().
			Comment("channel-side payment time; the entitlement window starts here for membership"),
	}
}

func (PayOrder) Edges() []ent.Edge {
	return nil
}

func (PayOrder) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
		local_mixins.TenantMixin{},
	}
}

func (PayOrder) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_no").Unique(),
		// Same-grant reuse guard: one live pending order per (user,
		// scope). Trial scope_refs are non-null so the guard binds; a
		// period's null ref never conflicts (NULLS DISTINCT) — concurrent
		// membership creations may both live and both pay, each purchase
		// stacking its own grant at settle.
		index.Fields("tenant_id", "user_id", "scope_kind", "scope_ref").Unique().
			Annotations(entsql.IndexWhere("state = 'pending'")),
		index.Fields("state", "expires_at"),
		index.Fields("tenant_id", "user_id", "created_at"),
	}
}

func (PayOrder) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "pay_order"},
	}
}
