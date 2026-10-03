package schema

import (
	"testing"

	"entgo.io/ent"

	"cyber-ecosystem/app/services/agent/internal/ent/schema/local_mixins"
)

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
		AgentConfig{},
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
