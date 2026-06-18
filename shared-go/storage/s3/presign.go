package s3

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"cyber-ecosystem/shared-go/storage"
)

type presignSVC struct {
	b      *backend
	bucket string
}

func newPresign(b *backend, bucket string) storage.Presign { return &presignSVC{b: b, bucket: bucket} }

func (p *presignSVC) PresignUpload(ctx context.Context, key string, size int64, ttl time.Duration) (string, error) {
	if err := storage.ValidateKey(key); err != nil {
		return "", err
	}
	if p.b.cfg.MaxFileSize > 0 && size > p.b.cfg.MaxFileSize {
		return "", storage.ErrSizeExceeded // server-side pre-check only; URL enforces no Content-Length
	}
	// Content-Type is deliberately NOT set: binding it into SigV4 would force
	// clients to echo the exact header or get SignatureDoesNotMatch. Leaving it
	// out lets clients upload any format.
	req, err := s3.NewPresignClient(p.b.client).PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	}, func(o *s3.PresignOptions) { o.Expires = presignTTLOrDefault(ttl) })
	if err != nil {
		return "", mapError(err, "presign upload")
	}
	return req.URL, nil
}

func (p *presignSVC) PresignDownload(ctx context.Context, key string, ttl time.Duration) (string, error) {
	if err := storage.ValidateKey(key); err != nil {
		return "", err
	}
	req, err := s3.NewPresignClient(p.b.client).PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	}, func(o *s3.PresignOptions) { o.Expires = presignTTLOrDefault(ttl) })
	if err != nil {
		return "", mapError(err, "presign download")
	}
	return req.URL, nil
}

func (p *presignSVC) PresignUploadPart(ctx context.Context, key, uploadID string, partNum int32, ttl time.Duration) (string, error) {
	if err := storage.ValidateKey(key); err != nil {
		return "", err
	}
	if uploadID == "" || partNum < 1 || partNum > 10000 {
		return "", storage.ErrInvalidArgument
	}
	req, err := s3.NewPresignClient(p.b.client).PresignUploadPart(ctx, &s3.UploadPartInput{
		Bucket:     aws.String(p.bucket),
		Key:        aws.String(key),
		UploadId:   aws.String(uploadID),
		PartNumber: aws.Int32(partNum),
	}, func(o *s3.PresignOptions) { o.Expires = presignTTLOrDefault(ttl) })
	if err != nil {
		return "", mapError(err, "presign upload part")
	}
	return req.URL, nil
}
