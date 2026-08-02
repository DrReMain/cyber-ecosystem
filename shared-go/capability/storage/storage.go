package storage

// View is a bucket-scoped object surface: the four object sub-interfaces
// operating within a single bucket. Storage embeds a View bound to the default
// (configured) bucket; For returns a View bound to any bucket.
type View struct {
	Object    Object
	List      List
	Presign   Presign
	Multipart Multipart
}

// Storage is the storage capability root. The embedded View is the default
// (configured) bucket — single-bucket callers use s.Object.Upload directly.
// Bucket manages buckets at runtime; For returns a per-bucket view for
// multi-tenant use (one bucket per tenant):
//
//	tenant := platform.GetStorage().For("tenant-42")
//	tenant.Object.Upload(ctx, key, ...)
type Storage struct {
	View          // default (configured) bucket
	Bucket Bucket // bucket CRUD: Create/Exists/Delete/List

	scopeFn func(string) *View // backend-provided per-bucket view factory
}

// NewStorage assembles a Storage from a default-bucket view, a bucket manager,
// and a per-bucket view factory. Backends call this; the scope closure stays off
// the public surface. Callers use the fields and For, never scopeFn.
func NewStorage(defaultView View, buckets Bucket, scope func(string) *View) *Storage {
	return &Storage{View: defaultView, Bucket: buckets, scopeFn: scope}
}

// For returns a bucket-scoped view for multi-tenant use (one bucket per
// tenant). Backends MUST supply a scope factory via NewStorage — the s3 backend
// always does, so For never returns nil in practice; the nil guard is purely
// defensive against a misconfigured backend.
func (s *Storage) For(bucket string) *View {
	if s.scopeFn == nil {
		return nil
	}
	return s.scopeFn(bucket)
}
