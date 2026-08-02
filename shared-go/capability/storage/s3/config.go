package s3

import (
	"fmt"
	"time"
)

// Config maps conf.Data.Storage to aws-sdk-go-v2 options. Backend-agnostic
// storage code never sees this type; only the s3 adapter and the service
// platform layer do. Zero-value Duration/size fields fall back to the defaults
// below inside NewClient / New.
type Config struct {
	Endpoint           string
	AccessKey          string // empty → aws default credential chain
	SecretKey          string // empty → default chain
	Bucket             string
	Region             string
	UsePathStyle       bool          // MinIO=true, AWS S3=false
	MaxFileSize        int64         // bytes; 0 → no limit
	MultipartThreshold int64         // bytes; 0 → 8MiB
	PartSize           int64         // bytes; 0 → 5MiB (S3 minimum)
	PresignTTL         time.Duration // 0 → 15m
}

const (
	defaultMultipartThreshold int64 = 8 << 20 // 8 MiB
	defaultPartSize           int64 = 5 << 20 // 5 MiB (S3 minimum part size)
	defaultPresignTTL               = 15 * time.Minute
	maxPresignTTL                   = 7 * 24 * time.Hour // AWS SigV4 caps presigned URLs at 7 days; MinIO tolerates longer.
	minPartSize               int64 = 5 << 20            // S3 minimum part size (non-last parts)
)

// validate checks S3 invariants so misconfiguration fails at boot, not as a
// runtime EntityTooSmall mid-upload.
func (c Config) validate() error {
	if c.PartSize > 0 && c.PartSize < minPartSize {
		return fmt.Errorf("storage: PartSize %d < 5MiB S3 minimum", c.PartSize)
	}
	return nil
}

func partSizeOrDefault(v int64) int64 {
	if v > 0 {
		return v
	}
	return defaultPartSize
}

func multipartThresholdOrDefault(v int64) int64 {
	if v > 0 {
		return v
	}
	return defaultMultipartThreshold
}

func presignTTLOrDefault(v time.Duration) time.Duration {
	if v <= 0 {
		return defaultPresignTTL
	}
	if v > maxPresignTTL {
		return maxPresignTTL
	}
	return v
}
