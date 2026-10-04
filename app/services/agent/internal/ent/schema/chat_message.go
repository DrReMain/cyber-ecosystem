package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"

	"cyber-ecosystem/app/services/agent/internal/ent/schema/local_mixins"
)

type ChatMessage struct {
	ent.Schema
}

func (ChatMessage) Fields() []ent.Field {
	return []ent.Field{
		field.String("session_id").NotEmpty().MaxRuneLen(20).
			Comment("owning chat_session id; no FK — ownership is enforced by querying through the owner-filtered session"),
		field.String("role").Default("").MaxRuneLen(16).
			Comment("OpenAI-compatible role vocabulary; free-form so S4 tool roles need no migration"),
		field.Text("content").Default("").
			Comment("turn text; unbounded by design — parts-JSONB is the additive S4+ evolution for multimodal"),
		field.Text("reasoning").Default("").
			Comment("accumulated model reasoning on the assistant turn; empty on user turns"),
		field.String("model").Default("").MaxRuneLen(128).
			Comment("model id that produced the assistant turn; empty on user turns"),
		field.String("finish").Default("").MaxRuneLen(16).
			Comment("upstream finish_reason on the assistant turn (stop/length/aborted); empty on user turns"),
	}
}

func (ChatMessage) Edges() []ent.Edge {
	return nil
}

func (ChatMessage) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
		local_mixins.TenantMixin{},
		local_mixins.SoftDeleteMixin{},
	}
}

func (ChatMessage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("session_id", "created_at"),
	}
}

func (ChatMessage) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "chat_message"},
	}
}
