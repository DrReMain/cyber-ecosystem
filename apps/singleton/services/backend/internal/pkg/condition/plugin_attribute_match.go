package condition

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
)

var validAttrOps = map[string]bool{"eq": true, "neq": true, "gt": true, "gte": true, "lt": true, "lte": true, "in": true}

type attributeMatchConfig struct {
	Key    string   `json:"key"`
	Op     string   `json:"op"`
	Value  string   `json:"value,omitempty"`
	Values []string `json:"values,omitempty"`
}

// AttributeMatchPlugin checks if a user attribute matches the configured condition.
type AttributeMatchPlugin struct{}

func (p *AttributeMatchPlugin) Type() string { return "attribute_match" }

func (p *AttributeMatchPlugin) Evaluate(ctx context.Context, config string) (bool, error) {
	cfg, err := p.parseConfig(config)
	if err != nil {
		return false, err
	}
	attrs := UserAttributesFromContext(ctx)
	actual, exists := attrs[cfg.Key]
	if !exists {
		return cfg.Op == "neq", nil
	}
	switch cfg.Op {
	case "eq":
		return actual == cfg.Value, nil
	case "neq":
		return actual != cfg.Value, nil
	case "in":
		return slices.Contains(cfg.Values, actual), nil
	case "gt", "gte", "lt", "lte":
		return compareOrdered(actual, cfg.Value, cfg.Op)
	default:
		return false, fmt.Errorf("attribute_match unsupported op: %q", cfg.Op)
	}
}

func (p *AttributeMatchPlugin) ValidateConfig(config string) error {
	_, err := p.parseConfig(config)
	return err
}

func (p *AttributeMatchPlugin) parseConfig(raw string) (*attributeMatchConfig, error) {
	if raw == "" {
		return nil, fmt.Errorf("attribute_match config is required")
	}
	var cfg attributeMatchConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal attribute_match config: %w", err)
	}
	if cfg.Key == "" {
		return nil, fmt.Errorf("attribute_match config missing key")
	}
	if !validAttrOps[cfg.Op] {
		return nil, fmt.Errorf("attribute_match unsupported op: %q", cfg.Op)
	}
	if cfg.Op == "in" && len(cfg.Values) == 0 {
		return nil, fmt.Errorf("attribute_match op 'in' requires non-empty values")
	}
	return &cfg, nil
}
