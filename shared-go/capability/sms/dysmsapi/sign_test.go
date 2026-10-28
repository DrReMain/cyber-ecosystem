package dysmsapi

import (
	"testing"
)

// TestPercentEncode pins the RPC protocol's escaping to RFC 3986 truth:
// space becomes %20 (never +), * becomes %2A, ~ stays, everything else
// outside the unreserved set encodes uppercase-hex.
func TestPercentEncode(t *testing.T) {
	cases := []struct{ in, want string }{
		{"abcXYZ019", "abcXYZ019"},
		{"a b", "a%20b"},
		{"a+b", "a%2Bb"},
		{"a*b", "a%2Ab"},
		{"a~b", "a~b"},
		{"a=b", "a%3Db"},
		{"a&b", "a%26b"},
		{"a/b", "a%2Fb"},
		{"中文", "%E4%B8%AD%E6%96%87"},
		{"{\"code\":\"1\"}", "%7B%22code%22%3A%221%22%7D"},
	}
	for _, tc := range cases {
		if got := percentEncode(tc.in); got != tc.want {
			t.Errorf("percentEncode(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestCanonicalQuerySorts: the signature covers the sorted serialized form,
// so parameter insertion order must not survive into the canonical string.
func TestCanonicalQuerySorts(t *testing.T) {
	a := map[string]string{"zeta": "1", "alpha": "2", "Mid": "3"}
	b := map[string]string{"alpha": "2", "Mid": "3", "zeta": "1"}
	if got, want := canonicalQuery(a), canonicalQuery(b); got != want {
		t.Fatalf("insertion order leaked into canonical query: %q vs %q", got, want)
	}
	if want := "Mid=3&alpha=2&zeta=1"; canonicalQuery(a) != want {
		t.Fatalf("canonicalQuery = %q, want %q", canonicalQuery(a), want)
	}
}

// TestStringToSignShape: the signed payload is the method, the encoded
// slash, and the encoded query — each element itself encoded once more.
func TestStringToSignShape(t *testing.T) {
	if got, want := stringToSign("a=b"), "POST&%2F&a%3Db"; got != want {
		t.Fatalf("stringToSign = %q, want %q", got, want)
	}
}

// TestSignStringToSignGoldenVectors pins the HMAC-SHA1-base64 primitive to
// vectors computed by openssl outside this codebase — the signature never
// regresses into self-consistent wrongness.
func TestSignStringToSignGoldenVectors(t *testing.T) {
	cases := []struct{ secret, sts, want string }{
		// printf 'POST&%2F&hello' | openssl dgst -sha1 -hmac 'testsecret&' -binary | base64
		{"testsecret", "POST&%2F&hello", "ucQWKQUFeOLqW6+LoR5wBQ9jxF0="},
		// printf 'POST&%2F&c2Vjb25k' | openssl dgst -sha1 -hmac 'anotherSecret&' -binary | base64
		{"anotherSecret", "POST&%2F&c2Vjb25k", "YbZaQHmfzahXcPdqCW8TQKdDklo="},
	}
	for _, tc := range cases {
		if got := signStringToSign(tc.secret, tc.sts); got != tc.want {
			t.Errorf("signStringToSign(%q) = %q, want %q", tc.sts, got, tc.want)
		}
	}
}
