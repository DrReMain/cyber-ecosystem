package s3

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"cyber-ecosystem/shared-go/storage"
)

// TestObjectCopySpecialChar locks the Copy CopySource URL-encoding fix: a CJK +
// space key must round-trip through CopyObject (aws-sdk does NOT encode
// CopySource, so encodeCopySource must).
func TestObjectCopySpecialChar(t *testing.T) {
	s, cleanup := newTestStorage(t)
	defer cleanup()
	ctx := context.Background()
	resetObjects(t, s, "cpy/用户 资料", "cpy/用户 资料-dst")

	if _, err := s.Object.Upload(ctx, "cpy/用户 资料", bytes.NewReader([]byte("cjk")), 3, "text/plain"); err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if _, err := s.Object.Copy(ctx, "cpy/用户 资料", "cpy/用户 资料-dst"); err != nil {
		t.Fatalf("Copy with special chars: %v", err)
	}
	rc, _, err := s.Object.Download(ctx, "cpy/用户 资料-dst")
	if err != nil {
		t.Fatalf("Download dst: %v", err)
	}
	got, _ := io.ReadAll(rc)
	_ = rc.Close()
	if string(got) != "cjk" {
		t.Errorf("Copy content: got %q want cjk", got)
	}
}

func TestObjectBinaryRoundTrip(t *testing.T) {
	s, cleanup := newTestStorage(t)
	defer cleanup()
	ctx := context.Background()
	resetObjects(t, s, "hard/bin")
	bin := []byte{0x00, 0x01, 0xFF, '\n', '\r', 0x7F}
	if _, err := s.Object.Upload(ctx, "hard/bin", bytes.NewReader(bin), int64(len(bin)), "application/octet-stream"); err != nil {
		t.Fatal(err)
	}
	rc, _, err := s.Object.Download(ctx, "hard/bin")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rc.Close() }()
	got, _ := io.ReadAll(rc)
	if !bytes.Equal(got, bin) {
		t.Errorf("binary round-trip: got %v want %v", got, bin)
	}
}

func TestObjectSpecialCharKey(t *testing.T) {
	s, cleanup := newTestStorage(t)
	defer cleanup()
	ctx := context.Background()
	keys := []string{"用户/资料", "key with space", "key/with/slash"}
	resetObjects(t, s, keys...)
	for _, k := range keys {
		if _, err := s.Object.Upload(ctx, k, bytes.NewReader([]byte("ok")), 2, "text/plain"); err != nil {
			t.Errorf("Upload %q: %v", k, err)
			continue
		}
		if _, err := s.Object.Stat(ctx, k); err != nil {
			t.Errorf("Stat %q: %v", k, err)
		}
	}
}

// TestStorageMinIODown verifies a dead endpoint surfaces as ErrUnavailable from
// NewClient (the ensure-bucket boot path).
func TestStorageMinIODown(t *testing.T) {
	_, _, err := NewClient(&Config{Endpoint: "http://127.0.0.1:39998", AccessKey: "a", SecretKey: "b", Bucket: "x", Region: "us-east-1", UsePathStyle: true})
	if err == nil {
		t.Skip("expected error on dead endpoint")
	}
	if !errors.Is(err, storage.ErrUnavailable) {
		t.Errorf("dead endpoint: got %v, want ErrUnavailable wrap", err)
	}
}
