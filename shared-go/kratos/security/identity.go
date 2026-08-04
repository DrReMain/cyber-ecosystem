package security

import "context"

type Subject struct {
	UserID    string
	TenantID  string
	SessionID string
}

type subjectKey struct{}

func WithSubject(ctx context.Context, s *Subject) context.Context {
	return context.WithValue(ctx, subjectKey{}, s)
}

func SubjectFromCtx(ctx context.Context) (*Subject, bool) {
	s, ok := ctx.Value(subjectKey{}).(*Subject)
	return s, ok
}
