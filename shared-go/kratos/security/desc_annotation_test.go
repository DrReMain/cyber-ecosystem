package security_test

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	extv1 "cyber-ecosystem/gen/go/cyber/ext/v1"
)

// TestEveryBusinessRPCDeclaresAction asserts that every method in the cyber.*
// namespace declares an action in its method desc, so the audit classifier's
// authoritative level is always populated — a missing action would silently
// fall to the WRITE floor (over-collection, the invisible failure direction).
// Unlike access there are no exemptions: even the deliberately
// access-unannotated ListResource carries its desc facts.
func TestEveryBusinessRPCDeclaresAction(t *testing.T) {
	seen := 0
	protoregistry.GlobalFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		pkg := string(fd.Package())
		if !strings.HasPrefix(pkg, "cyber.") || pkg == "cyber.ext.v1" {
			return true
		}
		services := fd.Services()
		for i := 0; i < services.Len(); i++ {
			methods := services.Get(i).Methods()
			for j := 0; j < methods.Len(); j++ {
				m := methods.Get(j)
				seen++
				d, ok := proto.GetExtension(m.Options(), extv1.E_Method).(*extv1.MethodDesc)
				if !ok || d == nil || d.GetAction() == extv1.Action_ACTION_UNSPECIFIED {
					t.Errorf("%s: no cyber.ext.v1.method action declared — audit would fall to the WRITE floor", m.FullName())
				}
			}
		}
		return true
	})
	if seen == 0 {
		t.Fatal("no cyber.* methods discovered — descriptor registration is broken")
	}
}

// TestDatascopeImpliesAdminBusiness enforces the narrowing contract's shape:
// row-level narrowing rides the workforce business plane only — an annotated
// datascope on a builtin or non-ADMIN face would offer configuration that can
// never take effect (the grant editor's scope picker keys on this).
func TestDatascopeImpliesAdminBusiness(t *testing.T) {
	seen := 0
	protoregistry.GlobalFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		pkg := string(fd.Package())
		if !strings.HasPrefix(pkg, "cyber.") || pkg == "cyber.ext.v1" {
			return true
		}
		services := fd.Services()
		for i := 0; i < services.Len(); i++ {
			methods := services.Get(i).Methods()
			for j := 0; j < methods.Len(); j++ {
				m := methods.Get(j)
				if !proto.GetExtension(m.Options(), extv1.E_Datascope).(bool) {
					continue
				}
				seen++
				access := proto.GetExtension(m.Options(), extv1.E_Access).(extv1.Access)
				if access != extv1.Access_ACCESS_ADMIN {
					t.Errorf("%s: datascope must ride ACCESS_ADMIN, got access=%v", m.FullName(), access)
				}
			}
		}
		return true
	})
	if seen == 0 {
		t.Fatal("no datascope-annotated methods discovered — the narrowing contract lost all its rows")
	}
}
