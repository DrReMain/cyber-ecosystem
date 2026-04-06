package datascope

type SelfScopePlugin struct{}

func (p *SelfScopePlugin) Type() string { return "self" }

func (p *SelfScopePlugin) ValidateConfig(string) error { return nil }

func (p *SelfScopePlugin) Merge(_ RoleScope, _ *ScopeSnapshot, result *EffectiveScope) error {
	result.SelfFilter = true
	return nil
}
