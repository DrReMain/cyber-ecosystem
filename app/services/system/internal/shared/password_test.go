package shared

import (
	"testing"

	systempb "cyber-ecosystem/gen/go/cyber/system/v1"
)

func TestPasswordPolicyError(t *testing.T) {
	// The prototype's own rule, one law for every fresh-choice seam: at
	// least six runes carrying both an ASCII letter and an ASCII digit —
	// pinned so the law can never quietly narrow.
	cases := []struct {
		name      string
		candidate string
		wantErr   bool
	}{
		{"letters and digits pass", "abc123", false},
		{"mixed case with digits passes", "Ab1c2d", false},
		{"longer than six passes", "correct horse 1", false},
		{"digits-only is refused", "123456", true},
		{"letters-only is refused", "abcdef", true},
		{"five runes is refused even when mixed", "ab1c2", true},
		{"empty is refused", "", true},
		{"symbols alone carry neither class", "!!!!!!", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := PasswordPolicyError(tc.candidate)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("PasswordPolicyError(%q) = nil, want the policy error", tc.candidate)
				}
				if !systempb.IsSystemPasswordPolicy(err) {
					t.Fatalf("error = %v, want SystemPasswordPolicy", err)
				}
			} else if err != nil {
				t.Fatalf("PasswordPolicyError(%q) = %v, want nil", tc.candidate, err)
			}
		})
	}
}
