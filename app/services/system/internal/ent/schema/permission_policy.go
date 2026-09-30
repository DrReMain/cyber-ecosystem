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

// Permission-to-policy link; a permission survives only when ALL linked
// policies pass. Rows hard-delete together with their permission.
type PermissionPolicy struct {
	ent.Schema
}

func (PermissionPolicy) Fields() []ent.Field {
	return []ent.Field{
		field.String("permission_id").NotEmpty().MaxRuneLen(20).Immutable(),
		field.String("policy_id").NotEmpty().MaxRuneLen(20).Immutable(),
	}
}

func (PermissionPolicy) Edges() []ent.Edge {
	return nil
}

func (PermissionPolicy) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
		local_mixins.TenantMixin{},
	}
}

func (PermissionPolicy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("permission_id", "policy_id").Unique(),
	}
}

func (PermissionPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "permission_policy"},
	}
}
