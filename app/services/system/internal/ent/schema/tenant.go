package schema

import (
	"regexp"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"
)

const codeJoinKeyPattern = `^[a-z][a-z0-9-]{0,18}[a-z0-9]$`

type Tenant struct {
	ent.Schema
}

func (Tenant) Fields() []ent.Field {
	return []ent.Field{
		field.String("code").NotEmpty().MaxRuneLen(20).Immutable().
			Match(regexp.MustCompile(codeJoinKeyPattern)).
			Comment("join key: tenant-scoped rows carry this value in tenant_id"),
		field.String("name").NotEmpty().MaxRuneLen(128),
		field.Bool("enabled").Default(true).
			Comment("false rejects the tenant's inhabitants at login and verification (a check, not a revocation)"),
	}
}

func (Tenant) Edges() []ent.Edge {
	return nil
}

func (Tenant) Mixin() []ent.Mixin {
	// No SoftDelete: a deleted row must not hold the code unique hostage.
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
	}
}

func (Tenant) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code").Unique(),
	}
}

func (Tenant) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "tenant"},
	}
}
