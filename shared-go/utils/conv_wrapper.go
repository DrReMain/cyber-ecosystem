package utils

import (
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// Wrap converts *T to proto wrapper type. Returns nil if v is nil.
func Wrap[T any, W any](v *T, constructor func(T) *W) *W {
	if v == nil {
		return nil
	}
	return constructor(*v)
}

// Unwrap extracts *T from proto wrapper. Returns nil if w is nil.
func Unwrap[T any](w interface {
	GetValue() T
}) *T {
	if w == nil {
		return nil
	}
	v := w.GetValue()
	return &v
}

// --- Predefined wrapper constructors: direct type mapping ---

var (
	StringW = wrapperspb.String // *string  → *StringValue
	UInt32W = wrapperspb.UInt32 // *uint32  → *UInt32Value
	Int32W  = wrapperspb.Int32  // *int32   → *Int32Value
	BoolW   = wrapperspb.Bool   // *bool    → *BoolValue
	UInt64W = wrapperspb.UInt64 // *uint64  → *UInt64Value
	Int64W  = wrapperspb.Int64  // *int64   → *Int64Value
	FloatW  = wrapperspb.Float  // *float32 → *FloatValue
	DoubleW = wrapperspb.Double // *float64 → *DoubleValue
	BytesW  = wrapperspb.Bytes  // *[]byte  → *BytesValue
)

// --- Predefined wrapper constructors: cross-type mapping ---

var (
	UInt32FromUint8  = func(v uint8) *wrapperspb.UInt32Value { return wrapperspb.UInt32(uint32(v)) }
	UInt32FromUint16 = func(v uint16) *wrapperspb.UInt32Value { return wrapperspb.UInt32(uint32(v)) }
	UInt64FromUint8  = func(v uint8) *wrapperspb.UInt64Value { return wrapperspb.UInt64(uint64(v)) }
	UInt64FromUint16 = func(v uint16) *wrapperspb.UInt64Value { return wrapperspb.UInt64(uint64(v)) }
	UInt64FromUint32 = func(v uint32) *wrapperspb.UInt64Value { return wrapperspb.UInt64(uint64(v)) }
	Int32FromInt8    = func(v int8) *wrapperspb.Int32Value { return wrapperspb.Int32(int32(v)) }
	Int32FromInt16   = func(v int16) *wrapperspb.Int32Value { return wrapperspb.Int32(int32(v)) }
	Int64FromInt8    = func(v int8) *wrapperspb.Int64Value { return wrapperspb.Int64(int64(v)) }
	Int64FromInt16   = func(v int16) *wrapperspb.Int64Value { return wrapperspb.Int64(int64(v)) }
	Int64FromInt32   = func(v int32) *wrapperspb.Int64Value { return wrapperspb.Int64(int64(v)) }
)
