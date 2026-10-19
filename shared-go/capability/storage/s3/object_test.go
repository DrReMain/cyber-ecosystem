package s3

import (
	"errors"
	"io"
	"strings"
	"testing"

	"cyber-ecosystem/shared-go/capability/storage"
)

func TestSizeLimitReader(t *testing.T) {
	r := &sizeLimitReader{r: strings.NewReader("hello world"), max: 5}
	if _, err := io.ReadAll(r); !errors.Is(err, storage.ErrSizeExceeded) {
		t.Fatalf("over limit: err=%v, want ErrSizeExceeded", err)
	}

	r2 := &sizeLimitReader{r: strings.NewReader("hi"), max: 100}
	all, err := io.ReadAll(r2)
	if err != nil || string(all) != "hi" {
		t.Fatalf("under limit: all=%q err=%v", all, err)
	}

	r3 := &sizeLimitReader{r: strings.NewReader("no limit here"), max: 0}
	all3, err := io.ReadAll(r3)
	if err != nil || string(all3) != "no limit here" {
		t.Fatalf("max=0: all=%q err=%v", all3, err)
	}
}

func TestEncodeCopySource(t *testing.T) {
	if got := encodeCopySource("b", "dir/key with space"); got != "b/dir/key%20with%20space" {
		t.Errorf("space/slash: got %s", got)
	}
	if got := encodeCopySource("b", "a/b/c"); strings.Contains(got, "%2F") {
		t.Errorf("'/' must be preserved, got %s", got)
	}
	// multibyte fully percent-encoded
	mb := encodeCopySource("b", "用户/资料")
	for _, r := range mb {
		if r > 127 {
			t.Errorf("multibyte not encoded: %s (rune %U)", mb, r)
		}
	}
}
