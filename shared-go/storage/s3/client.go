package s3

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"cyber-ecosystem/shared-go/storage"
)

// NewClient builds an S3 client against the configured backend (MinIO or any
// S3-compatible cloud). Credentials follow the aws default chain unless both
// AccessKey and SecretKey are set (then static creds override — dev/MinIO). It
// validates config, ensures the default bucket exists, and returns a no-op
// cleanup (the S3 client owns no closeable resource).
//
// opts are applied to the s3.Options AFTER the config-driven options; the
// platform layer passes observability.S3Options() here to attach OTel
// middlewares. This package must NOT import the observability package — the
// dependency is one-way (platform → storage).
func NewClient(cfg *Config, opts ...func(*s3.Options)) (*s3.Client, func(), error) {
	if cfg == nil {
		return nil, nil, fmt.Errorf("%w: nil config", storage.ErrInvalidArgument)
	}
	if cfg.Bucket == "" {
		return nil, nil, fmt.Errorf("%w: bucket is required", storage.ErrInvalidArgument)
	}
	if err := cfg.validate(); err != nil {
		return nil, nil, err
	}

	// Credential resolution (AWS IMDS / SSO / profile) can be slow on first
	// call — give it its own budget separate from the bucket ping.
	loadCtx, loadCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer loadCancel()
	awsCfg, err := config.LoadDefaultConfig(loadCtx, config.WithRegion(cfg.Region))
	if err != nil {
		return nil, nil, fmt.Errorf("%w: load aws config: %w", storage.ErrUnavailable, err)
	}
	if cfg.AccessKey != "" && cfg.SecretKey != "" {
		awsCfg.Credentials = credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		o.UsePathStyle = cfg.UsePathStyle
		for _, opt := range opts {
			opt(o)
		}
	})

	bucketCtx, bucketCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer bucketCancel()
	if err := ensureBucket(bucketCtx, client, cfg.Bucket, cfg.Region); err != nil {
		return nil, nil, fmt.Errorf("%w: ensure bucket: %w", storage.ErrUnavailable, err)
	}
	return client, func() {}, nil
}

// ensureBucket creates the bucket if it does not already exist. Idempotent:
// BucketAlreadyExists / BucketAlreadyOwnedByYou are treated as success. For AWS
// regions other than us-east-1, CreateBucket requires a LocationConstraint
// (us-east-1 is the default and must NOT specify one, else
// IllegalLocationConstraintException).
func ensureBucket(ctx context.Context, client *s3.Client, bucket, region string) error {
	if _, err := client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(bucket)}); err == nil {
		return nil
	}
	in := &s3.CreateBucketInput{Bucket: aws.String(bucket)}
	if region != "" && region != "us-east-1" {
		in.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(region),
		}
	}
	_, err := client.CreateBucket(ctx, in)
	if err != nil {
		var bae *types.BucketAlreadyExists
		var boy *types.BucketAlreadyOwnedByYou
		if errors.As(err, &bae) || errors.As(err, &boy) {
			return nil
		}
		return err
	}
	return nil
}
