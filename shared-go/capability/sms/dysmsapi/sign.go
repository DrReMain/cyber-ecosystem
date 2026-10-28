package dysmsapi

import (
	"crypto/hmac"
	"crypto/sha1" // #nosec G505 -- the Aliyun RPC signature v1 protocol mandates HMAC-SHA1: a wire contract, not a security choice
	"encoding/base64"
	"net/url"
	"sort"
	"strings"
)

// percentEncode is the RPC protocol's own escaping: RFC 3986 unreserved
// characters stay, everything else percent-encodes uppercase, and the two
// URL-safe characters url.Values.Encode treats specially revert to the
// protocol's expectation (+ becomes %20, * becomes %2A, ~ stays).
func percentEncode(s string) string {
	var b strings.Builder
	for _, c := range []byte(url.QueryEscape(s)) {
		switch c {
		case '+':
			b.WriteString("%20")
		case '*':
			b.WriteString("%2A")
		case '%':
			// QueryEscape already emits uppercase hex; pass through.
			b.WriteByte(c)
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}

// canonicalQuery renders the sorted, percent-encoded parameter string the
// signature covers — byte-exact order matters, the protocol signs the
// serialized form.
func canonicalQuery(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte('&')
		}
		b.WriteString(percentEncode(k))
		b.WriteByte('=')
		b.WriteString(percentEncode(params[k]))
	}
	return b.String()
}

// stringToSign builds the signed payload: the method, the encoded slash,
// and the encoded query, each element itself encoded once more.
func stringToSign(canonical string) string {
	return "POST&" + percentEncode("/") + "&" + percentEncode(canonical)
}

// signStringToSign returns the base64 HMAC-SHA1 of the string-to-sign
// under the secret plus the protocol's trailing ampersand.
func signStringToSign(secret, sts string) string {
	mac := hmac.New(sha1.New, []byte(secret+"&"))
	mac.Write([]byte(sts))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
