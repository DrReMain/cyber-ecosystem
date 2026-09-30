package authz

import (
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	extv1 "cyber-ecosystem/gen/go/cyber/ext/v1"
	// The package-level collector runs at init; registering this service's
	// protos here makes the ordering guarantee explicit (imported packages
	// fully initialize first) instead of depending on sibling import order.
	_ "cyber-ecosystem/gen/go/cyber/system/v1"
)

var builtinOperations = collectBuiltinOperations()

func collectBuiltinOperations() []string {
	const protoPrefix = "cyber/system/v1/"
	var ops []string
	protoregistry.GlobalFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		if !strings.HasPrefix(fd.Path(), protoPrefix) {
			return true
		}
		services := fd.Services()
		for i := 0; i < services.Len(); i++ {
			svc := services.Get(i)
			methods := svc.Methods()
			for j := 0; j < methods.Len(); j++ {
				m := methods.Get(j)
				if proto.GetExtension(m.Options(), extv1.E_Builtin).(bool) {
					ops = append(ops, "/"+string(svc.FullName())+"/"+string(m.Name()))
				}
			}
		}
		return true
	})
	return ops
}
