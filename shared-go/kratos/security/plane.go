package security

// The plane vocabulary is the security axis's compile-time half (pattern:
// small vocabulary at compile time, instances as data rows). backend = the
// operator world: seats read organizational grants. member = the consumer
// world: seats are structurally denied every gated face — C-end access rides
// the builtin (session-only) rung plus app-domain logic. Platform supers are
// backend seats in the default tenant sealed by the platform guard and a
// seeded super grant; no third identity lane exists.
const (
	PlaneBackend = "backend"
	PlaneMember  = "member"
)

// ValidPlane reports membership in the compile-time vocabulary — the single
// source every plane input validates against.
func ValidPlane(p string) bool {
	return p == PlaneBackend || p == PlaneMember
}
