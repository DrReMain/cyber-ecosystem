package s3

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"cyber-ecosystem/shared-go/storage"
)

type multipartSVC struct {
	b      *backend
	bucket string
}

func newMultipart(b *backend, bucket string) storage.Multipart {
	return &multipartSVC{b: b, bucket: bucket}
}

func (m *multipartSVC) Create(ctx context.Context, key, contentType string) (string, error) {
	if err := storage.ValidateKey(key); err != nil {
		return "", err
	}
	out, err := m.b.client.CreateMultipartUpload(ctx, &s3.CreateMultipartUploadInput{
		Bucket:      aws.String(m.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", mapError(err, "multipart create")
	}
	return aws.ToString(out.UploadId), nil
}

func (m *multipartSVC) UploadPart(ctx context.Context, key, uploadID string, partNum int32, r io.Reader, size int64) (string, error) {
	if err := storage.ValidateKey(key); err != nil {
		return "", err
	}
	if uploadID == "" || partNum < 1 || partNum > 10000 {
		return "", storage.ErrInvalidArgument
	}
	// size is a Content-Length hint (lets the SDK size the request and lets the
	// backend reject an obviously-wrong value). Per-part size is NOT enforced
	// against MaxFileSize: parts may be up to 5GiB (S3 limit), and a multipart
	// object's total size is only known at Complete — total-size enforcement is
	// a business-layer concern.
	in := &s3.UploadPartInput{
		Bucket:     aws.String(m.bucket),
		Key:        aws.String(key),
		UploadId:   aws.String(uploadID),
		PartNumber: aws.Int32(partNum),
		Body:       r,
	}
	if size > 0 {
		in.ContentLength = aws.Int64(size)
	}
	out, err := m.b.client.UploadPart(ctx, in)
	if err != nil {
		return "", mapError(err, "multipart upload part")
	}
	return aws.ToString(out.ETag), nil
}

func (m *multipartSVC) ListParts(ctx context.Context, key, uploadID string) ([]storage.CompletedPart, error) {
	if err := storage.ValidateKey(key); err != nil {
		return nil, err
	}
	if uploadID == "" {
		return nil, storage.ErrInvalidArgument
	}
	out, err := m.b.client.ListParts(ctx, &s3.ListPartsInput{
		Bucket:   aws.String(m.bucket),
		Key:      aws.String(key),
		UploadId: aws.String(uploadID),
	})
	if err != nil {
		return nil, mapError(err, "multipart list parts")
	}
	parts := make([]storage.CompletedPart, 0, len(out.Parts))
	for _, p := range out.Parts {
		parts = append(parts, storage.CompletedPart{PartNumber: aws.ToInt32(p.PartNumber), ETag: aws.ToString(p.ETag)})
	}
	return parts, nil
}

func (m *multipartSVC) Complete(ctx context.Context, key, uploadID string, parts []storage.CompletedPart) (*storage.ObjectInfo, error) {
	if err := storage.ValidateKey(key); err != nil {
		return nil, err
	}
	if uploadID == "" {
		return nil, storage.ErrInvalidArgument
	}
	if len(parts) == 0 {
		return nil, storage.ErrInvalidArgument
	}
	completed := make([]types.CompletedPart, len(parts))
	for i, p := range parts {
		completed[i] = types.CompletedPart{PartNumber: aws.Int32(p.PartNumber), ETag: aws.String(p.ETag)}
	}
	out, err := m.b.client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:          aws.String(m.bucket),
		Key:             aws.String(key),
		UploadId:        aws.String(uploadID),
		MultipartUpload: &types.CompletedMultipartUpload{Parts: completed},
	})
	if err != nil {
		return nil, mapError(err, "multipart complete")
	}
	return &storage.ObjectInfo{
		Key:  aws.ToString(out.Key),
		ETag: aws.ToString(out.ETag),
	}, nil
}

func (m *multipartSVC) Abort(ctx context.Context, key, uploadID string) error {
	if err := storage.ValidateKey(key); err != nil {
		return err
	}
	if uploadID == "" {
		return storage.ErrInvalidArgument
	}
	_, err := m.b.client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(m.bucket),
		Key:      aws.String(key),
		UploadId: aws.String(uploadID),
	})
	if err != nil {
		return mapError(err, "multipart abort")
	}
	return nil
}
