package condition

import (
	"context"
	"fmt"
)

// ConditionPlugin is a self-contained condition type plugin.
type ConditionPlugin interface {
	// Type returns the unique condition type identifier (e.g. "time_range").
	Type() string
	// Evaluate checks whether the condition passes for the given config.
	Evaluate(ctx context.Context, config string) (bool, error)
	// ValidateConfig validates the JSON config at write time.
	ValidateConfig(config string) error
}

// ConditionRegistry manages condition plugins.
type ConditionRegistry struct {
	plugins map[string]ConditionPlugin
}

// NewConditionRegistry creates an empty registry.
func NewConditionRegistry() *ConditionRegistry {
	return &ConditionRegistry{plugins: make(map[string]ConditionPlugin)}
}

// Register adds a plugin.
func (r *ConditionRegistry) Register(p ConditionPlugin) {
	r.plugins[p.Type()] = p
}

// Get returns the plugin for a condition type.
func (r *ConditionRegistry) Get(conditionType string) (ConditionPlugin, bool) {
	p, ok := r.plugins[conditionType]
	return p, ok
}

// Validate delegates to the plugin's ValidateConfig.
func (r *ConditionRegistry) Validate(conditionType, config string) error {
	p, ok := r.plugins[conditionType]
	if !ok {
		return fmt.Errorf("unknown condition_type: %q", conditionType)
	}
	return p.ValidateConfig(config)
}

// Evaluate delegates to the plugin's Evaluate.
func (r *ConditionRegistry) Evaluate(ctx context.Context, conditionType, config string) (bool, error) {
	p, ok := r.plugins[conditionType]
	if !ok {
		return false, fmt.Errorf("unknown condition_type: %q", conditionType)
	}
	return p.Evaluate(ctx, config)
}

// NewBuiltinConditionRegistry creates a registry with all built-in plugins registered.
func NewBuiltinConditionRegistry() *ConditionRegistry {
	r := NewConditionRegistry()
	for _, p := range BuiltinConditionPlugins() {
		r.Register(p)
	}
	return r
}

// BuiltinConditionPlugins returns the default set of condition plugins.
func BuiltinConditionPlugins() []ConditionPlugin {
	return []ConditionPlugin{
		&TimeRangePlugin{},
		&IPRangePlugin{},
		&DayOfWeekPlugin{},
		&AttributeMatchPlugin{},
	}
}
