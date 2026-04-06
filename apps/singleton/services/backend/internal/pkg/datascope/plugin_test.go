package datascope

import "testing"

func TestAllScopePlugin_ValidateConfig(t *testing.T) {
	p := &AllScopePlugin{}
	if err := p.ValidateConfig(""); err != nil {
		t.Errorf("AllScopePlugin.ValidateConfig('') = %v, want nil", err)
	}
	if err := p.ValidateConfig("{}"); err != nil {
		t.Errorf("AllScopePlugin.ValidateConfig('{}') = %v, want nil", err)
	}
}

func TestSelfScopePlugin_ValidateConfig(t *testing.T) {
	p := &SelfScopePlugin{}
	if err := p.ValidateConfig(""); err != nil {
		t.Errorf("SelfScopePlugin.ValidateConfig('') = %v, want nil", err)
	}
}

func TestDeptScopePlugin_ValidateConfig(t *testing.T) {
	p := &DeptScopePlugin{}
	if err := p.ValidateConfig(""); err != nil {
		t.Errorf("DeptScopePlugin.ValidateConfig('') = %v, want nil", err)
	}
}

func TestAttributeScopePlugin_ValidateConfig(t *testing.T) {
	p := &AttributeScopePlugin{}
	tests := []struct {
		name    string
		config  string
		wantErr bool
	}{
		{"valid", `{"rules":[{"field":"status","op":"eq","value":"active"}]}`, false},
		{"empty_rules", `{"rules":[]}`, true},
		{"missing_field", `{"rules":[{"op":"eq","value":"active"}]}`, true},
		{"invalid_op", `{"rules":[{"field":"status","op":"contains","value":"active"}]}`, true},
		{"empty_config", "", false},
		{"invalid_json", "{invalid}", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := p.ValidateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("AttributeScopePlugin.ValidateConfig(%q) error = %v, wantErr %v", tt.config, err, tt.wantErr)
			}
		})
	}
}

func TestScopePluginRegistry(t *testing.T) {
	r := NewScopePluginRegistry()
	r.Register(&AllScopePlugin{})
	p, ok := r.Get("all")
	if !ok || p == nil {
		t.Fatal("expected plugin to be found")
	}
	_, ok = r.Get("nonexistent")
	if ok {
		t.Fatal("expected plugin not to be found")
	}
	if err := r.Validate("all", ""); err != nil {
		t.Errorf("Validate(all, '') = %v, want nil", err)
	}
	if err := r.Validate("nonexistent", ""); err == nil {
		t.Error("Validate(nonexistent, '') = nil, want error")
	}
}

func TestBuiltinScopePlugins(t *testing.T) {
	r := NewBuiltinScopePluginRegistry()
	expected := []string{"all", "self", "dept", "attribute"}
	for _, k := range expected {
		if _, ok := r.Get(k); !ok {
			t.Errorf("registry missing plugin %q", k)
		}
	}
}

func TestAllScopePlugin_Merge(t *testing.T) {
	p := &AllScopePlugin{}
	result := &EffectiveScope{}
	scope := RoleScope{ScopeType: "all"}
	if err := p.Merge(scope, &ScopeSnapshot{}, result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsAll {
		t.Error("expected IsAll = true")
	}
}

func TestSelfScopePlugin_Merge(t *testing.T) {
	p := &SelfScopePlugin{}
	result := &EffectiveScope{}
	scope := RoleScope{ScopeType: "self"}
	if err := p.Merge(scope, &ScopeSnapshot{}, result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.SelfFilter {
		t.Error("expected SelfFilter = true")
	}
}

func TestDeptScopePlugin_Merge(t *testing.T) {
	p := &DeptScopePlugin{}
	result := &EffectiveScope{}
	snap := &ScopeSnapshot{DeptIDs: []string{"d1", "d2"}}
	scope := RoleScope{ScopeType: "dept"}
	if err := p.Merge(scope, snap, result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.DeptFilter {
		t.Error("expected DeptFilter = true")
	}
	if len(result.DeptIDs) != 2 {
		t.Errorf("expected 2 DeptIDs, got %d", len(result.DeptIDs))
	}
}

func TestAttributeScopePlugin_Merge(t *testing.T) {
	p := &AttributeScopePlugin{}
	result := &EffectiveScope{}
	snap := &ScopeSnapshot{Attributes: map[string]string{"department": "engineering"}}
	scope := RoleScope{
		ScopeType:   "attribute",
		ScopeConfig: `{"rules":[{"field":"dept","op":"eq","valueSource":"user_attribute","valueAttr":"department"}],"logic":"and"}`,
	}
	if err := p.Merge(scope, snap, result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.AttributeFilter {
		t.Error("expected AttributeFilter = true")
	}
	if len(result.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(result.Rules))
	}
	if result.Rules[0].Value != "engineering" {
		t.Errorf("expected value 'engineering', got %q", result.Rules[0].Value)
	}
}
