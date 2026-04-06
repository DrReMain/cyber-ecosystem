package datascope

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

type AttributeScopePlugin struct{}

func (p *AttributeScopePlugin) Type() string { return "attribute" }

func (p *AttributeScopePlugin) ValidateConfig(config string) error {
	if config == "" {
		return nil
	}
	var cfg ScopeConfig
	if err := json.Unmarshal([]byte(config), &cfg); err != nil {
		return fmt.Errorf("unmarshal scope_config: %w", err)
	}
	if len(cfg.Rules) == 0 {
		return errors.New("attribute scope config has no rules")
	}
	validOps := map[string]bool{"eq": true, "neq": true, "gt": true, "gte": true, "lt": true, "lte": true, "in": true}
	validCastTypes := map[string]bool{"text": true, "numeric": true, "": true}
	for _, r := range cfg.Rules {
		if r.Field == "" {
			return errors.New("attribute scope rule missing field")
		}
		if !validOps[r.Op] {
			return fmt.Errorf("attribute scope rule invalid op: %q", r.Op)
		}
		if !validCastTypes[r.CastType] {
			return fmt.Errorf("attribute scope rule invalid castType: %q", r.CastType)
		}
		if r.CastType == "numeric" && r.Value != "" {
			if _, err := strconv.ParseFloat(r.Value, 64); err != nil {
				return fmt.Errorf("attribute scope rule numeric castType with non-numeric value: %q", r.Value)
			}
		}
	}
	return nil
}

func (p *AttributeScopePlugin) Merge(scope RoleScope, snap *ScopeSnapshot, result *EffectiveScope) error {
	result.AttributeFilter = true
	if scope.ScopeConfig == "" {
		return nil
	}
	var cfg ScopeConfig
	if err := json.Unmarshal([]byte(scope.ScopeConfig), &cfg); err != nil {
		return nil
	}
	for i := range cfg.Rules {
		if cfg.Rules[i].ValueSource == "user_attribute" && cfg.Rules[i].ValueAttr != "" {
			if v, ok := snap.Attributes[cfg.Rules[i].ValueAttr]; ok {
				cfg.Rules[i].Value = v
			}
		}
	}
	result.Rules = append(result.Rules, cfg.Rules...)
	if cfg.Logic != "" {
		result.Logic = cfg.Logic
	}
	return nil
}
