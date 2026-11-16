package guards

import (
	"path/filepath"
	"strings"
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
	schemas := map[string]interface{ Mixin() []ent.Mixin }{
		"User": schema.User{}, "Dept": schema.Dept{}, "Role": schema.Role{}, "Permission": schema.Permission{},
		"AuthzPolicy": schema.AuthzPolicy{}, "PermissionPolicy": schema.PermissionPolicy{},
		"PrincipalRole": schema.PrincipalRole{}, "Population": schema.Population{},
		"AuditLog": schema.AuditLog{}, "File": schema.File{}, "AuditExport": schema.AuditExport{},
		"Tenant": schema.Tenant{}, "Membership": schema.Membership{},
		"Credential": schema.Credential{}, "AuthzStamp": schema.AuthzStamp{},
	}
	for name, s := range schemas {
		last := 0
		for _, m := range s.Mixin() {
			if r := rank(m); r > 0 {
				if r <= last {
					t.Errorf("%s: behavioral mixins out of order (want Tenant → Datascope → SoftDelete → Stamp)", name)
				}
				last = r
			}
		}
	}
}

// TestSchemaCensus pins the guard's schema list to the schema directory:
// a new schema file without a guard entry fails here by name, so the order
// law above can never silently lose coverage.
func TestSchemaCensus(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("..", "ent", "schema", "*.go"))
	if err != nil || len(files) == 0 {
		t.Fatalf("schema glob: %v (files=%d)", err, len(files))
	}
	// Guarded set is keyed by schema file base name (policy.go declares
	// AuthzPolicy — the Go name avoids ent's predeclared "Policy").
	guarded := map[string]bool{
		"user": true, "dept": true, "role": true, "permission": true, "policy": true,
		"permission_policy": true, "principal_role": true, "population": true,
		"audit_log": true, "file": true, "audit_export": true, "tenant": true,
		"membership": true, "credential": true, "authz_stamp": true,
	}
	for _, f := range files {
		base := strings.TrimSuffix(filepath.Base(f), ".go")
		if base == "generate" {
			continue
		}
		if !guarded[base] {
			t.Errorf("schema file %s.go has no mixin-order guard entry — add it to TestMixinOrder", base)
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
