package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"
)

// UserFeedback is the ticket root of one feedback conversation: what
// the list rows render and the state machine the conversation walks —
// pending → replied, a member follow-up walking replied back to
// pending, closed terminal. Content is denormalized as the
// conversation's first message so the lists need no join; the
// conversation itself lives in feedback_message rows appended in lock
// step with every state move. Same global shape as agreement_consent —
// no tenant column and no tenant filtering: the ticket addresses a
// roster identity with no tenant axis.
type UserFeedback struct {
	ent.Schema
}

func (UserFeedback) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id").NotEmpty().MaxLen(20).Immutable().
			Comment("the asking member identity"),
		field.Enum("category").Values("feature_bug", "assessment", "report", "payment", "account", "other").Immutable().
			Comment("the submit form's six problem types"),
		field.String("content").NotEmpty().MaxLen(500).Immutable().
			Comment("the first message of the conversation, denormalized for the list rows"),
		field.String("contact").MaxLen(128).Optional().Nillable().Immutable().
			Comment("where the operator may reach the member; absent = in-app contact only"),
		field.Enum("status").Values("pending", "replied", "closed").Default("pending").
			Comment("the conversation's state: follow-up walks replied back to pending; closed is terminal"),
	}
}

func (UserFeedback) Edges() []ent.Edge {
	return nil
}

func (UserFeedback) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
	}
}

func (UserFeedback) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "updated_at"),
		index.Fields("status", "created_at"),
	}
}

func (UserFeedback) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "user_feedback"},
	}
}
