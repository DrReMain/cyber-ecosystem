package s3

import (
	"context"
	"io"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"cyber-ecosystem/shared-go/storage"
)

type objectSVC struct {
	b      *backend
	bucket string
}

func newObject(b *backend, bucket string) storage.Object { return &objectSVC{b: b, bucket: bucket} }

// sizeLimitReader counts bytes read and returns ErrSizeExceeded once max is
// crossed. Streaming-safe: enforces max for chunked/unknown-size uploads
// without buffering the whole body. The underlying reader's own error is
// preserved (a real read error wins over the size-limit signal).
type sizeLimitReader struct {
	r   io.Reader
	max int64
	n   int64
}

func (l *sizeLimitReader) Read(p []byte) (int, error) {
	n, err := l.r.Read(p)
	l.n += int64(n)
	if err != nil {
		return n, err
	}
	if l.max > 0 && l.n > l.max {
		return n, storage.ErrSizeExceeded
	}
	return n, nil
}

func (o *objectSVC) Upload(ctx context.Context, key string, r io.Reader, size int64, contentType string) (*storage.ObjectInfo, error) {
	if err := storage.ValidateKey(key); err != nil {
		return nil, err
	}
	if o.b.cfg.MaxFileSize > 0 && size > o.b.cfg.MaxFileSize {
		return nil, storage.ErrSizeExceeded // known-size pre-check
	}
	// Wrap only with a limit: MaxFileSize==0 passes the reader through unchanged,
	// preserving any io.ReadSeeker for transfermanager. The counted bytes
	// backfill Size on the streamed (size<0) path.
	var slr *sizeLimitReader
	body := r
	if o.b.cfg.MaxFileSize > 0 {
		slr = &sizeLimitReader{r: r, max: o.b.cfg.MaxFileSize}
		body = slr
	}
	out, err := o.b.uploader.UploadObject(ctx, &transfermanager.UploadObjectInput{
		Bucket:      aws.String(o.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return nil, mapError(err, "upload")
	}
	info := &storage.ObjectInfo{
		Key:         key,
		ETag:        aws.ToString(out.ETag),
		ContentType: contentType,
	}
	switch {
	case size > 0:
		info.Size = size
	case slr != nil:
		info.Size = slr.n
	}
	return info, nil
}

func (o *objectSVC) Download(ctx context.Context, key string) (io.ReadCloser, *storage.ObjectInfo, error) {
	if err := storage.ValidateKey(key); err != nil {
		return nil, nil, err
	}
	out, err := o.b.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(o.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, nil, mapError(err, "download")
	}
	info := &storage.ObjectInfo{
		Key:          key,
		Size:         aws.ToInt64(out.ContentLength),
		ContentType:  aws.ToString(out.ContentType),
		ETag:         aws.ToString(out.ETag),
		LastModified: aws.ToTime(out.LastModified),
		Metadata:     out.Metadata,
	}
	return out.Body, info, nil
}

func (o *objectSVC) Delete(ctx context.Context, key string) error {
	if err := storage.ValidateKey(key); err != nil {
		return err
	}
	_, err := o.b.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(o.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return mapError(err, "delete")
	}
	return nil
}

func (o *objectSVC) Stat(ctx context.Context, key string) (*storage.ObjectInfo, error) {
	if err := storage.ValidateKey(key); err != nil {
		return nil, err
	}
	out, err := o.b.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(o.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, mapError(err, "stat")
	}
	return &storage.ObjectInfo{
		Key:          key,
		Size:         aws.ToInt64(out.ContentLength),
		ContentType:  aws.ToString(out.ContentType),
		ETag:         aws.ToString(out.ETag),
		LastModified: aws.ToTime(out.LastModified),
		Metadata:     out.Metadata,
	}, nil
}

func (o *objectSVC) Copy(ctx context.Context, srcKey, dstKey string) (*storage.ObjectInfo, error) {
	if err := storage.ValidateKey(srcKey); err != nil {
		return nil, err
	}
	if err := storage.ValidateKey(dstKey); err != nil {
		return nil, err
	}
	out, err := o.b.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(o.bucket),
		Key:        aws.String(dstKey),
		CopySource: aws.String(encodeCopySource(o.bucket, srcKey)),
	})
	if err != nil {
		return nil, mapError(err, "copy")
	}
	return &storage.ObjectInfo{
		Key:  dstKey,
		ETag: aws.ToString(out.CopyObjectResult.ETag),
	}, nil
}

// encodeCopySource builds the X-Amz-Copy-Source value ("bucket/key") with each
// path segment of the key URL-encoded. '/' separators are preserved (valid in
// S3 keys); spaces, '+', '%', and multibyte bytes are percent-encoded.
// aws-sdk-go-v2 does NOT encode CopySource, so this must be done at the call
// site — otherwise CopyObject fails or copies the wrong object for non-trivial
// keys (spaces / unicode / special chars).
func encodeCopySource(bucket, key string) string {
	parts := strings.Split(key, "/")
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	return bucket + "/" + strings.Join(parts, "/")
}
