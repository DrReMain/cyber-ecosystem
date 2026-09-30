package utils

// SliceMap applies f to each element of s, returning a new slice.
// Returns nil if s is empty.
func SliceMap[T any, R any](s []T, f func(T) R) []R {
	if len(s) == 0 {
		return nil
	}
	result := make([]R, 0, len(s))
	for _, v := range s {
		result = append(result, f(v))
	}
	return result
}

// SliceMapErr applies f to each element, stopping on first error.
func SliceMapErr[T any, R any](s []T, f func(T) (R, error)) ([]R, error) {
	if len(s) == 0 {
		return nil, nil
	}
	result := make([]R, 0, len(s))
	for _, v := range s {
		r, err := f(v)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}
