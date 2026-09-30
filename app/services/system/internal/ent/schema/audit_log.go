package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"
)

// Append-only: inserts only, no updates, no soft delete. tenant is a plain
// column so denied and subject-less requests can audit too; the
// subject-derived TenantMixin does not apply.
type AuditLog struct {
	ent.Schema
}

func (AuditLog) Fields() []ent.Field {
	return []ent.Field{
		field.String("tenant").Default("").MaxRuneLen(20).Immutable().
			Comment("subject tenant; empty on denied or subject-less events"),
		field.String("actor").Default("").MaxRuneLen(20).Immutable().
			Comment("subject user id; empty on subject-less events"),
		field.String("principal_type").Default("").MaxRuneLen(16).Immutable().
			Comment("user; empty together with actor"),
		field.String("operation").NotEmpty().MaxRuneLen(255).Immutable().
			Comment("proto operation path, e.g. /cyber.system.v1.UserService/CreateUser"),
		field.String("http_method").Default("").MaxRuneLen(16).Immutable(),
		field.String("http_path").Default("").MaxRuneLen(512).Immutable(),
		field.Int("status").Default(0).Immutable().
			Comment("HTTP status of the outcome"),
		field.Int("latency_ms").Default(0).Immutable(),
		field.String("ip").Default("").MaxRuneLen(64).Immutable(),
		field.String("user_agent").Default("").MaxRuneLen(512).Immutable(),
		field.String("deny_reason").Default("").MaxRuneLen(64).Immutable().
			Comment("NO_GRANT | ABAC_CONSTRAINT; empty when allowed"),
		field.String("event_id").Default("").MaxRuneLen(40).Immutable().
			Comment("publisher-generated dedup key; unique index absorbs redelivery"),
	}
}

func (AuditLog) Edges() []ent.Edge {
	return nil
}

func (AuditLog) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
	}
}

func (AuditLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant", "created_at"),
		index.Fields("event_id").Unique(),
	}
}

func (AuditLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "audit_log"},
	}
}
