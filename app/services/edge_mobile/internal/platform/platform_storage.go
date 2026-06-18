package platform

import (
	"fmt"

	"cyber-ecosystem/shared-go/storage"
	storageS3 "cyber-ecosystem/shared-go/storage/s3"

	"cyber-ecosystem/app/services/edge_mobile/internal/conf"
)

// NewStorage builds the S3-backed storage container for edge_mobile. Returning
// the cleanup from the provider lets wire chain it for graceful shutdown and
// partial-injection rollback; the S3 client owns no closeable resource, so the
// cleanup is a no-op.
func NewStorage(c *conf.Data) (*storage.Storage, func(), error) {
	sc := c.GetStorage()
	if sc == nil {
		return nil, nil, fmt.Errorf("storage config is required")
	}
	cfg := toStorageConfig(sc)
	client, closeFn, err := storageS3.NewClient(&cfg)
	if err != nil {
		return nil, nil, err
	}
	return storageS3.New(client, cfg), closeFn, nil
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
		MaxFileSize:        sc.GetMaxFileSize(),
		MultipartThreshold: s3c.GetMultipartThreshold(),
		PartSize:           s3c.GetPartSize(),
		PresignTTL:         s3c.GetPresignTtl().AsDuration(), // durationpb.AsDuration is nil-safe (→0 → default TTL)
	}
}
