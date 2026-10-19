package storage

import (
	"context"
	"time"
)

// Presign produces time-limited pre-signed URLs so clients can upload/download
// directly to/from the backend without proxying bytes through the service.
type Presign interface {
	// PresignUpload returns a PUT URL. Content-Type is NOT bound into the
	// signature (clients upload any format); size is a pre-check only — the
	// URL enforces no Content-Length.
	PresignUpload(ctx context.Context, key string, size int64, ttl time.Duration) (string, error)
	// PresignDownload returns a GET URL; opts are S3 response-header overrides
	// signed into the URL (attachment filename, inline disposition,
	// content-type). The zero value serves the object's stored headers.
	PresignDownload(ctx context.Context, key string, ttl time.Duration, opts DownloadOptions) (string, error)
	PresignUploadPart(ctx context.Context, key, uploadID string, partNum int32, ttl time.Duration) (string, error) // PUT part URL (client-direct multipart)
}

// DownloadOptions carry response-header overrides for a presigned download:
// how the browser should treat the bytes (save-as vs inline preview) and what
// filename to use. Empty fields are omitted from the signature.
type DownloadOptions struct {
	ContentType        string
	ContentDisposition string
}
