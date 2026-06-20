package platform

import (
	"fmt"

	"cyber-ecosystem/shared-go/kratos/observability"
	"cyber-ecosystem/shared-go/storage"
	storageS3 "cyber-ecosystem/shared-go/storage/s3"

	"cyber-ecosystem/app/services/edge_mobile/internal/conf"
)

func NewStorage(c *conf.Data) (*storage.Storage, func(), error) {
	sc := c.GetStorage()
	if sc == nil {
		return nil, nil, fmt.Errorf("storage config is required")
	}
	cfg := toStorageConfig(sc)
	// S3Options() appends the aws-sdk-v2 OTel middlewares (trace + metrics) to
	// the client; storage/s3 stays free of the observability import.
	client, closeFn, err := storageS3.NewClient(&cfg, observability.S3Options())
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
