package security

import "context"

// Subject is the authenticated identity carried through the request
// context: written once by the auth middleware, read by tenant resolution,
// authorization, and the data mixins.
type Subject struct {
	UserID     string
	TenantID   string
	Population string
	Archetype  string
	SessionID  string
	Extra      map[string]any
}

type subjectKey struct{}

func WithSubject(ctx context.Context, s *Subject) context.Context {
	return context.WithValue(ctx, subjectKey{}, s)
}

type registryPlaneKey struct{}

// RegistryPlane marks a read as registry-plane: subject-derived narrowing
// (tenant, datascope) lifts for its query — the registry reads scoped rows
// as plain data, identity discovery by a global key being the shape. Write
// hooks reject the plane outright: a registry read must never birth or
// mutate a scoped row, and the silent default-tenant back-fill is exactly
// the bug that law prevents.
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
