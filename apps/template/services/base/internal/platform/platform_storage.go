package platform

import (
	"fmt"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/shared-go/storage"
	s3storage "cyber-ecosystem/shared-go/storage/s3"

	"cyber-ecosystem/apps/template/services/base/internal/conf"
)

func NewStorage(c *conf.Data, logger log.Logger) (storage.Storage, error) {
	if c.Storage == nil || c.Storage.S3 == nil {
		return nil, fmt.Errorf("storage config is required")
	}
	s3Cfg := c.Storage.S3
	return s3storage.New(s3storage.Config{
		Endpoint:  s3Cfg.Endpoint,
		AccessKey: s3Cfg.AccessKey,
		SecretKey: s3Cfg.SecretKey,
		Bucket:    s3Cfg.Bucket,
		Region:    s3Cfg.Region,
		Logger:    logger,
	})
}
