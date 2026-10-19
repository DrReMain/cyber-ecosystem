package platform

import (
	"context"
	"fmt"
	"sync"

	"cyber-ecosystem/shared-go/capability/storage"
	storageS3 "cyber-ecosystem/shared-go/capability/storage/s3"
	"cyber-ecosystem/shared-go/kratos/observability"
	"cyber-ecosystem/shared-go/kratos/security"

	"cyber-ecosystem/app/services/system/internal/conf"
)

// TenantStorage routes the object plane per tenancy domain. The platform
// tenant (security.DefaultTenant, the seeded registry row) rides the
// configured default bucket; every other domain rides the default bucket's
// name with the domain suffixed, provisioned lazily once per process. The
// resolution is the tenant mixin's precedence — an explicit domain switch
// over the subject — so a pass over a domain's rows hits the same domain's
// bucket.
type TenantStorage struct {
	*storage.Storage

	defaultBucket string
	ensured       sync.Map // bucket name → struct{} (provisioned once per process)
}

func (t *TenantStorage) ForTenant(ctx context.Context) *storage.View {
	tenant := security.TenantFromCtx(ctx)
	if tenant == security.DefaultTenant {
		return &t.View
	}
	bucket := t.defaultBucket + "-" + tenant
	if _, loaded := t.ensured.LoadOrStore(bucket, struct{}{}); !loaded {
		// Best-effort provisioning: a failure un-marks the bucket and lets
		// the operation itself surface the classified error.
		if err := t.Bucket.Create(ctx, bucket); err != nil {
			t.ensured.Delete(bucket)
		}
	}
	return t.For(bucket)
}

func NewStorage(c *conf.Data) (*TenantStorage, func(), error) {
	sc := c.GetStorage()
	if sc == nil {
		return nil, nil, fmt.Errorf("storage config is required")
	}
	cfg := toStorageConfig(sc)
	client, publicPresigner, closeFn, err := storageS3.NewClient(&cfg, observability.S3Options())
	if err != nil {
		return nil, nil, err
	}
	return &TenantStorage{Storage: storageS3.New(client, publicPresigner, cfg), defaultBucket: cfg.Bucket}, closeFn, nil
}

func toStorageConfig(sc *conf.Data_Storage) storageS3.Config {
	s3c := sc.GetS3()
	return storageS3.Config{
		Endpoint:           s3c.GetEndpoint(),
		AccessKey:          s3c.GetAccessKey(),
		SecretKey:          s3c.GetSecretKey(),
		Bucket:             s3c.GetBucket(),
		Region:             s3c.GetRegion(),
		UsePathStyle:       s3c.GetUsePathStyle(),
		MaxFileSize:        s3c.GetMaxFileSize(),
		MultipartThreshold: s3c.GetMultipartThreshold(),
		PartSize:           s3c.GetPartSize(),
		PresignTTL:         s3c.GetPresignTtl().AsDuration(), // durationpb.AsDuration is nil-safe (→0 → default TTL)
		PublicEndpoint:     s3c.GetPublicEndpoint(),
	}
}
