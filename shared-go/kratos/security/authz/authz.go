// Package authz holds the authorization contract shared by engines and
// transports: the Authorizer interface, the Decision it produces, and the
// OperationAuthz middleware that enforces decisions after authentication.
// The ABAC plugin seam and its kind vocabulary live in plugin.go — the one
// file here that grows when a policy kind is added.
package authz

import (
	"context"

	"github.com/go-kratos/kratos/v3/errors"
	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/transport"

	"cyber-ecosystem/shared-go/kratos/security"
)

var (
	ErrMissingSubject   = errors.Unauthorized("MISSING_SUBJECT", "")
	ErrPermissionDenied = errors.Forbidden("PERMISSION_DENIED", "")
	ErrPolicyDenied     = errors.Forbidden("AUTHZ_POLICY_DENIED", "")
)

const (
	DenyNoGrant        = "NO_GRANT"
	DenyAbacConstraint = "ABAC_CONSTRAINT"
	AllowBuiltin       = "BUILTIN"
)

const (
	ScopeKindAll      = "all"
	ScopeKindSelf     = "self"
	ScopeKindDeptTree = "dept_tree"
)

const ScopeParamDeptIDs = "dept_ids"

// Authorizer decides whether a subject may perform an operation. The
// decision is computed per request and never cached across requests.
type Authorizer interface {
	Decide(ctx context.Context, subject *security.Subject, operation string) (*Decision, error)
}

// Decision is the outcome of one authorization check.
type Decision struct {
	Allowed bool
	Reason  string // deny reason; AllowBuiltin on a baseline allow; empty when allowed by grants
	Scopes  []ScopeDescriptor
}

// ScopeDescriptor is the data range attached to a surviving permission.
// Kind: "" (tenant-wide, the absence of narrowing) | all | self | dept_tree.
// Params mirrors the permission's scope_params JSON.
type ScopeDescriptor struct {
	Entity string
	Kind   string
	Params map[string]any
}

type decisionKey struct{}

func withDecision(ctx context.Context, d *Decision) context.Context {
	return context.WithValue(ctx, decisionKey{}, d)
}

func DecisionFromCtx(ctx context.Context) (*Decision, bool) {
	d, ok := ctx.Value(decisionKey{}).(*Decision)
	return d, ok
}

func OperationAuthz(a Authorizer) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			subject, ok := security.SubjectFromCtx(ctx)
			if !ok {
				return nil, ErrMissingSubject
			}
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return nil, ErrMissingSubject
			}
			d, err := a.Decide(ctx, subject, tr.Operation())
			if err != nil {
				return nil, err
			}
			// An engine that returns no decision at all is broken; failing
			// closed beats panicking on the nil deref below.
			if d == nil {
				return nil, ErrPermissionDenied
			}
			if !d.Allowed {
				if d.Reason == DenyAbacConstraint {
					return nil, ErrPolicyDenied
				}
				return nil, ErrPermissionDenied
			}
			return handler(withDecision(ctx, d), req)
		}
	}
}
