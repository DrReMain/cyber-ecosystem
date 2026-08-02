package storage

import (
	"context"
	"time"
)

// BucketInfo is metadata for a bucket.
type BucketInfo struct {
	Name      string
	CreatedAt time.Time
}

// Bucket manages buckets at runtime — multi-tenant use: each tenant gets its
// own bucket, created/provisioned then operated on via Storage.For.
//
// Create is idempotent (an existing bucket is not an error). Delete fails on a
// non-empty bucket (S3 native behavior); the caller empties it first. Exists
// distinguishes "absent" (false, nil) from a real error.
type Bucket interface {
	Create(ctx context.Context, bucket string) error
	Exists(ctx context.Context, bucket string) (bool, error)
	Delete(ctx context.Context, bucket string) error
	List(ctx context.Context) ([]BucketInfo, error)
}
