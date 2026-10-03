package security_test

import (
	"os/exec"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	// Register the generated cyber.* descriptors so the walk below sees the
	// whole contract surface. A new service adds one import line here.
	_ "cyber-ecosystem/gen/go/cyber/agent/v1"
	extv1 "cyber-ecosystem/gen/go/cyber/ext/v1"
	_ "cyber-ecosystem/gen/go/cyber/shared/common/v1"
	_ "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
	_ "cyber-ecosystem/gen/go/cyber/system/v1"
)

// accessExemptions lists deliberately unannotated RPCs; every entry must stay
// deliberate (unannotated business RPCs fail at runtime under DefaultGuard).
// The unannotated-rejection path itself is covered by TestDefaultGuard.
var accessExemptions = map[protoreflect.FullName]bool{}

// TestEveryBusinessRPCDeclaresAccess asserts that every method in the cyber.*
// namespace carries an explicit access annotation, so a forgotten annotation
// fails here instead of 503-ing in production.
func TestEveryBusinessRPCDeclaresAccess(t *testing.T) {
	seen := 0
	protoregistry.GlobalFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		pkg := string(fd.Package())
		if !strings.HasPrefix(pkg, "cyber.") {
			return true
		}
		if pkg == "cyber.ext.v1" {
			return true // repo-level extensions declare no services
		}
		services := fd.Services()
		for i := 0; i < services.Len(); i++ {
			methods := services.Get(i).Methods()
			for j := 0; j < methods.Len(); j++ {
				m := methods.Get(j)
				seen++
				full := m.FullName()
				if accessExemptions[full] {
					continue
				}
				acc, ok := proto.GetExtension(m.Options(), extv1.E_Access).(extv1.Access)
				if !ok || acc == extv1.Access_ACCESS_UNSPECIFIED {
					t.Errorf("%s: no cyber.ext.v1.access annotation — deny-by-default guard would 503", full)
				}
			}
		}
		return true
	})
	if seen == 0 {
		t.Fatal("no cyber.* methods discovered — descriptor registration is broken")
	}
	// Exemption entries must keep pointing at real methods, or they rot silently.
	for name := range accessExemptions {
		d, err := protoregistry.GlobalFiles.FindDescriptorByName(name)
		if err != nil {
			t.Errorf("access exemption %q no longer matches any method — remove it", name)
			continue
		}
		if _, ok := d.(protoreflect.MethodDescriptor); !ok {
			t.Errorf("access exemption %q resolves to a non-method descriptor — remove it", name)
		}
	}
}

// TestBuiltinImpliesAdmin asserts builtin operations sit on the ADMIN
// audience: a PUBLIC method never needs builtin (it is reachable without
// a session, so the baseline grant would be meaningless there).
func TestBuiltinImpliesAdmin(t *testing.T) {
	builtins := 0
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
				if !proto.GetExtension(m.Options(), extv1.E_Builtin).(bool) {
					continue
				}
				builtins++
				acc, ok := proto.GetExtension(m.Options(), extv1.E_Access).(extv1.Access)
				if !ok || acc != extv1.Access_ACCESS_ADMIN {
					t.Errorf("%s: builtin must be ACCESS_ADMIN — PUBLIC needs no builtin", m.FullName())
				}
			}
		}
		return true
	})
	if builtins == 0 {
		t.Fatal("no builtin operations discovered — descriptor registration may be broken")
	}
}

// TestAccessImportsComplete guards the blank-import list above: it diffs the
// packages under gen/go/cyber (per `go list`) against what protoregistry
// actually has registered. A new service without its import line fails here
// loudly instead of being silently skipped by the access walk.
func TestAccessImportsComplete(t *testing.T) {
	out, err := exec.Command("go", "list", "cyber-ecosystem/gen/go/cyber/...").Output()
	if err != nil {
		t.Skipf("go list unavailable: %v", err)
	}
	want := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		// v1connect holds no proto descriptors; ext is registered via the
		// named extv1 import above (its proto path `ext/v1/...` does not map
		// to its go package `cyber/ext/v1` — the PACKAGE_DIRECTORY_MATCH gap).
		if line != "" && !strings.HasSuffix(line, "/v1connect") && !strings.Contains(line, "/ext/") {
			want[line] = true
		}
	}
	registered := map[string]bool{}
	protoregistry.GlobalFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		dir := fd.Path()
		if i := strings.LastIndexByte(dir, '/'); i >= 0 {
			registered["cyber-ecosystem/gen/go/"+dir[:i]] = true
		}
		return true
	})
	for p := range want {
		if !registered[p] {
			t.Errorf("%s has generated descriptors but is not blank-imported — add it to this file's imports", p)
		}
	}
}
