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

type Dept struct {
	ent.Schema
}

func (Dept) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().MaxRuneLen(64),
		field.String("parent_id").Optional().Nillable().MaxRuneLen(20).
			Comment("null = tree root"),
		field.String("remark").Default("").MaxRuneLen(255),
	}
}

func (Dept) Edges() []ent.Edge {
	return nil
}

func (Dept) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
		local_mixins.TenantMixin{},
		local_mixins.DatascopeMixin{DeptField: "id"}, // No UserField: self is not expressible on org rows
		local_mixins.SoftDeleteMixin{},
	}
}

func (Dept) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "parent_id", "name").Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),
	}
}

func (Dept) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "dept"},
	}
}
