package scopes

import (
	"strings"
	"testing"

	"entgo.io/ent/dialect/sql"

	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"
)

// render applies a result to a scratch selector and returns the query SQL,
// so predicate assertions read as what the database would see.
func render(r Result) (string, []any) {
	s := sql.Select("*").From(sql.Table("rows"))
	if !r.Unscoped {
		r.Predicate(s)
	}
	q, args := s.Query()
	return q, args
}

func TestCompile(t *testing.T) {
	user := kauthz.ScopeDescriptor{Kind: kauthz.ScopeKindSelf}
	dept := kauthz.ScopeDescriptor{
		Kind:   kauthz.ScopeKindDeptTree,
		Params: map[string]any{kauthz.ScopeParamDeptIDs: []string{"d1", "d2"}},
	}
	deptEmpty := kauthz.ScopeDescriptor{Kind: kauthz.ScopeKindDeptTree}
	unknown := kauthz.ScopeDescriptor{Kind: "region"}
	userFields := Fields{User: "user_id", Dept: "dept_id"}
	deptOnly := Fields{Dept: "id"} // e.g. the dept entity: self not expressible

	tests := []struct {
		name        string
		descriptors []kauthz.ScopeDescriptor
		fields      Fields
		wantState   string // filtered | unscoped | closed
		wantSQL     string
		wantArgs    []any
	}{
		{"empty set is the builtin baseline", nil, userFields, "unscoped", "", nil},
		{"tenant-wide kind dominates", []kauthz.ScopeDescriptor{user, {Kind: ""}}, userFields, "unscoped", "", nil},
		{"all kind dominates", []kauthz.ScopeDescriptor{{Kind: kauthz.ScopeKindAll}}, userFields, "unscoped", "", nil},
		{"self compiles to owner equality", []kauthz.ScopeDescriptor{user}, userFields, "filtered", "`user_id` = ?", []any{"u1"}},
		{"dept_tree compiles to id set membership", []kauthz.ScopeDescriptor{dept}, userFields, "filtered", "`dept_id` IN (?, ?)", []any{"d1", "d2"}},
		{"dept_tree with no subtree yields no rows", []kauthz.ScopeDescriptor{deptEmpty}, userFields, "filtered", "FALSE", nil},
		{"self on an entity without the dimension fails closed", []kauthz.ScopeDescriptor{user}, deptOnly, "closed", "false", nil},
		{"unknown kind alone fails closed", []kauthz.ScopeDescriptor{unknown}, userFields, "closed", "false", nil},
		{"unknown kind beside an honorable one over-narrows to it", []kauthz.ScopeDescriptor{unknown, user}, userFields, "filtered", "`user_id` = ?", []any{"u1"}},
		{"inexpressible kind beside an expressible one over-narrows to it", []kauthz.ScopeDescriptor{user, dept}, deptOnly, "filtered", "`id` IN (?, ?)", []any{"d1", "d2"}},
		{"malformed dept params fail like an empty subtree", []kauthz.ScopeDescriptor{{Kind: kauthz.ScopeKindDeptTree, Params: map[string]any{kauthz.ScopeParamDeptIDs: "d1"}}}, userFields, "filtered", "FALSE", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Compile(tt.descriptors, tt.fields, "u1")
			switch tt.wantState {
			case "unscoped":
				if !r.Unscoped {
					t.Fatalf("result = %+v, want unscoped", r)
				}
				return
			case "closed":
				if !r.Closed {
					t.Fatalf("result = %+v, want closed", r)
				}
			case "filtered":
				if r.Unscoped || r.Closed || r.Predicate == nil {
					t.Fatalf("result = %+v, want filtered", r)
				}
			}
			q, args := render(r)
			if !strings.Contains(q, tt.wantSQL) {
				t.Fatalf("sql = %q, want containing %q", q, tt.wantSQL)
			}
			if len(args) != len(tt.wantArgs) {
				t.Fatalf("args = %v, want %v", args, tt.wantArgs)
			}
			for i := range args {
				if args[i] != tt.wantArgs[i] {
					t.Fatalf("args = %v, want %v", args, tt.wantArgs)
				}
			}
		})
	}
}
