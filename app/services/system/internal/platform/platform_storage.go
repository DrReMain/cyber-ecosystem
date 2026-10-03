package platform

import (
	"fmt"

	"cyber-ecosystem/shared-go/capability/storage"
	storageS3 "cyber-ecosystem/shared-go/capability/storage/s3"
	"cyber-ecosystem/shared-go/kratos/observability"

	"cyber-ecosystem/app/services/system/internal/conf"
)

func NewStorage(c *conf.Data) (*storage.Storage, func(), error) {
	sc := c.GetStorage()
	if sc == nil {
		return nil, nil, fmt.Errorf("storage config is required")
	}
	cfg := toStorageConfig(sc)
	client, publicPresigner, closeFn, err := storageS3.NewClient(&cfg, observability.S3Options())
	if err != nil {
		return nil, nil, err
	}
	return storageS3.New(client, publicPresigner, cfg), closeFn, nil
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
