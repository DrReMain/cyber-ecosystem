package storage

import (
	"context"
	"io"
	"time"
)

// CompletedPart is a part uploaded to an in-progress multipart upload.
type CompletedPart struct {
	PartNumber int32
	ETag       string
	Size       int64
}

// PendingUpload is an in-progress multipart session visible at the bucket
// level, not tied to any one key. Initiated is the standard S3 field but not
// universal — SeaweedFS omits it (documented in the conformance probe) — so
// consumers must not depend on it alone for age-based decisions.
type PendingUpload struct {
	Key       string
	UploadID  string
	Initiated time.Time
}

// Multipart is the explicit low-level multipart API for client-direct
// resumable large uploads: the service creates an upload, the client uploads
// parts (directly via PresignUploadPart or proxied), and the service completes.
// ListParts is the resume primitive — a reconnecting client learns which parts
// already exist and uploads only the missing ones.
type Multipart interface {
	Create(ctx context.Context, key, contentType string) (uploadID string, err error)
	UploadPart(ctx context.Context, key, uploadID string, partNum int32, r io.Reader, size int64) (etag string, err error)
	ListParts(ctx context.Context, key, uploadID string) ([]CompletedPart, error)
	Complete(ctx context.Context, key, uploadID string, parts []CompletedPart) (*ObjectInfo, error)
	Abort(ctx context.Context, key, uploadID string) error

	// ListMultipartUploads walks every in-progress multipart session whose key
	// starts with prefix (empty = all), streaming one page at a time into visit
	// so callers never hold the full listing in memory. An error from visit
	// stops the walk and is returned wrapped. This is the orphan-sweep
	// primitive: the only surface reaching sessions no metadata row points at.
	ListMultipartUploads(ctx context.Context, prefix string, visit func(PendingUpload) error) error
}
