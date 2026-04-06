package security

import "strings"

// MatchResource matches a pattern against an operation string.
// Supports: "*" (match all), "prefix/*" (prefix match), exact match.
func MatchResource(pattern, operation string) bool {
	if pattern == "" || pattern == "*" {
		return true
	}
	prefix, ok := strings.CutSuffix(pattern, "/*")
	if ok {
		return strings.HasPrefix(operation, prefix+"/")
	}
	return pattern == operation
}
