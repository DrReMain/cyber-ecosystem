package storage

import (
	"context"
	"time"
)

// Presign produces time-limited pre-signed URLs so clients can upload/download
// directly to/from the backend without proxying bytes through the service.
type Presign interface {
	// PresignUpload returns a PUT URL. Content-Type is intentionally NOT bound
	// into the signature, so clients may upload any format (they set their own
	// headers). size is a server-side pre-check only — the URL itself enforces
	// no Content-Length, so a client can stream more or less than size.
	PresignUpload(ctx context.Context, key string, size int64, ttl time.Duration) (string, error)
	PresignDownload(ctx context.Context, key string, ttl time.Duration) (string, error)
	PresignUploadPart(ctx context.Context, key, uploadID string, partNum int32, ttl time.Duration) (string, error) // PUT part URL (client-direct multipart)
}
