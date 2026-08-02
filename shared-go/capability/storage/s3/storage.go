package s3

import (
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"cyber-ecosystem/shared-go/capability/storage"
)

// backend holds the S3 dependencies shared by every sub-interface impl. It does
// not hold a bucket: each View binds its own bucket, so one backend serves the
// default bucket and any number of tenant buckets.
type backend struct {
	client   *s3.Client
	uploader *transfermanager.Client
	cfg      Config
}

// view returns a bucket-scoped object surface bound to the given bucket.
func (b *backend) view(bucket string) *storage.View {
	return &storage.View{
		Object:    newObject(b, bucket),
		List:      newList(b, bucket),
		Presign:   newPresign(b, bucket),
		Multipart: newMultipart(b, bucket),
	}
}

// New wires the default (configured) bucket view + a bucket manager + a
// per-bucket view factory. The default view is the configured bucket
// (auto-created at NewClient); For(bucket) yields tenant-scoped views for
// multi-tenant use.
func New(client *s3.Client, cfg Config) *storage.Storage {
	b := &backend{
		client: client,
		uploader: transfermanager.New(client, func(o *transfermanager.Options) {
			o.PartSizeBytes = partSizeOrDefault(cfg.PartSize)
			o.MultipartUploadThreshold = multipartThresholdOrDefault(cfg.MultipartThreshold)
		}),
		cfg: cfg,
	}
	defaultView := b.view(cfg.Bucket)
	return storage.NewStorage(*defaultView, &bucketSVC{b: b}, func(name string) *storage.View {
		return b.view(name)
	})
}
