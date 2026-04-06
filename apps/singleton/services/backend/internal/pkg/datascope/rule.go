package datascope

import (
	"fmt"
	"strings"

	"entgo.io/ent/dialect/sql"
)

// BuildRulePredicates builds predicates from filter rules without applying them.
// Column values are CAST to text before comparison to avoid type mismatch errors
// when comparing integer/other columns against string attribute values.
func BuildRulePredicates(rules []FilterRule) []func(*sql.Selector) {
	preds := make([]func(*sql.Selector), 0, len(rules))
	for _, rule := range rules {
		if p := buildRulePredicate(rule); p != nil {
			preds = append(preds, p)
		}
	}
	return preds
}

func buildRulePredicate(rule FilterRule) func(*sql.Selector) {
	if rule.Value == "" {
		return nil
	}
	castType := rule.CastType
	if castType == "" {
		castType = "text"
	}
	colExpr := fmt.Sprintf(`CAST("%s" AS %s)`, rule.Field, castType)
	switch rule.Op {
	case "eq":
		return func(s *sql.Selector) { s.Where(sql.EQ(colExpr, rule.Value)) }
	case "neq":
		return func(s *sql.Selector) { s.Where(sql.NEQ(colExpr, rule.Value)) }
	case "gt":
		return func(s *sql.Selector) { s.Where(sql.GT(colExpr, rule.Value)) }
	case "gte":
		return func(s *sql.Selector) { s.Where(sql.GTE(colExpr, rule.Value)) }
	case "lt":
		return func(s *sql.Selector) { s.Where(sql.LT(colExpr, rule.Value)) }
	case "lte":
		return func(s *sql.Selector) { s.Where(sql.LTE(colExpr, rule.Value)) }
	case "in":
		parts := strings.Split(rule.Value, ",")
		ifaces := make([]any, len(parts))
		for i, p := range parts {
			ifaces[i] = strings.TrimSpace(p)
		}
		return func(s *sql.Selector) {
			s.Where(sql.In(colExpr, ifaces...))
		}
	default:
		return nil
	}
}
