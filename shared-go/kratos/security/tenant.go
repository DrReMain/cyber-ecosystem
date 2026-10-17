package security

import "context"

// DefaultTenant is the tenant used when a context carries no subject or the
// subject has no tenant (internal paths, framework handlers). It is also the
// platform tenant — the seeded registry row whose seat faces the platform
// plane — so the default bucket and the platform plane share one constant.
const DefaultTenant = "default"

type tenantKey struct{}

// WithTenant pins the tenancy domain on the context: an explicit switch that
// outranks the subject. Row routing (the tenant mixin) and bucket routing
// (per-tenant storage) resolve through this one precedence, so a pass over a
// domain hits that domain's rows and that domain's bucket in the same pass.
func WithTenant(ctx context.Context, tenant string) context.Context {
	return context.WithValue(ctx, tenantKey{}, tenant)
}

// ExplicitTenantFromCtx reports the WithTenant switch, if set to a domain.
func ExplicitTenantFromCtx(ctx context.Context) (string, bool) {
	t, ok := ctx.Value(tenantKey{}).(string)
	if !ok || t == "" {
		return "", false
	}
	return t, true
}

// TenantFromCtx resolves the request's tenant: an explicit domain switch
// wins over the subject, and absent both it falls back to DefaultTenant, so
// it always returns a usable value.
func TenantFromCtx(ctx context.Context) string {
	if t, ok := ExplicitTenantFromCtx(ctx); ok {
		return t
	}
	if s, ok := SubjectFromCtx(ctx); ok && s.TenantID != "" {
		return s.TenantID
	}
	return DefaultTenant
}
