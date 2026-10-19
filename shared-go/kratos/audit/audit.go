// Package audit collects operation outcomes for the audit trail: the Audit
// middleware emits allow-side WRITE actions and every authorization denial
// into the Sink port (implemented by the application's audit usecase), and
// the Emitter is the queued background publisher the persistence side drives
// toward the carrier through a PublishFunc.
package audit

import (
	"context"
	"errors"
	"time"

	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"
)

// Event is one audited operation outcome: what was attempted, by whom, and
// how it ended. The JSON shape is the on-wire payload contract shared by the
// publishing and consuming sides.
type Event struct {
	TenantID      string    `json:"tenant_id"`
	Actor         string    `json:"actor"`          // subject user id; empty for subject-less events (login)
	PrincipalType string    `json:"principal_type"` // "user"; empty together with Actor
	Operation     string    `json:"operation"`      // "/cyber.system.v1.UserService/CreateUser"
	HTTPMethod    string    `json:"http_method"`
	HTTPPath      string    `json:"http_path"`
	Status        int       `json:"status"`
	LatencyMs     int64     `json:"latency_ms"`
	IP            string    `json:"ip"`
	UserAgent     string    `json:"user_agent"`
	DenyReason    string    `json:"deny_reason"` // NO_GRANT | ABAC_CONSTRAINT; empty when allowed
	OccurredAt    time.Time `json:"occurred_at"`
}

// Deny reason vocabulary — the authz engine's Decision reasons verbatim.
const (
	DenyNoGrant        = kauthz.DenyNoGrant
	DenyAbacConstraint = kauthz.DenyAbacConstraint
)

// Sink is the port the Audit middleware hands collected events to; the
// application's audit usecase is the intended implementor. Implementations
// must not block the request path.
type Sink interface {
	Emit(ctx context.Context, ev *Event)
}

// The authz deny sentinels captured before any server-side rebinding wraps
// them (app wiring rebinds the package vars onto mapped errors, keeping the
// original as the cause). Identity comparison down the cause chain
// therefore survives the rebinding.
var (
	sentinelPermissionDenied = kauthz.ErrPermissionDenied
	sentinelPolicyDenied     = kauthz.ErrPolicyDenied
)

// denyReasonOf classifies an authorization denial by sentinel identity on
// the error chain. Everything else — success, auth failure, validation,
// business errors — is not a denial and returns "".
func denyReasonOf(err error) string {
	for e := err; e != nil; e = errors.Unwrap(e) {
		switch {
		case errors.Is(e, sentinelPermissionDenied):
			return DenyNoGrant
		case errors.Is(e, sentinelPolicyDenied):
			return DenyAbacConstraint
		}
	}
	return ""
}
