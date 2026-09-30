package s3

import (
	"context"
	"errors"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/capability/storage"
)

// Argument guards must reject invalid input before any network call, so these
// run against the zero-endpoint newTestBackend.
func TestMultipartArgumentGuards(t *testing.T) {
	m := newMultipart(newTestBackend(), "t")
	ctx := context.Background()

	if _, err := m.Create(ctx, "", "text/plain"); !errors.Is(err, storage.ErrInvalidArgument) {
		t.Errorf("Create(empty key): %v", err)
	}
	if _, err := m.UploadPart(ctx, "k", "", 1, nil, 0); !errors.Is(err, storage.ErrInvalidArgument) {
		t.Errorf("UploadPart(empty uploadID): %v", err)
	}
	if _, err := m.UploadPart(ctx, "k", "id", 0, nil, 0); !errors.Is(err, storage.ErrInvalidArgument) {
		t.Errorf("UploadPart(partNum=0): %v", err)
	}
	if _, err := m.UploadPart(ctx, "k", "id", 10001, nil, 0); !errors.Is(err, storage.ErrInvalidArgument) {
		t.Errorf("UploadPart(partNum>10000): %v", err)
	}
	if _, err := m.ListParts(ctx, "k", ""); !errors.Is(err, storage.ErrInvalidArgument) {
		t.Errorf("ListParts(empty uploadID): %v", err)
	}
	if _, err := m.Complete(ctx, "k", "id", nil); !errors.Is(err, storage.ErrInvalidArgument) {
		t.Errorf("Complete(no parts): %v", err)
	}
	if err := m.Abort(ctx, "k", ""); !errors.Is(err, storage.ErrInvalidArgument) {
		t.Errorf("Abort(empty uploadID): %v", err)
	}
}

func TestPresignArgumentGuards(t *testing.T) {
	b := newTestBackend()
	b.cfg.MaxFileSize = 100
	p := newPresign(b, "t")
	ctx := context.Background()

	if _, err := p.PresignUpload(ctx, "", 10, time.Minute); !errors.Is(err, storage.ErrInvalidArgument) {
		t.Errorf("PresignUpload(empty key): %v", err)
	}
	if _, err := p.PresignUpload(ctx, "k", 101, time.Minute); !errors.Is(err, storage.ErrSizeExceeded) {
		t.Errorf("PresignUpload(over max): %v", err)
	}
	if _, err := p.PresignUploadPart(ctx, "k", "id", 0, time.Minute); !errors.Is(err, storage.ErrInvalidArgument) {
		t.Errorf("PresignUploadPart(partNum=0): %v", err)
	}
	if _, err := p.PresignDownload(ctx, "", time.Minute, storage.DownloadOptions{}); !errors.Is(err, storage.ErrInvalidArgument) {
		t.Errorf("PresignDownload(empty key): %v", err)
	}
}

func TestBucketArgumentGuards(t *testing.T) {
	bk := &bucketSvc{b: newTestBackend()}
	ctx := context.Background()

	if err := bk.Create(ctx, ""); !errors.Is(err, storage.ErrInvalidArgument) {
		t.Errorf("Create(empty bucket): %v", err)
	}
	if _, err := bk.Exists(ctx, ""); !errors.Is(err, storage.ErrInvalidArgument) {
		t.Errorf("Exists(empty bucket): %v", err)
	}
	if err := bk.Delete(ctx, ""); !errors.Is(err, storage.ErrInvalidArgument) {
		t.Errorf("Delete(empty bucket): %v", err)
	}
}
