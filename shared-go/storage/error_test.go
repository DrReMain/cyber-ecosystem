package storage

import (
	"errors"
	"testing"

	kratoserrors "github.com/go-kratos/kratos/v3/errors"
)

func TestHandleStorageError(t *testing.T) {
	errs := &StorageDefaultError{
		NotFound:        kratoserrors.New(404, "NOT_FOUND", "x"),
		Forbidden:       kratoserrors.New(403, "FORBIDDEN", "x"),
		SizeExceeded:    kratoserrors.New(413, "SIZE_EXCEED", "x"),
		InvalidArgument: kratoserrors.New(400, "INVALID_ARGUMENT", "x"),
		Unavailable:     kratoserrors.New(503, "UNAVAILABLE", "x"),
	}
	cases := []struct {
		in   error
		want int32 // http code (kratoserrors.Error.Code is int32)
	}{
		{ErrNotFound, 404},
		{ErrForbidden, 403},
		{ErrSizeExceeded, 413},
		{ErrInvalidArgument, 400},
		{errors.New("boom"), 503}, // unknown → Unavailable
	}
	for _, c := range cases {
		got := HandleStorageError(c.in, errs)
		var ke *kratoserrors.Error
		if !errors.As(got, &ke) || ke.Code != c.want {
			t.Errorf("HandleStorageError(%v): code=%d, want %d", c.in, ke.Code, c.want)
		}
	}
}

func TestValidateStorageDefaultErrorRejectsNilSlot(t *testing.T) {
	if err := ValidateStorageDefaultError(&StorageDefaultError{}); err == nil {
		t.Fatal("want error for all-nil slots")
	}
	// Unavailable optional: nil Unavailable must still pass.
	good := &StorageDefaultError{
		NotFound:        kratoserrors.New(404, "n", ""),
		Forbidden:       kratoserrors.New(403, "f", ""),
		SizeExceeded:    kratoserrors.New(413, "s", ""),
		InvalidArgument: kratoserrors.New(400, "i", ""),
	}
	if err := ValidateStorageDefaultError(good); err != nil {
		t.Fatalf("nil Unavailable should be allowed: %v", err)
	}
}

func TestValidateKey(t *testing.T) {
	if err := ValidateKey(""); err != ErrInvalidArgument {
		t.Errorf("empty key: got %v, want ErrInvalidArgument", err)
	}
	if err := ValidateKey("ok"); err != nil {
		t.Errorf("valid key: got %v", err)
	}
	if err := ValidateKey(string(make([]byte, maxKeyLen+1))); err != ErrInvalidArgument {
		t.Errorf("overlong key: got %v, want ErrInvalidArgument", err)
	}
}
