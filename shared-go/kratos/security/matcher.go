// Package security holds the access-policy primitives shared across the auth,
// authz, and datascope subpackages: MatchAccess drives the selector, and
// DefaultGuard catches unannotated business RPCs.
package security

import (
	"context"
	"strings"
	"sync"

	"github.com/go-kratos/kratos/v3/middleware/selector"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	extv1 "cyber-ecosystem/gen/go/cyber/ext/v1"
)

var accessCache sync.Map // operation string → extv1.Access

// AccessOf resolves the cyber.ext.v1.access annotation of a Kratos operation
// ("/pkg.Svc/Method") and returns it verbatim — no defaulting. An unset
// annotation, a missing extension, a non-method descriptor, or an unknown
// operation all yield ACCESS_UNSPECIFIED; the caller decides how to route it.
// Framework handlers (health/reflection) bypass the middleware chain entirely,
// so they never reach this. Results are cached per operation.
func AccessOf(operation string) extv1.Access {
	if v, ok := accessCache.Load(operation); ok {
		return v.(extv1.Access)
	}
	level := resolveAccess(operation)
	accessCache.Store(operation, level)
	return level
}

func resolveAccess(operation string) extv1.Access {
	// "/pkg.Svc/Method" → "pkg.Svc.Method" (protoreflect.FullName)
	name := strings.ReplaceAll(strings.TrimPrefix(operation, "/"), "/", ".")
	desc, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(name))
	if err != nil {
		return extv1.Access_ACCESS_UNSPECIFIED
	}
	md, ok := desc.(protoreflect.MethodDescriptor)
	if !ok {
		return extv1.Access_ACCESS_UNSPECIFIED
	}
	opts := md.Options()
	if opts == nil || !proto.HasExtension(opts, extv1.E_Access) {
		return extv1.Access_ACCESS_UNSPECIFIED
	}
	level, ok := proto.GetExtension(opts, extv1.E_Access).(extv1.Access)
	if !ok {
		return extv1.Access_ACCESS_UNSPECIFIED
	}
	return level
}

// MatchAccess returns a selector.MatchFunc that reports whether an operation's
// Access annotation equals the given level. Route one middleware chain per
// audience:
//
//	selector.Server(mw...).Match(security.MatchAccess(extv1.Access_ACCESS_ADMIN)).Build()
//
// Only business RPCs reach the middleware chain (framework handlers like
// health/reflection bypass it), so each level maps to exactly the business
// RPCs explicitly annotated with it.
func MatchAccess(want extv1.Access) selector.MatchFunc {
	return func(_ context.Context, operation string) bool {
		return AccessOf(operation) == want
	}
}
