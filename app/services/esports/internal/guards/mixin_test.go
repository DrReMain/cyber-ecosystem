package guards

import (
	"testing"

	"entgo.io/ent"

	"cyber-ecosystem/app/services/esports/internal/ent/schema"
	"cyber-ecosystem/app/services/esports/internal/ent/schema/local_mixins"
)

// The behavioral mixins must land in a fixed order on every table that
// carries them (tenant first, soft delete after) — the read interceptor
// and the write hooks both depend on it. Every schema with a local mixin
// is listed, hard-delete tables included, so a new table joins the guard
// the moment it opts in; ActivationCode and its peers carry TenantMixin
// alone, which the ranking tolerates by design.
func TestMixinOrder(t *testing.T) {
	rank := func(m ent.Mixin) int {
		switch m.(type) {
		case local_mixins.TenantMixin:
			return 1
		case local_mixins.SoftDeleteMixin:
			return 2
		}
		return 0
	}
	schemas := []interface{ Mixin() []ent.Mixin }{
		schema.Profile{},
		schema.ActivationCode{},
		schema.AbilityScore{},
		schema.MatchConfig{},
		schema.Game{},
		schema.Entitlement{},
		schema.MatchSession{},
		schema.QuestionBank{},
		schema.Sku{},
		schema.TestConfig{},
		schema.PayOrder{},
		schema.TestRecord{},
		schema.ReportSnapshot{},
		schema.Test{},
	}
	for _, s := range schemas {
		last := 0
		for _, m := range s.Mixin() {
			if r := rank(m); r > 0 {
				if r <= last {
					t.Fatalf("%T: behavioral mixins out of order (want Tenant → SoftDelete)", s)
				}
				last = r
			}
		}
	}
}
