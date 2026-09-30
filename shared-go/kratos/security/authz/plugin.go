package authz

import (
	"context"

	"cyber-ecosystem/shared-go/kratos/security"
)

// Policy kinds match the proto PolicyParams oneof case names verbatim; the
// kind is derived server-side from the oneof case, never submitted.
const (
	PolicyKindTimeWindow = "time_window"
	PolicyKindCalendar   = "calendar"
)

// AttrTimeNow carries the evaluation-moment wall clock; plugins read it from
// attrs instead of the global clock so evaluation stays a pure function.
const AttrTimeNow = "time.now"

// CalendarBase values — the periodic default a calendar policy falls back to
// when no override matches the day.
const (
	CalendarBaseAll      = "ALL"
	CalendarBaseWeekdays = "WEEKDAYS"
	CalendarBaseNone     = "NONE"
)

// AttributeSource resolves subject/request attributes for ABAC evaluation.
// Resolve is called lazily — only when a matched policy declares one of the
// source's keys — so implementations may assume a policy context and need
// no eager work of their own.
type AttributeSource interface {
	Keys() []string
	Resolve(ctx context.Context, subject *security.Subject) (map[string]any, error)
}

// ConstraintPolicy is a registered policy-kind plugin. Implementations are
// compiled into the binary; params arrive from stored configuration and
// carry data only, never logic. A permission survives only when every
// linked policy of its grant evaluates true.
type ConstraintPolicy interface {
	Kind() string
	// Attrs lists the attribute keys this kind consumes; the engine resolves
	// only sources whose keys intersect the union over matched grants, each
	// source at most once per request.
	Attrs() []string
	Evaluate(ctx context.Context, attrs map[string]any, params map[string]any) (bool, error)
}
