package security

import "context"

// DefaultTenant is the tenant used when a context carries no subject or the
// subject has no tenant (internal paths, framework handlers).
const DefaultTenant = "default"

// TenantFromCtx resolves the request's tenant from the subject. Absent a
// subject or a tenant it falls back to DefaultTenant, so it always returns
// a usable value.
func TenantFromCtx(ctx context.Context) string {
	if s, ok := SubjectFromCtx(ctx); ok && s.TenantID != "" {
		return s.TenantID
	}
	return DefaultTenant
}
