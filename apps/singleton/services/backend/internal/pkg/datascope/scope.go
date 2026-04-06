package datascope

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"

	"cyber-ecosystem/apps/singleton/services/backend/internal/pkg/security"
)

// EffectiveScope is the resolved data permission scope for a user on a specific operation.
type EffectiveScope struct {
	IsAll           bool
	SelfFilter      bool
	DeptFilter      bool
	AttributeFilter bool
	DeptIDs         []string
	Rules           []FilterRule
	Logic           string
	ExtraPredicates []func(*sql.Selector)
}

// FilterRule is a single data permission filter rule.
type FilterRule struct {
	Field       string `json:"field"`
	Op          string `json:"op"`
	ValueSource string `json:"valueSource"`
	ValueAttr   string `json:"valueAttr,omitempty"`
	Value       string `json:"value,omitempty"`
	CastType    string `json:"castType,omitempty"` // "text" (default) or "numeric"
}

// RoleScope is a data scope rule associated with a role.
type RoleScope struct {
	RoleCode       string
	ScopeType      string
	ScopeConfig    string
	TargetResource string
}

// ScopeSnapshot is the cached data permission snapshot per user.
type ScopeSnapshot struct {
	Roles      []string          `json:"roles"`
	Scopes     []RoleScope       `json:"scopes"`
	DeptIDs    []string          `json:"dept_ids"`
	Attributes map[string]string `json:"attributes"`
	CachedAt   time.Time         `json:"cached_at"`
}

// ScopeConfig is the JSON structure stored in scope_config field.
type ScopeConfig struct {
	Rules []FilterRule `json:"rules"`
	Logic string       `json:"logic"`
}

// ScopeResolveFunc is a function type for lazy scope resolution.
// Middleware injects a closure; Ent mixin calls it at query time.
type ScopeResolveFunc func(ctx context.Context, userID, operation string) (*EffectiveScope, error)

// Context Keys ---------------------------------------------------------------------------------------------------------

type scopeResolverKey struct{}
type scopeUserIDKey struct{}
type skipDataScopeKey struct{}

// WithScopeResolver stores the ScopeResolveFunc in context.
func WithScopeResolver(ctx context.Context, fn ScopeResolveFunc) context.Context {
	return context.WithValue(ctx, scopeResolverKey{}, fn)
}

// ScopeResolverFromContext retrieves the ScopeResolveFunc from context.
func ScopeResolverFromContext(ctx context.Context) (ScopeResolveFunc, bool) {
	fn, ok := ctx.Value(scopeResolverKey{}).(ScopeResolveFunc)
	return fn, ok
}

// WithScopeUserID stores the user ID in context.
func WithScopeUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, scopeUserIDKey{}, userID)
}

// ScopeUserIDFromContext retrieves the user ID from context.
func ScopeUserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(scopeUserIDKey{}).(string)
	return id, ok
}

// SkipDataScope returns a context that bypasses data permission filtering.
func SkipDataScope(ctx context.Context) context.Context {
	return context.WithValue(ctx, skipDataScopeKey{}, true)
}

// SkipDataScopeFromContext checks if data permission filtering should be skipped.
func SkipDataScopeFromContext(ctx context.Context) bool {
	b, _ := ctx.Value(skipDataScopeKey{}).(bool)
	return b
}

// Utility ---------------------------------------------------------------------------------------------------------------

// MatchResource delegates to security.MatchResource.
func MatchResource(pattern, operation string) bool {
	return security.MatchResource(pattern, operation)
}

// UniqueStrings deduplicates a string slice.
func UniqueStrings(s []string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(s))
	for _, v := range s {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

const scopeSnapshotPrefix = "data_scope:"

// SnapshotCacheKey returns the cache key for a user's scope snapshot.
func SnapshotCacheKey(userID string) string {
	return fmt.Sprintf("%s%s", scopeSnapshotPrefix, userID)
}

// MarshalSnapshot serializes a snapshot to JSON bytes.
func MarshalSnapshot(snapshot *ScopeSnapshot) ([]byte, error) {
	return json.Marshal(snapshot)
}

// UnmarshalSnapshot deserializes a snapshot from JSON bytes.
func UnmarshalSnapshot(data []byte) (*ScopeSnapshot, error) {
	var snapshot ScopeSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, err
	}
	return &snapshot, nil
}
