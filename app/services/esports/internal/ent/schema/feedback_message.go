package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"
)

// FeedbackMessage is one turn of a feedback conversation — the initial
// submission included, so the detail reads as one ordered list.
// Append-only by construction: every column is immutable, no row is
// ever rewritten or deleted, because the conversation is the record of
// what was asked and answered. State lives on the root; a turn's
// arrival and the root's state move commit together. Global like its
// root — no tenant column, matching the ticket it belongs to.
type FeedbackMessage struct {
	ent.Schema
}

func (FeedbackMessage) Fields() []ent.Field {
	return []ent.Field{
		field.String("feedback_id").NotEmpty().MaxLen(20).Immutable().
			Comment("the conversation root this turn belongs to"),
		field.Enum("sender").Values("user", "admin").Immutable().
			Comment("who spoke this turn"),
		field.String("content").NotEmpty().MaxLen(2000).Immutable().
			Comment("the turn's text"),
		field.String("admin_id").MaxLen(20).Optional().Nillable().Immutable().
			Comment("the operator behind an admin turn; absent on member turns"),
	}
}

func (FeedbackMessage) Edges() []ent.Edge {
	return nil
}

func (FeedbackMessage) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
	}
}

func (FeedbackMessage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("feedback_id", "created_at"),
	}
}

func (FeedbackMessage) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "feedback_message"},
	}
}
