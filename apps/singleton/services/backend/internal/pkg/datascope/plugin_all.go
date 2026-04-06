package datascope

type AllScopePlugin struct{}

func (p *AllScopePlugin) Type() string { return "all" }

func (p *AllScopePlugin) ValidateConfig(string) error { return nil }

func (p *AllScopePlugin) Merge(_ RoleScope, _ *ScopeSnapshot, result *EffectiveScope) error {
	result.IsAll = true
	return nil
}
