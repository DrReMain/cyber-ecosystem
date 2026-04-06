package security

import "context"

type SessionValidator interface {
	ValidateSession(ctx context.Context, sessionID string) error
}

type Authorizer interface {
	Authorize(ctx context.Context, subject, operation string) error
}

type ConditionChecker interface {
	CheckConditions(ctx context.Context, subject, operation string) error
}
