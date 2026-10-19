// Package security holds the access-policy primitives shared across its
// auth and authz subpackages: the Subject identity carried in request
// contexts, tenant resolution and the population archetype vocabulary,
// audience matching for middleware routing, the guard that fails
// unannotated business RPCs closed, and the platform-plane structural
// guard.
package security
