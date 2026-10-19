package s3

import (
	"errors"
	"testing"

	"github.com/aws/smithy-go"

	"cyber-ecosystem/shared-go/capability/storage"
)

func smithyErr(code string) error {
	return &smithy.GenericAPIError{Code: code, Message: "x"}
}

func TestMapError(t *testing.T) {
	known := []struct {
		in   error
		want error
	}{
		{smithyErr("NoSuchKey"), storage.ErrNotFound},
		{smithyErr("NotFound"), storage.ErrNotFound},
		{smithyErr("NoSuchUpload"), storage.ErrNotFound},
		{smithyErr("NoSuchBucket"), storage.ErrNotFound},
		{smithyErr("AccessDenied"), storage.ErrForbidden},
		{smithyErr("Forbidden"), storage.ErrForbidden},
		{smithyErr("InvalidPart"), storage.ErrInvalidArgument},
		{smithyErr("InvalidPartOrder"), storage.ErrInvalidArgument},
		{smithyErr("EntityTooSmall"), storage.ErrInvalidArgument},
		{smithyErr("BucketNotEmpty"), storage.ErrInvalidArgument},
		{smithyErr("SlowDown"), storage.ErrUnavailable},
		{smithyErr("ServiceUnavailable"), storage.ErrUnavailable},
		{smithyErr("InternalError"), storage.ErrUnavailable},
		{smithyErr("RequestTimeout"), storage.ErrUnavailable},
	}
	for _, c := range known {
		if got := mapError(c.in, "op"); !errors.Is(got, c.want) {
			t.Errorf("mapError(%v): got %v, want %v", c.in, got, c.want)
		}
	}
	// unknown codes fall through wrapped, never a known sentinel
	for _, in := range []error{smithyErr("SomeUnknownCode"), errors.New("plain net error")} {
		got := mapError(in, "op")
		for _, sentinel := range []error{storage.ErrNotFound, storage.ErrForbidden, storage.ErrInvalidArgument, storage.ErrUnavailable} {
			if errors.Is(got, sentinel) {
				t.Errorf("mapError(%v): unexpectedly mapped to %v", in, sentinel)
			}
		}
	}
	if got := mapError(nil, "op"); got != nil {
		t.Errorf("mapError(nil) = %v, want nil", got)
	}
}
