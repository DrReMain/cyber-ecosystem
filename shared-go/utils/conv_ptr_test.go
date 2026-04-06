package utils

import (
	"testing"
)

func TestPtr(t *testing.T) {
	v := 42
	p := Ptr(v)
	if p == nil || *p != v {
		t.Errorf("Ptr(%d) = %v, want %d", v, p, v)
	}
}

func TestDeref(t *testing.T) {
	v := 42
	if got := Deref(&v, 0); got != v {
		t.Errorf("Deref(&%d, 0) = %d, want %d", v, got, v)
	}
	if got := Deref[int](nil, 0); got != 0 {
		t.Errorf("Deref(nil, 0) = %d, want 0", got)
	}
}

func TestPtrApply(t *testing.T) {
	v := 10
	got := PtrApply(&v, func(x int) string { return "ok" })
	if got == nil || *got != "ok" {
		t.Errorf("PtrApply = %v, want 'ok'", got)
	}
	if got := PtrApply[int, string](nil, func(x int) string { return "ok" }); got != nil {
		t.Errorf("PtrApply(nil, f) = %v, want nil", got)
	}
}

func TestConvPtr(t *testing.T) {
	v := uint8(5)
	got := ConvPtr[uint32, uint8](&v)
	if got == nil || *got != uint32(5) {
		t.Errorf("ConvPtr[uint32,uint8](&5) = %v, want 5", got)
	}
	var nilPtr *uint8
	if got := ConvPtr[uint32, uint8](nilPtr); got != nil {
		t.Errorf("ConvPtr[uint32,uint8](nil) = %v, want nil", got)
	}
}
