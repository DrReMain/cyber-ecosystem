package shared

import (
	"unicode/utf8"

	systempb "cyber-ecosystem/gen/go/cyber/system/v1"
)

func PasswordPolicyError(candidate string) error {
	hasLetter, hasDigit := false, false
	for _, r := range candidate {
		switch {
		case ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z'):
			hasLetter = true
		case '0' <= r && r <= '9':
			hasDigit = true
		}
	}
	// The prototype's own rule enforced server-side: at least six runes
	// carrying both an ASCII letter and an ASCII digit — one law for every
	// seam that accepts a freshly chosen password. The hash-only callers
	// (login rehash, auto-register's closed plane, the boot seed) stay
	// outside on purpose: none of them carries a member's fresh choice.
	if utf8.RuneCountInString(candidate) < 6 || !hasLetter || !hasDigit {
		return systempb.ErrorSystemPasswordPolicy("")
	}
	return nil
}
