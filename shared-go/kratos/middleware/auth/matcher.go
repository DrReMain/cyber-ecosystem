package auth

import (
	"context"
	"strings"
	"sync"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"github.com/go-kratos/kratos/v2/middleware/selector"

	"cyber-ecosystem/contracts/go/auth"
)

// NewWhiteListMatcher creates a selector.MatchFunc that determines whether
// a given operation requires authentication based on the proto-defined
// (auth.public_access) option. Operations marked with public_access=true
// will skip authentication (MatchFunc returns false).
func NewWhiteListByPublicAccessInProtoMatcher() selector.MatchFunc {
	cache := sync.Map{}

	return func(ctx context.Context, operation string) bool {
		// Check cache first
		if val, ok := cache.Load(operation); ok {
			return val.(bool)
		}

		// Convert operation format from "/package.Service/Method" to "package.Service.Method"
		// Kratos uses "/package.Service/Method" but protoregistry uses "package.Service.Method"
		protoPath := strings.ReplaceAll(strings.TrimPrefix(operation, "/"), "/", ".")

		// Find the method descriptor in the global registry
		desc, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(protoPath))

		// Default: require auth if descriptor not found or no public_access option
		shouldAuth := true

		if err == nil {
			// Type assert to MethodDescriptor
			methodDesc, ok := desc.(protoreflect.MethodDescriptor)
			if !ok {
				cache.Store(operation, true)
				return true
			}

			// Check if the method has the public_access extension
			if proto.HasExtension(methodDesc.Options(), auth.E_PublicAccess) {
				isPublic := proto.GetExtension(methodDesc.Options(), auth.E_PublicAccess).(bool)
				// If public_access=true, skip auth (return false)
				// If public_access=false, require auth (return true)
				shouldAuth = !isPublic
			}
		}

		cache.Store(operation, shouldAuth)
		return shouldAuth
	}
}
