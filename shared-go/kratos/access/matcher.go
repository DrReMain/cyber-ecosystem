package access

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

func NewWhiteListByPublicAccessInProtoMatcher() selector.MatchFunc {
	cache := sync.Map{}

	return func(ctx context.Context, operation string) bool {
		if val, ok := cache.Load(operation); ok {
			return val.(bool)
		}

		protoPath := strings.ReplaceAll(strings.TrimPrefix(operation, "/"), "/", ".")

		desc, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(protoPath))

		shouldAuth := true

		if err == nil {
			methodDesc, ok := desc.(protoreflect.MethodDescriptor)
			if !ok {
				cache.Store(operation, true)
				return true
			}

			if proto.HasExtension(methodDesc.Options(), auth.E_PublicAccess) {
				isPublic := proto.GetExtension(methodDesc.Options(), auth.E_PublicAccess).(bool)
				shouldAuth = !isPublic
			}
		}

		cache.Store(operation, shouldAuth)
		return shouldAuth
	}
}
