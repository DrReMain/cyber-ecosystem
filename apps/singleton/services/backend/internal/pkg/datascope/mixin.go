package datascope

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

// DataScopeMixin provides department_id and owner_id fields for data permission filtering.
// Interceptors and hooks remain in the app's local_mixins package since they depend on generated types.
type DataScopeMixin struct {
	mixin.Schema
	DeptField  string
	OwnerField string
}

// Fields adds department_id and/or created_by columns to the entity.
func (d DataScopeMixin) Fields() []ent.Field {
	fields := []ent.Field{}
	if d.DeptField != "" {
		fields = append(fields, field.String(d.DeptField).Optional().Nillable().MaxLen(20))
	}
	if d.OwnerField != "" {
		fields = append(fields, field.String(d.OwnerField).Optional().Nillable().MaxLen(20))
	}
	return fields
}

// Indexes adds indexes for the permission fields.
func (d DataScopeMixin) Indexes() []ent.Index {
	indexes := []ent.Index{}
	if d.DeptField != "" {
		indexes = append(indexes, index.Fields(d.DeptField))
	}
	if d.OwnerField != "" {
		indexes = append(indexes, index.Fields(d.OwnerField))
	}
	return indexes
}
