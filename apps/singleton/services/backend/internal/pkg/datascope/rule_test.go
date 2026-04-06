package datascope

import "testing"

func TestBuildRulePredicates_EmptyValue(t *testing.T) {
	rules := []FilterRule{{Field: "status", Op: "eq", Value: ""}}
	preds := BuildRulePredicates(rules)
	if len(preds) != 0 {
		t.Errorf("expected 0 predicates for empty value, got %d", len(preds))
	}
}

func TestBuildRulePredicates_UnsupportedOp(t *testing.T) {
	rules := []FilterRule{{Field: "status", Op: "contains", Value: "active"}}
	preds := BuildRulePredicates(rules)
	if len(preds) != 0 {
		t.Errorf("expected 0 predicates for unsupported op, got %d", len(preds))
	}
}

func TestBuildRulePredicates_SupportedOps(t *testing.T) {
	ops := []string{"eq", "neq", "gt", "gte", "lt", "lte", "in"}
	for _, op := range ops {
		t.Run(op, func(t *testing.T) {
			rules := []FilterRule{{Field: "col", Op: op, Value: "val"}}
			preds := BuildRulePredicates(rules)
			if len(preds) != 1 {
				t.Errorf("expected 1 predicate for op %q, got %d", op, len(preds))
			}
		})
	}
}

func TestBuildRulePredicates_CastType(t *testing.T) {
	rules := []FilterRule{{Field: "level", Op: "gt", Value: "3", CastType: "numeric"}}
	preds := BuildRulePredicates(rules)
	if len(preds) != 1 {
		t.Fatalf("expected 1 predicate, got %d", len(preds))
	}
}

func TestBuildRulePredicates_MultipleRules(t *testing.T) {
	rules := []FilterRule{
		{Field: "status", Op: "eq", Value: "active"},
		{Field: "level", Op: "gte", Value: "5", CastType: "numeric"},
		{Field: "region", Op: "in", Value: "us,eu,ap"},
	}
	preds := BuildRulePredicates(rules)
	if len(preds) != 3 {
		t.Errorf("expected 3 predicates, got %d", len(preds))
	}
}
