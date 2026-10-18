package bootstrap

import (
	"testing"

	"cyber-ecosystem/shared-go/kratos/security"

	"cyber-ecosystem/app/services/system/internal/conf"
)

// The declared list is deploy-time vocabulary: the shapes that would break
// the registry's invariants — an empty code, the protected default, an
// off-vocabulary archetype, a duplicate — fail the validation wholesale,
// never the boot halfway through.
func TestValidateDeclaredPopulations(t *testing.T) {
	row := func(code, archetype string) *conf.Seed_Population {
		return &conf.Seed_Population{Code: code, Name: code, Archetype: archetype}
	}
	for _, tc := range []struct {
		name string
		rows []*conf.Seed_Population
		want bool
	}{
		{"empty list is legal", nil, true},
		{"one valid patron row", []*conf.Seed_Population{row("member", security.ArchetypePatron)}, true},
		{"one valid engine row", []*conf.Seed_Population{row("staff", security.ArchetypeEngine)}, true},
		{"empty code", []*conf.Seed_Population{row("", security.ArchetypePatron)}, false},
		{"protected default", []*conf.Seed_Population{row(security.DefaultPopulation, security.ArchetypePatron)}, false},
		{"off-vocabulary archetype", []*conf.Seed_Population{row("member", "sympath")}, false},
		{"duplicate code", []*conf.Seed_Population{row("member", security.ArchetypePatron), row("member", security.ArchetypeEngine)}, false},
	} {
		err := validateDeclaredPopulations(tc.rows)
		if (err == nil) != tc.want {
			t.Fatalf("%s: err = %v, want ok=%v", tc.name, err, tc.want)
		}
	}
}
