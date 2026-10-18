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

type Profile struct {
	ent.Schema
}

func (Profile) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id").NotEmpty().MaxRuneLen(20).Immutable().
			Comment("owning member identity (system user id); one live row per member"),
		field.String("nickname").Optional().Nillable().MaxRuneLen(12).
			Comment("display name; 2-12 chars once set (boundary-validated, not a DB concern)"),
		field.Int("birth_year").Optional().Nillable().
			Comment("birth year; year+month is the collected granularity, a day is fabricated precision"),
		field.Int("birth_month").Optional().Nillable().Min(1).Max(12).
			Comment("birth month 1-12"),
		field.String("avatar_url").Optional().Nillable().MaxRuneLen(500).
			Comment("avatar URL; wechat prefill hydrates identity-side, storage lives here"),
		field.Time("guardian_consented_at").Optional().Nillable().
			Comment("under-13 hard-gate stamp, server-written only; null = not consented"),
	}
}

func (Profile) Edges() []ent.Edge {
	return nil
}

func (Profile) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
		local_mixins.TenantMixin{},
		local_mixins.SoftDeleteMixin{},
	}
}

func (Profile) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "created_at"),
		index.Fields("user_id").Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),
	}
}

func (Profile) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "profile"},
	}
}
