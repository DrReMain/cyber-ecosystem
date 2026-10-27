package audit

import "testing"

// The CSV export's formula-injection guard: attacker-controlled cells
// (user agent, operation, path, deny reason, actor, ip) start life as
// request data, and a leading =, +, -, @, or tab makes Excel/Sheets execute
// the cell as a formula when an admin opens the export. The sanitizer
// prefixes such cells with an apostrophe; every other value passes through
// byte-identical, and numeric columns never route through it (a negative
// number is data, not a formula).
func TestSanitizeAuditCell(t *testing.T) {
	cases := []struct{ in, want string }{
		{"=HYPERLINK(\"http://evil\",\"click\")", "'=HYPERLINK(\"http://evil\",\"click\")"},
		{"+SUM(A1:A9)", "'+SUM(A1:A9)"},
		{"-2+3|cmd", "'-2+3|cmd"},
		{"@SUM(A1)", "'@SUM(A1)"},
		{"\tTAB(1)", "'\tTAB(1)"},
		{"Mozilla/5.0 (Macintosh)", "Mozilla/5.0 (Macintosh)"},
		{"/cyber.system.v1.UserService/CreateUser", "/cyber.system.v1.UserService/CreateUser"},
		{"a=b stays literal", "a=b stays literal"}, // only the first byte can arm a formula
		{"", ""},
	}
	for _, c := range cases {
		if got := sanitizeAuditCell(c.in); got != c.want {
			t.Fatalf("sanitizeAuditCell(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
