package s3

import (
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"cyber-ecosystem/shared-go/capability/storage"
)

// backend holds the S3 dependencies shared by every sub-interface impl; each
// View binds its own bucket, so one backend serves all buckets.
type backend struct {
	client          *s3.Client
	presigner       *s3.PresignClient
	publicPresigner *s3.PresignClient // browser-facing base; nil = same base as presigner
	uploader        *transfermanager.Client
	cfg             Config
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

// New wires the default (configured) bucket view, a bucket manager, a
// per-bucket view factory, and the resolved limits. publicPresigner is the
// browser-facing presign client from NewClient (nil when no public endpoint
// is configured — presign then mints on the data-plane base).
func New(client *s3.Client, publicPresigner *s3.PresignClient, cfg Config) *storage.Storage {
	b := &backend{
		client:          client,
		presigner:       s3.NewPresignClient(client),
		publicPresigner: publicPresigner,
		uploader: transfermanager.New(client, func(o *transfermanager.Options) {
			o.PartSizeBytes = partSizeOrDefault(cfg.PartSize)
			o.MultipartUploadThreshold = multipartThresholdOrDefault(cfg.MultipartThreshold)
		}),
		cfg: cfg,
	}
	limits := storage.Limits{
		MaxFileSize:        cfg.MaxFileSize,
		MultipartThreshold: multipartThresholdOrDefault(cfg.MultipartThreshold),
		PartSize:           partSizeOrDefault(cfg.PartSize),
		PresignTTL:         presignTTLOrDefault(cfg.PresignTTL),
	}
	defaultView := b.view(cfg.Bucket)
	return storage.NewStorage(*defaultView, &bucketSvc{b: b}, func(name string) *storage.View {
		return b.view(name)
	}, limits)
}
