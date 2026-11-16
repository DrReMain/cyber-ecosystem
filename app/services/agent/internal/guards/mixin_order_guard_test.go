package guards

import (
	"path/filepath"
	"strings"
	"testing"

	"entgo.io/ent"

	"cyber-ecosystem/app/services/agent/internal/ent/schema"
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
	schemas := map[string]interface{ Mixin() []ent.Mixin }{
		"AgentConfig": schema.AgentConfig{},
		"ChatSession": schema.ChatSession{},
		"ChatMessage": schema.ChatMessage{},
		"AuditLog":    schema.AuditLog{},
	}
	for name, s := range schemas {
		last := 0
		for _, m := range s.Mixin() {
			if r := rank(m); r > 0 {
				if r <= last {
					t.Errorf("%s: behavioral mixins out of order (want Tenant → SoftDelete)", name)
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
	guarded := map[string]bool{"agent_config": true, "chat_session": true, "chat_message": true, "audit_log": true}
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
