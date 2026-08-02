package s3

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/aws/smithy-go"

	"cyber-ecosystem/shared-go/capability/storage"
)

// smithyErr builds an error that mapError inspects via errors.As(&smithy.APIError).
func smithyErr(code string) error { return &smithy.GenericAPIError{Code: code, Message: "x"} }

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
	// unknown codes fall through to a wrapped default (NOT a known sentinel)
	for _, in := range []error{smithyErr("SomeUnknownCode"), errors.New("plain net error")} {
		got := mapError(in, "op")
		for _, sentinel := range []error{storage.ErrNotFound, storage.ErrForbidden, storage.ErrInvalidArgument, storage.ErrUnavailable} {
			if errors.Is(got, sentinel) {
				t.Errorf("mapError(%v): unexpectedly mapped to a known sentinel %v", in, sentinel)
			}
		}
	}
	if got := mapError(nil, "op"); got != nil {
		t.Errorf("mapError(nil): got %v, want nil", got)
	}
}

func TestSizeLimitReader(t *testing.T) {
	// over limit → ErrSizeExceeded
	r := &sizeLimitReader{r: strings.NewReader("hello world"), max: 5}
	if _, err := io.ReadAll(r); !errors.Is(err, storage.ErrSizeExceeded) {
		t.Fatalf("over-limit: err=%v want ErrSizeExceeded", err)
	}
	// under limit → full read, no error
	r2 := &sizeLimitReader{r: strings.NewReader("hi"), max: 100}
	all, err := io.ReadAll(r2)
	if err != nil || string(all) != "hi" {
		t.Fatalf("under-limit: all=%q err=%v", all, err)
	}
	// max==0 → no limit
	r3 := &sizeLimitReader{r: strings.NewReader("no limit here"), max: 0}
	all3, err := io.ReadAll(r3)
	if err != nil || string(all3) != "no limit here" {
		t.Fatalf("max=0: all=%q err=%v", all3, err)
	}
}

func TestEncodeCopySource(t *testing.T) {
	// spaces encoded, '/' preserved
	if got := encodeCopySource("b", "dir/key with space"); got != "b/dir/key%20with%20space" {
		t.Errorf("space/slash: got %s", got)
	}
	// '/' must never become %2F (it's a valid path separator in S3 keys)
	if got := encodeCopySource("b", "a/b/c"); strings.Contains(got, "%2F") {
		t.Errorf("'/' must be preserved, got %s", got)
	}
	// multibyte fully percent-encoded (no raw non-ASCII bytes survive)
	mb := encodeCopySource("b", "用户/资料")
	for _, r := range mb {
		if r > 127 {
			t.Errorf("multibyte not encoded: %s (rune %U)", mb, r)
		}
	}
}
