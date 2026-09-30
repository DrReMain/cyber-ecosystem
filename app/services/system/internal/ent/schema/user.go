package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"

	"cyber-ecosystem/app/services/system/internal/ent/schema/local_mixins"
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("email").NotEmpty().MaxRuneLen(128),
		field.String("password_hash").Sensitive().MaxRuneLen(128).
			Comment("argon2 digest; plaintext is never stored"),
		field.String("dept_id").Optional().Nillable().MaxRuneLen(20),
		field.String("avatar").Optional().Nillable().MaxRuneLen(20).
			Comment("referenced file id; display URLs are minted at read time"),
		field.Bool("enabled").Default(true).
			Comment("false rejects login and drops grants at engine compile"),
	}
}

func (User) Edges() []ent.Edge {
	return nil
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
		local_mixins.TenantMixin{},
		local_mixins.DatascopeMixin{UserField: "id", DeptField: "dept_id"},
		local_mixins.SoftDeleteMixin{},
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "email").Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),
	}
}

func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "user"},
	}
}
