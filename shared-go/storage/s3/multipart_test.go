package s3

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"cyber-ecosystem/shared-go/storage"
)

func TestMultipartFullFlowAndResume(t *testing.T) {
	s, cleanup := newTestStorage(t)
	defer cleanup()
	ctx := context.Background()
	resetObjects(t, s, "mp/full")

	uploadID, err := s.Multipart.Create(ctx, "mp/full", "text/plain")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	// S3 requires each non-last part ≥ 5MiB; part1 is 5MiB, part2 (last) is small.
	etag1, err := s.Multipart.UploadPart(ctx, "mp/full", uploadID, 1, bytes.NewReader(bytes.Repeat([]byte("a"), 5<<20)), 5<<20)
	if err != nil {
		t.Fatalf("UploadPart 1: %v", err)
	}
	etag2, err := s.Multipart.UploadPart(ctx, "mp/full", uploadID, 2, bytes.NewReader([]byte("part2")), 5)
	if err != nil {
		t.Fatalf("UploadPart 2: %v", err)
	}

	// ListParts — the resume primitive: a reconnecting client learns which parts exist.
	parts, err := s.Multipart.ListParts(ctx, "mp/full", uploadID)
	if err != nil {
		t.Fatalf("ListParts: %v", err)
	}
	if len(parts) != 2 {
		t.Fatalf("ListParts: got %d parts want 2", len(parts))
	}

	info, err := s.Multipart.Complete(ctx, "mp/full", uploadID, []storage.CompletedPart{
		{PartNumber: 1, ETag: etag1}, {PartNumber: 2, ETag: etag2},
	})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if info.ETag == "" {
		t.Errorf("completed: empty ETag")
	}
	// CompleteMultipartUpload output doesn't report size; verify the assembled
	// object via Stat.
	stat, err := s.Object.Stat(ctx, "mp/full")
	if err != nil {
		t.Fatalf("Stat after complete: %v", err)
	}
	if stat.Size != int64(5<<20)+5 {
		t.Errorf("assembled size: got %d want %d", stat.Size, int64(5<<20)+5)
	}
}

func TestMultipartAbort(t *testing.T) {
	s, cleanup := newTestStorage(t)
	defer cleanup()
	ctx := context.Background()
	resetObjects(t, s, "mp/abort")

	uploadID, err := s.Multipart.Create(ctx, "mp/abort", "text/plain")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := s.Multipart.UploadPart(ctx, "mp/abort", uploadID, 1, bytes.NewReader([]byte("x")), 1); err != nil {
		t.Fatalf("UploadPart: %v", err)
	}
	if err := s.Multipart.Abort(ctx, "mp/abort", uploadID); err != nil {
		t.Fatalf("Abort: %v", err)
	}
	// After abort, ListParts on the dead upload → ErrNotFound.
	if _, err := s.Multipart.ListParts(ctx, "mp/abort", uploadID); !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("post-abort ListParts: got %v want ErrNotFound", err)
	}
}

func TestMultipartListPartsNoSuchUpload(t *testing.T) {
	s, cleanup := newTestStorage(t)
	defer cleanup()
	if _, err := s.Multipart.ListParts(context.Background(), "mp/none", "bogus-upload-id"); !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("ListParts bogus uploadID: got %v want ErrNotFound", err)
	}
}
