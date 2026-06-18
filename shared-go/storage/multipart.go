package storage

import (
	"context"
	"io"
)

// CompletedPart is a part uploaded to an in-progress multipart upload.
type CompletedPart struct {
	PartNumber int32
	ETag       string
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
}
