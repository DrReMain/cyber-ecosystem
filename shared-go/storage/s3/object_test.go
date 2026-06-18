package s3

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"cyber-ecosystem/shared-go/storage"
)

func TestObjectUploadDownloadDelete(t *testing.T) {
	s, cleanup := newTestStorage(t)
	defer cleanup()
	ctx := context.Background()
	resetObjects(t, s, "obj/roundtrip")

	body := []byte("hello storage")
	if _, err := s.Object.Upload(ctx, "obj/roundtrip", bytes.NewReader(body), int64(len(body)), "text/plain"); err != nil {
		t.Fatalf("Upload: %v", err)
	}
	rc, info, err := s.Object.Download(ctx, "obj/roundtrip")
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	got, _ := io.ReadAll(rc)
	_ = rc.Close()
	if !bytes.Equal(got, body) {
		t.Fatalf("round-trip: got %q want %q", got, body)
	}
	if info.ContentType != "text/plain" {
		t.Errorf("ContentType: got %q want text/plain", info.ContentType)
	}
	if err := s.Object.Delete(ctx, "obj/roundtrip"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	// After delete, download → ErrNotFound.
	if _, _, err := s.Object.Download(ctx, "obj/roundtrip"); !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("post-delete Download: got %v, want ErrNotFound", err)
	}
}

func TestObjectSizeLimit(t *testing.T) {
	s, cleanup := newTestStorage(t)
	defer cleanup()
	ctx := context.Background()
	resetObjects(t, s, "obj/toobig")

	// MaxFileSize=5MiB; declare 6MiB → pre-check rejects without reading.
	if _, err := s.Object.Upload(ctx, "obj/toobig", bytes.NewReader(make([]byte, 6<<20)), 6<<20, "application/octet-stream"); !errors.Is(err, storage.ErrSizeExceeded) {
		t.Errorf("oversize known: got %v, want ErrSizeExceeded", err)
	}
	// size<0 (unknown) but body exceeds → streaming counter rejects mid-read.
	if _, err := s.Object.Upload(ctx, "obj/toobig", bytes.NewReader(make([]byte, 6<<20)), -1, "application/octet-stream"); !errors.Is(err, storage.ErrSizeExceeded) {
		t.Errorf("oversize streamed: got %v, want ErrSizeExceeded", err)
	}
}

func TestObjectStatCopy(t *testing.T) {
	s, cleanup := newTestStorage(t)
	defer cleanup()
	ctx := context.Background()
	resetObjects(t, s, "stat/src", "stat/dst", "stat/dst2")

	if _, err := s.Object.Upload(ctx, "stat/src", bytes.NewReader([]byte("stat-me")), 7, "text/plain"); err != nil {
		t.Fatal(err)
	}
	info, err := s.Object.Stat(ctx, "stat/src")
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Size != 7 {
		t.Errorf("Stat size: got %d want 7", info.Size)
	}
	if _, err := s.Object.Stat(ctx, "stat/missing"); !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("Stat missing: got %v want ErrNotFound", err)
	}
	if _, err := s.Object.Copy(ctx, "stat/src", "stat/dst"); err != nil {
		t.Fatalf("Copy: %v", err)
	}
	rc, _, err := s.Object.Download(ctx, "stat/dst")
	if err != nil {
		t.Fatalf("Download dst: %v", err)
	}
	got, _ := io.ReadAll(rc)
	_ = rc.Close()
	if string(got) != "stat-me" {
		t.Errorf("Copy content: got %q", got)
	}
	if _, err := s.Object.Copy(ctx, "stat/missing", "stat/dst2"); !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("Copy missing src: got %v want ErrNotFound", err)
	}
}
