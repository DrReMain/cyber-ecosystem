package storage

const maxKeyLen = 1024

// ValidateKey rejects empty or over-long object keys. Key prefixing is the
// caller's concern; storage only guards the structural invariants every backend
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
