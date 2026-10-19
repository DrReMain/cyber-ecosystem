package storage

import "time"

// View is a bucket-scoped object surface: the four object sub-interfaces
// operating within a single bucket. Storage embeds a View bound to the default
// (configured) bucket; For returns a View bound to any bucket.
type View struct {
	Object    Object
	List      List
	Presign   Presign
	Multipart Multipart
}

// Limits are the caller-relevant operational limits of the configured backend:
// what a module needs to route uploads (single vs multipart), pre-check sizes,
// and plan parts — without reaching into backend config.
type Limits struct {
	MaxFileSize        int64         // bytes; 0 = no limit
	MultipartThreshold int64         // bytes; single PUT below, multipart at/above
	PartSize           int64         // bytes; the backend's planned part size
	PresignTTL         time.Duration // effective default TTL for minted URLs
}

// Storage is the storage capability root. The embedded View is the default
// (configured) bucket; For returns a bucket-scoped view for multi-tenant use:
//
//	tenant := platform.GetStorage().For("tenant-42")
//	tenant.Object.Upload(ctx, key, ...)
type Storage struct {
	View          // default (configured) bucket
	Bucket Bucket // bucket CRUD: Create/Exists/Delete/List

	limits  Limits
	scopeFn func(string) *View // backend-provided per-bucket view factory
}

// NewStorage assembles a Storage from a default-bucket view, a bucket manager,
// a per-bucket view factory, and the resolved limits. Backends call this.
func NewStorage(defaultView View, buckets Bucket, scope func(string) *View, limits Limits) *Storage {
	return &Storage{View: defaultView, Bucket: buckets, limits: limits, scopeFn: scope}
}

// Limits exposes the backend's resolved operational limits (defaults applied).
func (s *Storage) Limits() Limits { return s.limits }

// For returns a bucket-scoped view for multi-tenant use (one bucket per
// tenant). Returns nil only if a backend did not supply a scope factory via
// NewStorage (the s3 backend always does).
func (s *Storage) For(bucket string) *View {
	if s.scopeFn == nil {
		return nil
	}
	return s.scopeFn(bucket)
}
