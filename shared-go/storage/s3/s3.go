package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"go.opentelemetry.io/otel/attribute"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/shared-go/storage"
)

type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	Region    string
	Logger    log.Logger
}

type S3Storage struct {
	client *s3.Client
	bucket string
	logger log.Logger
}

func New(cfg Config) (*S3Storage, error) {
	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")),
		config.WithRegion(cfg.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = true
	})
	return &S3Storage{
		client: client,
		bucket: cfg.Bucket,
		logger: cfg.Logger,
	}, nil
}

func (s *S3Storage) Upload(ctx context.Context, id string, name string, reader io.Reader, size int64, contentType string) (*storage.FileInfo, error) {
	start := time.Now()
	ctx, span := startSpan(ctx, "upload",
		attribute.String("s3.bucket", s.bucket),
		attribute.String("s3.key", id),
		attribute.Int64("s3.size", size),
		attribute.String("s3.content_type", contentType),
	)
	defer span.End()

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(id),
		Body:          reader,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
		Metadata: map[string]string{
			"original-name": name,
		},
	})
	latency := time.Since(start).Seconds()
	if err != nil {
		recordError(span, err)
		s.logOperation(ctx, "upload", latency, err, "s3.bucket", s.bucket, "s3.key", id, "s3.size", size)
		return nil, mapError(err, "upload")
	}
	s.logOperation(ctx, "upload", latency, nil, "s3.bucket", s.bucket, "s3.key", id, "s3.size", size)
	return &storage.FileInfo{
		ID:          id,
		Name:        name,
		ContentType: contentType,
		Size:        size,
	}, nil
}

func (s *S3Storage) Download(ctx context.Context, id string) (io.ReadCloser, *storage.FileInfo, error) {
	start := time.Now()
	ctx, span := startSpan(ctx, "download",
		attribute.String("s3.bucket", s.bucket),
		attribute.String("s3.key", id),
	)
	defer span.End()

	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(id),
	})
	latency := time.Since(start).Seconds()
	if err != nil {
		recordError(span, err)
		s.logOperation(ctx, "download", latency, err, "s3.bucket", s.bucket, "s3.key", id)
		return nil, nil, mapError(err, "download")
	}

	name := id
	if v, ok := resp.Metadata["original-name"]; ok {
		name = v
	}

	s.logOperation(ctx, "download", latency, nil, "s3.bucket", s.bucket, "s3.key", id, "s3.size", aws.ToInt64(resp.ContentLength))
	return resp.Body, &storage.FileInfo{
		ID:          id,
		Name:        name,
		ContentType: aws.ToString(resp.ContentType),
		Size:        aws.ToInt64(resp.ContentLength),
	}, nil
}

func (s *S3Storage) Delete(ctx context.Context, id string) error {
	start := time.Now()
	ctx, span := startSpan(ctx, "delete",
		attribute.String("s3.bucket", s.bucket),
		attribute.String("s3.key", id),
	)
	defer span.End()

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(id),
	})
	latency := time.Since(start).Seconds()
	if err != nil {
		recordError(span, err)
		s.logOperation(ctx, "delete", latency, err, "s3.bucket", s.bucket, "s3.key", id)
		return mapError(err, "delete")
	}
	s.logOperation(ctx, "delete", latency, nil, "s3.bucket", s.bucket, "s3.key", id)
	return nil
}

func (s *S3Storage) logOperation(ctx context.Context, op string, latency float64, err error, attrs ...any) {
	if s.logger == nil {
		return
	}
	logger := log.WithContext(ctx, s.logger)
	fields := []any{
		"msg", "Storage operation",
		"component", "storage",
		"backend", "s3",
		"operation", op,
		"latency", latency,
	}
	fields = append(fields, attrs...)
	if err != nil {
		fields = append(fields, "error", err.Error())
		_ = logger.Log(log.LevelError, fields...)
		return
	}
	_ = logger.Log(log.LevelInfo, fields...)
}

func mapError(err error, op string) error {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		code := apiErr.ErrorCode()
		switch {
		case code == "NoSuchKey" || code == "NotFound" || code == "404":
			return fmt.Errorf("%s: %w", op, storage.ErrNotFound)
		case code == "AccessDenied" || code == "Forbidden" || code == "403":
			return fmt.Errorf("%s: %w", op, storage.ErrForbidden)
		}
	}
	return fmt.Errorf("%s: %w", op, err)
}
