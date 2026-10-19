package s3

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"cyber-ecosystem/shared-go/capability/storage"
)

type presignSvc struct {
	b      *backend
	bucket string
}

func newPresign(b *backend, bucket string) storage.Presign { return &presignSvc{b: b, bucket: bucket} }

// signer picks the base URLs are minted against: the advertised (public)
// endpoint when configured — SigV4 signs the host, so the browser-facing base
// must do the signing.
func (p *presignSvc) signer() *s3.PresignClient {
	if p.b.publicPresigner != nil {
		return p.b.publicPresigner
	}
	return p.b.presigner
}

func (p *presignSvc) PresignUpload(ctx context.Context, key string, size int64, ttl time.Duration) (string, error) {
	if err := storage.ValidateKey(key); err != nil {
		return "", err
	}
	if p.b.cfg.MaxFileSize > 0 && size > p.b.cfg.MaxFileSize {
		return "", storage.ErrSizeExceeded // server-side pre-check only; URL enforces no Content-Length
	}
	req, err := p.signer().PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	}, func(o *s3.PresignOptions) { o.Expires = presignTTLOrDefault(ttl) })
	if err != nil {
		return "", mapError(err, "presign upload")
	}
	return req.URL, nil
}

func (p *presignSvc) PresignDownload(ctx context.Context, key string, ttl time.Duration, opts storage.DownloadOptions) (string, error) {
	if err := storage.ValidateKey(key); err != nil {
		return "", err
	}
	in := &s3.GetObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	}
	if opts.ContentType != "" {
		in.ResponseContentType = aws.String(opts.ContentType)
	}
	if opts.ContentDisposition != "" {
		in.ResponseContentDisposition = aws.String(opts.ContentDisposition)
	}
	req, err := p.signer().PresignGetObject(ctx, in, func(o *s3.PresignOptions) {
		o.Expires = presignTTLOrDefault(ttl)
	})
	if err != nil {
		return "", mapError(err, "presign download")
	}
	return req.URL, nil
}

func (p *presignSvc) PresignUploadPart(ctx context.Context, key, uploadID string, partNum int32, ttl time.Duration) (string, error) {
	if err := storage.ValidateKey(key); err != nil {
		return "", err
	}
	if uploadID == "" || partNum < 1 || partNum > 10000 {
		return "", storage.ErrInvalidArgument
	}
	req, err := p.signer().PresignUploadPart(ctx, &s3.UploadPartInput{
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
