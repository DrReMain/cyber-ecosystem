package security

import "context"

// Subject is the authenticated identity carried through the request
// context: written once by the auth middleware, read by tenant resolution,
// authorization, and the data mixins.
type Subject struct {
	UserID    string
	TenantID  string
	SessionID string
	Extra     map[string]any
}

type subjectKey struct{}

func WithSubject(ctx context.Context, s *Subject) context.Context {
	return context.WithValue(ctx, subjectKey{}, s)
}

func SubjectFromCtx(ctx context.Context) (*Subject, bool) {
	s, ok := ctx.Value(subjectKey{}).(*Subject)
	return s, ok
}
