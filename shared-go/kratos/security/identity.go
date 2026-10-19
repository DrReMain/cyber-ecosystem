package security

import "context"

// Subject is the authenticated identity carried through the request
// context: written once by the auth middleware, read by tenant resolution,
// authorization, and the data mixins.
type Subject struct {
	UserID     string
	TenantID   string
	Population string
	Plane      string
	SessionID  string
	Extra      map[string]any
}

type subjectKey struct{}

func WithSubject(ctx context.Context, s *Subject) context.Context {
	return context.WithValue(ctx, subjectKey{}, s)
}

type registryPlaneKey struct{}

// RegistryPlane lifts subject-derived narrowing (tenant, datascope) for one
// read; write hooks reject the plane — registry reads never write scoped rows.
func RegistryPlane(ctx context.Context) context.Context {
	return context.WithValue(ctx, registryPlaneKey{}, true)
}

func IsRegistryPlane(ctx context.Context) bool {
	_, plane := ctx.Value(registryPlaneKey{}).(bool)
	return plane
}

func SubjectFromCtx(ctx context.Context) (*Subject, bool) {
	s, _ := ctx.Value(subjectKey{}).(*Subject)
	if s == nil {
		return nil, false
	}
	return s, true
}
