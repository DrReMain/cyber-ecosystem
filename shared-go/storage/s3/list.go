package s3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"cyber-ecosystem/shared-go/storage"
)

type listSVC struct {
	b      *backend
	bucket string
}

func newList(b *backend, bucket string) storage.List { return &listSVC{b: b, bucket: bucket} }

func (l *listSVC) List(ctx context.Context, prefix, cursor string, maxKeys int32) (*storage.ListResult, error) {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(l.bucket),
		Prefix: aws.String(prefix),
	}
	if cursor != "" {
		input.ContinuationToken = aws.String(cursor)
	}
	// Clamp to S3's per-request ceiling: ListObjectsV2 returns at most 1000 keys
	// per call, so a larger maxKeys would be silently truncated.
	switch {
	case maxKeys <= 0 || maxKeys > 1000:
		input.MaxKeys = aws.Int32(1000)
	default:
		input.MaxKeys = aws.Int32(maxKeys)
	}
	out, err := l.b.client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, mapError(err, "list")
	}
	objects := make([]storage.ObjectInfo, 0, len(out.Contents))
	for _, o := range out.Contents {
		objects = append(objects, storage.ObjectInfo{
			Key:          aws.ToString(o.Key),
			Size:         aws.ToInt64(o.Size),
			ETag:         aws.ToString(o.ETag),
			LastModified: aws.ToTime(o.LastModified),
		})
	}
	return &storage.ListResult{Objects: objects, NextCursor: aws.ToString(out.NextContinuationToken)}, nil
}
