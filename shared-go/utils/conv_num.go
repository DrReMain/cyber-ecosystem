package utils

// --- Numeric constraints ---

type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

type Float interface {
	~float32 | ~float64
}

type Number interface {
	Integer | Unsigned | Float
}

// ConvNum converts a numeric value to target type R.
func ConvNum[R Number, T Number](src T) R {
	return R(src)
}
