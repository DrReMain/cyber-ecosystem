package datascope

type DeptScopePlugin struct{}

func (p *DeptScopePlugin) Type() string { return "dept" }

func (p *DeptScopePlugin) ValidateConfig(string) error { return nil }

func (p *DeptScopePlugin) Merge(_ RoleScope, snap *ScopeSnapshot, result *EffectiveScope) error {
	result.DeptFilter = true
	result.DeptIDs = append(result.DeptIDs, snap.DeptIDs...)
	return nil
}
