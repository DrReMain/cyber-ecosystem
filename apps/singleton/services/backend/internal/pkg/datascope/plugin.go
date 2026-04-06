package datascope

import "fmt"

// ScopePlugin is a self-contained scope type plugin.
type ScopePlugin interface {
	// Type returns the unique scope type identifier (e.g. "all", "dept").
	Type() string
	// ValidateConfig validates the JSON config at write time.
	ValidateConfig(config string) error
	// Merge merges this scope's contribution into the effective scope result.
	Merge(scope RoleScope, snap *ScopeSnapshot, result *EffectiveScope) error
}

// ScopePluginRegistry manages scope plugins.
type ScopePluginRegistry struct {
	plugins map[string]ScopePlugin
}

// NewScopePluginRegistry creates an empty registry.
func NewScopePluginRegistry() *ScopePluginRegistry {
	return &ScopePluginRegistry{plugins: make(map[string]ScopePlugin)}
}

// Register adds a plugin.
func (r *ScopePluginRegistry) Register(p ScopePlugin) {
	r.plugins[p.Type()] = p
}

// Get returns the plugin for a scope type.
func (r *ScopePluginRegistry) Get(scopeType string) (ScopePlugin, bool) {
	p, ok := r.plugins[scopeType]
	return p, ok
}

// Validate delegates to the plugin's ValidateConfig.
func (r *ScopePluginRegistry) Validate(scopeType, config string) error {
	p, ok := r.plugins[scopeType]
	if !ok {
		return fmt.Errorf("unknown scope_type: %q", scopeType)
	}
	return p.ValidateConfig(config)
}

// Merge delegates to the plugin's Merge.
func (r *ScopePluginRegistry) Merge(scope RoleScope, snap *ScopeSnapshot, result *EffectiveScope) error {
	p, ok := r.plugins[scope.ScopeType]
	if !ok {
		return fmt.Errorf("unknown scope_type: %q", scope.ScopeType)
	}
	return p.Merge(scope, snap, result)
}

// NewBuiltinScopePluginRegistry creates a registry with all built-in plugins registered.
func NewBuiltinScopePluginRegistry() *ScopePluginRegistry {
	r := NewScopePluginRegistry()
	for _, p := range BuiltinScopePlugins() {
		r.Register(p)
	}
	return r
}

// BuiltinScopePlugins returns the default set of scope plugins.
func BuiltinScopePlugins() []ScopePlugin {
	return []ScopePlugin{
		&AllScopePlugin{},
		&SelfScopePlugin{},
		&DeptScopePlugin{},
		&AttributeScopePlugin{},
	}
}
