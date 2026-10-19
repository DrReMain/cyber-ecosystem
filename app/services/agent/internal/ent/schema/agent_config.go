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

type AgentConfig struct {
	ent.Schema
}

func (AgentConfig) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id").NotEmpty().MaxRuneLen(20).
			Comment("owning user; one live row per user, enforced by the partial unique index"),
		field.String("base_url").NotEmpty().MaxRuneLen(512).
			Comment("OpenAI-compatible provider base URL"),
		field.String("api_key").Default("").MaxRuneLen(512).
			Comment("AES-GCM ciphertext, base64(nonce||ct); empty = no key set; never returned by any RPC"),
	}
}

func (AgentConfig) Edges() []ent.Edge {
	return nil
}

func (AgentConfig) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
		local_mixins.TenantMixin{},
		local_mixins.SoftDeleteMixin{},
	}
}

func (AgentConfig) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "created_at"),
		index.Fields("user_id").Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),
	}
}

func (AgentConfig) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "agent_config"},
	}
}
