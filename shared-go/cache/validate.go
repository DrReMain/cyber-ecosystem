package cache

import "strings"

const maxKeyLen = 1024

// ValidateKey rejects empty or over-long keys. Key prefixing is the caller's
// concern; the cache layer only guards the structural invariants every backend
// relies on (non-empty, bounded length).
func ValidateKey(key string) error {
	if key == "" {
		return ErrInvalidArgument
	}
	if len(key) > maxKeyLen {
		return ErrInvalidArgument
	}
	return nil
}

func ValidateKeys(keys ...string) error {
	for _, k := range keys {
		if err := ValidateKey(k); err != nil {
			return err
		}
	}
	return nil
}

func ValidatePairs(pairs map[string][]byte) error {
	for k := range pairs {
		if err := ValidateKey(k); err != nil {
			return err
		}
	}
	return nil
}

// sessionForbidden holds characters forbidden in session ids/keys because they
// break the "session:{id}:{key}" namespacing: ':' collides across sessions'
// keyspaces, and the glob metacharacters let a malicious id match unrelated
// keys in SCAN/KEYS (e.g. Destroy("*") deleting every session). Enforced at the
// cache.Session boundary so ALL backends inherit the isolation guarantee.
const sessionForbidden = ":*?[\\"

// ValidateSessionID rejects ids that would break session namespacing or enable
// pattern injection in any backend.
func ValidateSessionID(id string) error {
	if id == "" || strings.ContainsAny(id, sessionForbidden) {
		return ErrInvalidArgument
	}
	return nil
}

// ValidateSessionKey rejects keys containing ':' (would collide across the
// session's own fields or other sessions).
func ValidateSessionKey(key string) error {
	if key == "" || strings.Contains(key, ":") {
		return ErrInvalidArgument
	}
	return nil
}
