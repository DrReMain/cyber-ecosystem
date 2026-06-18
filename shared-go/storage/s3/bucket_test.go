package s3

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"cyber-ecosystem/shared-go/storage"
)

func TestBucketCRUD(t *testing.T) {
	s, cleanup := newTestStorage(t)
	defer cleanup()
	ctx := context.Background()
	bm := s.Bucket
	const bucket = "bucket-crud-test"

	_ = bm.Delete(ctx, bucket) // idempotent prep for leftover state
	if err := bm.Create(ctx, bucket); err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = bm.Delete(ctx, bucket) })

	exists, err := bm.Exists(ctx, bucket)
	if err != nil || !exists {
		t.Fatalf("Exists after create: exists=%v err=%v", exists, err)
	}
	// Create is idempotent: a second Create on the same bucket is not an error.
	if err := bm.Create(ctx, bucket); err != nil {
		t.Fatalf("Create idempotent: %v", err)
	}
	listed, err := bm.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	found := false
	for _, b := range listed {
		if b.Name == bucket {
			found = true
		}
	}
	if !found {
		t.Fatalf("List did not include %s", bucket)
	}
	if err := bm.Delete(ctx, bucket); err != nil {
		t.Fatalf("Delete (empty): %v", err)
	}
	if exists, _ := bm.Exists(ctx, bucket); exists {
		t.Fatalf("Exists after delete: want false")
	}
}

func TestBucketDeleteNonEmptyFails(t *testing.T) {
	s, cleanup := newTestStorage(t)
	defer cleanup()
	ctx := context.Background()
	const bucket = "bucket-nonempty-test"

	_ = s.Bucket.Delete(ctx, bucket)
	if err := s.Bucket.Create(ctx, bucket); err != nil {
		t.Fatalf("Create: %v", err)
	}
	v := s.For(bucket)
	if v == nil {
		t.Fatal("For returned nil")
	}
	t.Cleanup(func() {
		_ = v.Object.Delete(ctx, "keep")
		_ = s.Bucket.Delete(ctx, bucket)
	})
	if _, err := v.Object.Upload(ctx, "keep", bytes.NewReader([]byte("x")), 1, "text/plain"); err != nil {
		t.Fatalf("Upload: %v", err)
	}
	// Non-empty bucket must NOT be deletable (S3 native BucketNotEmpty).
	if err := s.Bucket.Delete(ctx, bucket); !errors.Is(err, storage.ErrInvalidArgument) {
		t.Errorf("Delete non-empty: got %v, want ErrInvalidArgument (BucketNotEmpty)", err)
	}
}
