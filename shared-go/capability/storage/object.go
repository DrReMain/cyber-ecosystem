package storage

import (
	"context"
	"io"
	"time"
)

// ObjectInfo is the metadata returned for an object. Metadata is the
// user-supplied/custom object metadata (x-amz-meta-*), distinct from
// ContentType.
type ObjectInfo struct {
	Key          string
	Size         int64
	ContentType  string
	ETag         string
	LastModified time.Time
	Metadata     map[string]string
}

// Object is the core object lifecycle. Upload streams from r; size>0 is a
// pre-check hint (size>max → ErrSizeExceeded before reading) and a
// Content-Length optimization, size<0 means unknown (enforced only by the
// streaming counter). Download returns an io.ReadCloser the caller MUST close.
type Object interface {
	Upload(ctx context.Context, key string, r io.Reader, size int64, contentType string) (*ObjectInfo, error)
	Download(ctx context.Context, key string) (io.ReadCloser, *ObjectInfo, error)
	Delete(ctx context.Context, key string) error
	Stat(ctx context.Context, key string) (*ObjectInfo, error) // HeadObject — no body
	Copy(ctx context.Context, srcKey, dstKey string) (*ObjectInfo, error)
}
