package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"
)

// UserMessage is one row of one member's letterbox. Unread is this
// row's own state, deliberately server-side: the badge asks the
// server, mark-read flips the row, and the count stays true across
// devices. Rows arrive from three writers — announcement publish (the
// delivery copy), the feedback reply closure (same transaction as the
// operator's answer), and the report/membership note writers when they
// wire in — and never leave: there is no delete face, a delivered
// letter is a fact. Content is denormalized per row on purpose: the
// announcement a letter copied may go offline later, the letter keeps
// what was sent. Same global shape as agreement_consent — no tenant
// column and no tenant filtering: the letterbox addresses roster
// identities, which carry no tenant axis.
type UserMessage struct {
	ent.Schema
}

func (UserMessage) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id").NotEmpty().MaxLen(20).Immutable().
			Comment("the receiving member identity"),
		field.Enum("type").Values("announcement", "report", "membership", "feedback_reply").Immutable().
			Comment("what kind of letter this is; decides the ref's domain (report→test record, feedback_reply→feedback)"),
		field.String("title").NotEmpty().MaxLen(128).Immutable().
			Comment("the headline the sender composed; every writer sets one, no fallback shape exists"),
		field.String("content").NotEmpty().MaxLen(5000).Immutable().
			Comment("the letter body, denormalized at delivery so later source changes never rewrite a sent letter"),
		field.String("ref_id").MaxLen(20).Optional().Nillable().Immutable().
			Comment("deep-link target whose domain the type names; absent when the letter carries no link"),
		field.Bool("is_read").Default(false).
			Comment("the row's own unread state — the only mutable column, flipped by mark-read"),
	}
}

func (UserMessage) Edges() []ent.Edge {
	return nil
}

func (UserMessage) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
	}
}

func (UserMessage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "is_read"),
		index.Fields("user_id", "created_at"),
	}
}

func (UserMessage) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "user_message"},
	}
}
