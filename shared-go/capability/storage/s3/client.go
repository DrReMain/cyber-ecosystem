package s3

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"cyber-ecosystem/shared-go/capability/storage"
)

// NewClient builds an S3 client, validates config, and ensures the default
// bucket exists. Credentials follow the aws default chain unless both
// AccessKey and SecretKey are set. The cleanup is a no-op (the client owns no
// closeable resource). The second return is the browser-facing presign client
// (nil when PublicEndpoint is unset — mint on the data-plane base instead).
//
// opts are applied after the config-driven options; the platform layer passes
// observability.S3Options() here, keeping the dependency one-way
// (platform → storage).
func NewClient(cfg *Config, opts ...func(*s3.Options)) (*s3.Client, *s3.PresignClient, func(), error) {
	if cfg == nil {
		return nil, nil, nil, fmt.Errorf("%w: nil config", storage.ErrInvalidArgument)
	}
	if cfg.Bucket == "" {
		return nil, nil, nil, fmt.Errorf("%w: bucket is required", storage.ErrInvalidArgument)
	}
	if err := cfg.validate(); err != nil {
		return nil, nil, nil, err
	}

	// Credential resolution (AWS IMDS / SSO / profile) can be slow on first
	// call — give it its own budget separate from the bucket ping.
	loadCtx, loadCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer loadCancel()
	awsCfg, err := config.LoadDefaultConfig(loadCtx, config.WithRegion(cfg.Region))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w: load aws config: %w", storage.ErrUnavailable, err)
	}
	if cfg.AccessKey != "" && cfg.SecretKey != "" {
		awsCfg.Credentials = credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		o.UsePathStyle = cfg.UsePathStyle
		// aws-sdk-go-v2 (core v1.43+) defaults to computing CRC32 integrity
		// checksums on every supported operation (WhenSupported). For presigned
		// URLs that means a browser doing a plain PUT without the checksum
		// header fails with a signature or validation error — the bytes never
		// carried one. Both knobs to WhenRequired (compute only when the
		// operation mandates it): presign targets non-SDK clients, and
		// integrity closes at confirm time (HEAD + declared-size check), not
		// in transport headers.
		o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
		o.ResponseChecksumValidation = aws.ResponseChecksumValidationWhenRequired
		for _, opt := range opts {
			opt(o)
		}
	})

	bucketCtx, bucketCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer bucketCancel()
	if err := createBucketIfAbsent(bucketCtx, client, cfg.Bucket, cfg.Region); err != nil {
		return nil, nil, nil, fmt.Errorf("%w: ensure bucket: %w", storage.ErrUnavailable, err)
	}

	// Browser-facing presign base: SigV4 signs the host, so URLs meant for
	// clients must be minted against the advertised endpoint. Same aws config
	// (credentials/region/checksum knobs), only the base differs.
	var publicPresigner *s3.PresignClient
	if cfg.PublicEndpoint != "" && cfg.PublicEndpoint != cfg.Endpoint {
		publicClient := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.PublicEndpoint)
			o.UsePathStyle = cfg.UsePathStyle
			o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
			o.ResponseChecksumValidation = aws.ResponseChecksumValidationWhenRequired
			for _, opt := range opts {
				opt(o)
			}
		})
		publicPresigner = s3.NewPresignClient(publicClient)
	}
	return client, publicPresigner, func() {}, nil
}
