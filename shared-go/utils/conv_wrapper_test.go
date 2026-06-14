package utils

import (
	"testing"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestWrap(t *testing.T) {
	if got := Wrap(nil, wrapperspb.String); got != nil {
		t.Errorf("Wrap(nil, String) = %v, want nil", got)
	}
	s := "hello"
	got := Wrap(&s, wrapperspb.String)
	if got == nil || got.GetValue() != "hello" {
		t.Errorf("Wrap(&hello, String) = %v, want hello", got)
	}
}

func TestUnwrap(t *testing.T) {
	if got := Unwrap[string](nil); got != nil {
		t.Errorf("Unwrap(nil) = %v, want nil", got)
	}
	w := wrapperspb.String("hello")
	got := Unwrap[string](w)
	if got == nil || *got != "hello" {
		t.Errorf("Unwrap(String(hello)) = %v, want hello", got)
	}
}

func TestPredefinedConstructors(t *testing.T) {
	v := uint8(42)
	got := UInt32FromUint8(v)
	if got.GetValue() != uint32(42) {
		t.Errorf("UInt32FromUint8(42) = %d, want 42", got.GetValue())
	}
}
