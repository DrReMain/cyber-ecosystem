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

// Grant atom: operation pattern + optional row scope + linked policies
// (AND semantics). effect is always "allow" today.
type Permission struct {
	ent.Schema
}

func (Permission) Fields() []ent.Field {
	return []ent.Field{
		field.String("role_id").NotEmpty().MaxRuneLen(20).Immutable(),
		field.String("operation").NotEmpty().MaxRuneLen(255).Immutable().
			Comment("operation pattern: exact match or prefix wildcard ending in /*"),
		field.String("effect").Default("allow").MaxRuneLen(16).
			Comment("reserved; always 'allow' in v1 (no deny)"),
		field.String("scope_kind").Default("").MaxRuneLen(32).
			Comment("row-scope narrowing: '' | all | self | dept_tree; empty = tenant-wide"),
		field.JSON("scope_params", map[string]any{}).Optional().
			Comment("scope payload (e.g. dept_ids); shape per scope_kind"),
	}
}

func (Permission) Edges() []ent.Edge {
	return nil
}

func (Permission) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
		local_mixins.TenantMixin{},
	}
}

func (Permission) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("role_id", "operation").Unique(),
	}
}

func (Permission) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "permission"},
	}
}
