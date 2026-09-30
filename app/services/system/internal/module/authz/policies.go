package authz

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"

	"cyber-ecosystem/shared-go/kratos/security"
	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"
)

const (
	policyPassed = "passed"
	policyFailed = "failed"
)

var policyPlugins = map[string]kauthz.ConstraintPolicy{}

var attributeSources = []kauthz.AttributeSource{timeSource{}}

func registerPolicy(p kauthz.ConstraintPolicy) {
	// A duplicate kind would silently override the first registration; a
	// boot-time panic turns that into a build mistake caught immediately.
	if _, dup := policyPlugins[p.Kind()]; dup {
		panic("authz: duplicate policy kind: " + p.Kind())
	}
	policyPlugins[p.Kind()] = p
}

func decodeParams(params map[string]any, target any) error {
	// A JSON roundtrip lands map-decoded numbers (float64) in typed fields —
	// plugins never read raw map values; dirt from direct DB writes surfaces
	// here and fails closed upstream.
	raw, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("encode params: %w", err)
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return fmt.Errorf("decode params: %w", err)
	}
	return nil
}

func survivingGrants(ctx context.Context, log *slog.Logger, subject *security.Subject, grants []GrantView) []GrantView {
	if !anyLinkedPolicy(grants) {
		return grants // no linked policy → zero attribute resolution
	}
	attrs, err := resolveAttributes(ctx, subject, neededAttrKeys(grants))
	if err != nil {
		log.Error("authz: attribute resolution failed; constraints fail closed", "error", err)
		return nil
	}
	out := make([]GrantView, 0, len(grants))
	for _, g := range grants {
		if passed, _ := evalGrantPolicies(ctx, log, attrs, g.Policies); passed {
			out = append(out, g)
		}
	}
	return out
}

func evalGrantPolicies(ctx context.Context, log *slog.Logger, attrs map[string]any, polys []PolicyView) (bool, []PolicyState) {
	// No short-circuit: explain wants every policy's state. One failed policy
	// eliminates the grant (AND semantics).
	states := make([]PolicyState, 0, len(polys))
	passed := true
	for _, p := range polys {
		state := PolicyState{Kind: p.Kind, Name: p.Name, State: policyPassed}
		plugin, registered := policyPlugins[p.Kind]
		if !registered {
			log.Error("authz: unregistered policy kind fails closed", "kind", p.Kind)
			state.State = policyFailed
		} else if pass, err := plugin.Evaluate(ctx, attrs, p.Params); err != nil {
			log.Warn("authz: policy evaluation failed; fails closed", "kind", p.Kind, "error", err)
			state.State = policyFailed
		} else if !pass {
			state.State = policyFailed
		}
		if state.State == policyFailed {
			passed = false
		}
		states = append(states, state)
	}
	return passed, states
}

func anyLinkedPolicy(grants []GrantView) bool {
	for _, g := range grants {
		if len(g.Policies) > 0 {
			return true
		}
	}
	return false
}

func neededAttrKeys(grants []GrantView) map[string]struct{} {
	var needed map[string]struct{}
	for _, g := range grants {
		for _, p := range g.Policies {
			plugin, ok := policyPlugins[p.Kind]
			if !ok {
				continue // unknown kinds fail closed at evaluation
			}
			for _, k := range plugin.Attrs() {
				if needed == nil {
					needed = make(map[string]struct{})
				}
				needed[k] = struct{}{}
			}
		}
	}
	return needed
}

func resolveAttributes(ctx context.Context, subject *security.Subject, needed map[string]struct{}) (map[string]any, error) {
	var attrs map[string]any
	for _, src := range attributeSources {
		var hit bool
		for _, k := range src.Keys() {
			if _, ok := needed[k]; ok {
				hit = true
				break
			}
		}
		if !hit {
			continue
		}
		resolved, err := src.Resolve(ctx, subject)
		if err != nil {
			return nil, fmt.Errorf("resolve attributes: %w", err)
		}
		if attrs == nil {
			attrs = make(map[string]any, len(resolved))
		}
		maps.Copy(attrs, resolved)
	}
	return attrs, nil
}
