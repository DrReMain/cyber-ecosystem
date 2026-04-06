package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"cyber-ecosystem/shared-go/orm/ent/mixins"

	"cyber-ecosystem/apps/singleton/services/backend/internal/data/ent/schema/local_mixins"
)

type WorkReport struct {
	ent.Schema
}

func (WorkReport) Fields() []ent.Field {
	return []ent.Field{
		field.String("title").NotEmpty().MaxLen(200),
		field.Text("content").Default(""),
		field.String("type").NotEmpty().MaxLen(20).Comment("daily/weekly/monthly"),
		field.Int("access_level").Default(1).Min(1).Max(5).Comment("security level 1-5"),
		field.String("region").Optional().Nillable().MaxLen(50),
		field.String("status").Default("draft").MaxLen(20).Comment("draft/submitted/approved"),
	}
}

func (WorkReport) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
		local_mixins.DataScopeMixin{DeptField: "department_id", OwnerField: "created_by"},
	}
}

func (WorkReport) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "work_report"},
	}
}
