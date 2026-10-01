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

type AuditExport struct {
	ent.Schema
}

func (AuditExport) Fields() []ent.Field {
	return []ent.Field{
		field.String("file_id").NotEmpty().MaxRuneLen(20).
			Comment("logical reference to the file row; status/size live there"),
		field.String("owner_id").NotEmpty().MaxRuneLen(20).
			Comment("initiating user snapshot; datascope self dimension"),
	}
}

func (AuditExport) Edges() []ent.Edge {
	return nil
}

func (AuditExport) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
		local_mixins.TenantMixin{},
		local_mixins.DatascopeMixin{UserField: "owner_id"},
	}
}

func (AuditExport) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "created_at"),
		index.Fields("file_id").Unique(),
	}
}

func (AuditExport) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "audit_export"},
	}
}
