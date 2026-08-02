package s3

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"cyber-ecosystem/shared-go/capability/storage"
)

type bucketSVC struct{ b *backend }

func (m *bucketSVC) Create(ctx context.Context, bucket string) error {
	if bucket == "" {
		return storage.ErrInvalidArgument
	}
	in := &s3.CreateBucketInput{Bucket: aws.String(bucket)}
	// AWS requires LocationConstraint for any region other than us-east-1 (the
	// default); us-east-1 must NOT specify one. MinIO ignores it.
	if r := m.b.cfg.Region; r != "" && r != "us-east-1" {
		in.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(r),
		}
	}
	_, err := m.b.client.CreateBucket(ctx, in)
	if err != nil {
		// Idempotent: already owned/existing is success (MinIO returns
		// BucketAlreadyOwnedByYou when re-creating).
		var bae *types.BucketAlreadyExists
		var boy *types.BucketAlreadyOwnedByYou
		if errors.As(err, &bae) || errors.As(err, &boy) {
			return nil
		}
		return mapError(err, "bucket create")
	}
	return nil
}

func (m *bucketSVC) Exists(ctx context.Context, bucket string) (bool, error) {
	if bucket == "" {
		return false, storage.ErrInvalidArgument
	}
	_, err := m.b.client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(bucket)})
	if err == nil {
		return true, nil
	}
	// Absent bucket → (false, nil); anything else propagates.
	mapped := mapError(err, "bucket exists")
	if errors.Is(mapped, storage.ErrNotFound) {
		return false, nil
	}
	return false, mapped
}

// Delete fails on a non-empty bucket (S3 native "BucketNotEmpty"); the caller
// empties it first. Deleting a non-existent bucket → ErrNotFound.
func (m *bucketSVC) Delete(ctx context.Context, bucket string) error {
	if bucket == "" {
		return storage.ErrInvalidArgument
	}
	_, err := m.b.client.DeleteBucket(ctx, &s3.DeleteBucketInput{Bucket: aws.String(bucket)})
	if err != nil {
		return mapError(err, "bucket delete")
	}
	return nil
}

func (m *bucketSVC) List(ctx context.Context) ([]storage.BucketInfo, error) {
	out, err := m.b.client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, mapError(err, "bucket list")
	}
	bs := make([]storage.BucketInfo, 0, len(out.Buckets))
	for _, bk := range out.Buckets {
		bs = append(bs, storage.BucketInfo{Name: aws.ToString(bk.Name), CreatedAt: aws.ToTime(bk.CreationDate)})
	}
	return bs, nil
}
