package utils

// Ptr returns a pointer to v.
func Ptr[T any](v T) *T { return &v }

// Deref safely dereferences p, returning def if nil.
func Deref[T any](p *T, def T) T {
	if p == nil {
		return def
	}
	return *p
}

// PtrApply applies f to the value pointed to by src. Returns nil if src is nil.
func PtrApply[T any, R any](src *T, f func(T) R) *R {
	if src == nil {
		return nil
	}
	v := f(*src)
	return &v
}

// ConvPtr converts *T to *R with nil safety.
func ConvPtr[R Number, T Number](src *T) *R {
	if src == nil {
		return nil
	}
	v := R(*src)
	return &v
}
