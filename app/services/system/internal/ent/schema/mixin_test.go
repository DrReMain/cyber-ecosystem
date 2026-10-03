package schema

import (
	"testing"

	"entgo.io/ent"

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
		}
		return 0
	}
	schemas := []interface{ Mixin() []ent.Mixin }{
		User{}, Dept{}, Role{}, Permission{}, AuthzPolicy{},
		PermissionPolicy{}, PrincipalRole{}, AuditLog{}, File{},
	}
	for _, s := range schemas {
		last := 0
		for _, m := range s.Mixin() {
			if r := rank(m); r > 0 {
				if r <= last {
					t.Fatalf("%T: behavioral mixins out of order (want Tenant → Datascope → SoftDelete)", s)
				}
				last = r
			}
		}
	}
}

// Dimension fields compile into raw SQL columns; a typo narrows nothing.
func TestDatascopeDimensionFields(t *testing.T) {
	for _, c := range []struct {
		name  string
		mixin local_mixins.DatascopeMixin
	}{
		{"User", local_mixins.DatascopeMixin{UserField: "id", DeptField: "dept_id"}},
		{"Dept", local_mixins.DatascopeMixin{DeptField: "id"}},
		{"File", local_mixins.DatascopeMixin{UserField: "owner_id"}},
	} {
		got := User{}.Mixin()
		if c.name == "Dept" {
			got = Dept{}.Mixin()
		}
		if c.name == "File" {
			got = File{}.Mixin()
		}
		found := false
		for _, m := range got {
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
