package utils

import "strings"

// MaskPhone is the one phone-masking home: the 3-****-4 channel form
// (153****0101) every face that shows a member's number to an operator
// renders. Shorter than seven runes is fully masked — a fragment that small
// identifies nobody and leaking its edges identifies the length.
func MaskPhone(phone string) string {
	if len(phone) < 7 {
		return strings.Repeat("*", len(phone))
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}
