package utils

import "strings"

const (
	ASC       = "asc"
	DESC      = "desc"
	SEPARATOR = ":"
)

type OrderBy struct {
	Field string
	Order string
}

func (ob *OrderBy) FieldString() string {
	return ob.Field
}

func (ob *OrderBy) OrderString() string {
	return ob.Order
}

func (ob *OrderBy) ASC() string {
	return ASC
}

func (ob *OrderBy) DESC() string {
	return DESC
}

func ParseOrderBy(orderBy []string) []*OrderBy {
	if len(orderBy) == 0 {
		return nil
	}

	rules := make([]*OrderBy, 0, len(orderBy))
	for _, s := range orderBy {
		parts := strings.Split(s, SEPARATOR)
		if len(parts) != 2 {
			continue
		}

		field := strings.TrimSpace(parts[0])
		order := strings.ToLower(strings.TrimSpace(parts[1]))

		if order != ASC && order != DESC {
			continue
		}

		rules = append(rules, &OrderBy{Field: field, Order: order})
	}
	return rules
}

func StringifyOrderBy(orderBy []*OrderBy) []string {
	if len(orderBy) == 0 {
		return nil
	}

	result := make([]string, 0, len(orderBy))
	for _, ob := range orderBy {
		result = append(result, ob.Field+SEPARATOR+ob.Order)
	}
	return result
}
