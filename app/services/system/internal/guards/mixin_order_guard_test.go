package guards

import (
	"testing"

	"entgo.io/ent"

	"cyber-ecosystem/app/services/system/internal/ent/schema"
	"cyber-ecosystem/app/services/system/internal/ent/schema/local_mixins"
)

func TestMixinOrder(t *testing.T) {
	rank := func(m ent.Mixin) int {
		switch m.(type) {
		case local_mixins.TenantMixin:
			return 1
		case local_mixins.DatascopeMixin:
			return 2
		case local_mixins.SoftDeleteMixin:
			return 3
		case local_mixins.AuthzStampMixin:
			return 4
		}
		return 0
	}
	schemas := []interface{ Mixin() []ent.Mixin }{
		schema.User{}, schema.Dept{}, schema.Role{}, schema.Permission{}, schema.AuthzPolicy{},
		schema.PermissionPolicy{}, schema.PrincipalRole{}, schema.Population{}, schema.AuditLog{}, schema.File{},
		schema.AuditExport{},
	}
	for _, s := range schemas {
		last := 0
		for _, m := range s.Mixin() {
			if r := rank(m); r > 0 {
				if r <= last {
					t.Fatalf("%T: behavioral mixins out of order (want Tenant → Datascope → SoftDelete → Stamp)", s)
				}
				last = r
			}
		}
	}
}

// Dimension fields compile into raw SQL columns; a typo narrows nothing.
func TestDatascopeDimensionFields(t *testing.T) {
	mixinsFor := map[string][]ent.Mixin{
		"User":        schema.User{}.Mixin(),
		"Dept":        schema.Dept{}.Mixin(),
		"File":        schema.File{}.Mixin(),
		"AuditExport": schema.AuditExport{}.Mixin(),
	}
	for _, c := range []struct {
		name  string
		mixin local_mixins.DatascopeMixin
	}{
		{"User", local_mixins.DatascopeMixin{UserField: "id", DeptField: "dept_id"}},
		{"Dept", local_mixins.DatascopeMixin{DeptField: "id"}},
		{"File", local_mixins.DatascopeMixin{UserField: "owner_id"}},
		{"AuditExport", local_mixins.DatascopeMixin{UserField: "owner_id"}},
	} {
		found := false
		for _, m := range mixinsFor[c.name] {
			if d, ok := m.(local_mixins.DatascopeMixin); ok {
				found = true
				if d != c.mixin {
					t.Fatalf("%s: DatascopeMixin = %+v, want %+v", c.name, d, c.mixin)
				}
			}
		}
		if !found {
			t.Fatalf("%s: DatascopeMixin missing", c.name)
		}
	}
}
