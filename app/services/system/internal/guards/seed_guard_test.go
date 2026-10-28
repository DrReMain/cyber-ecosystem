package guards

import (
	"testing"

	"cyber-ecosystem/shared-go/kratos/security"

	"cyber-ecosystem/app/services/system/internal/bootstrap"
	"cyber-ecosystem/app/services/system/internal/conf"
)

// The declared list is deploy-time vocabulary: the shapes that would break
// the registry's invariants — an empty code, the protected default, an
// off-vocabulary plane, a duplicate — fail the validation wholesale,
// never the boot halfway through.
func TestValidateDeclaredPopulations(t *testing.T) {
	row := func(code, plane string) *conf.Seed_Population {
		return &conf.Seed_Population{Code: code, Name: code, Plane: plane}
	}
	for _, tc := range []struct {
		name string
		rows []*conf.Seed_Population
		want bool
	}{
		{"empty list is legal", nil, true},
		{"one valid member row", []*conf.Seed_Population{row("member", security.PlaneMember)}, true},
		{"one valid backend row", []*conf.Seed_Population{row("staff", security.PlaneBackend)}, true},
		{"empty code", []*conf.Seed_Population{row("", security.PlaneMember)}, false},
		{"protected default", []*conf.Seed_Population{row(security.DefaultPopulation, security.PlaneBackend)}, false},
		{"off-vocabulary plane", []*conf.Seed_Population{row("member", "sympath")}, false},
		{"duplicate code", []*conf.Seed_Population{row("member", security.PlaneMember), row("member", security.PlaneBackend)}, false},
	} {
		err := bootstrap.ValidateDeclaredPopulations(tc.rows)
		if (err == nil) != tc.want {
			t.Fatalf("%s: err = %v, want ok=%v", tc.name, err, tc.want)
		}
	}
}
