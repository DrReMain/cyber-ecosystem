package s3

import (
	"errors"
	"fmt"

	"github.com/aws/smithy-go"

	"cyber-ecosystem/shared-go/storage"
)

// mapError translates an aws-sdk-go-v2 error into a storage sentinel. ErrSizeExceeded
// is NOT produced here — it comes from sizeLimitReader in-process.
func mapError(err error, op string) error {
	if err == nil {
		return nil
	}
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "NoSuchKey", "NotFound", "NoSuchUpload", "NoSuchBucket":
			return fmt.Errorf("%s: %w", op, storage.ErrNotFound)
		case "AccessDenied", "Forbidden":
			return fmt.Errorf("%s: %w", op, storage.ErrForbidden)
		case "InvalidPart", "InvalidArgument", "InvalidPartOrder", "EntityTooSmall", "BucketNotEmpty":
			return fmt.Errorf("%s: %w", op, storage.ErrInvalidArgument)
		case "SlowDown", "ServiceUnavailable", "InternalError", "RequestTimeout":
			// transient/throttling → caller may retry; map to Unavailable
			return fmt.Errorf("%s: %w", op, storage.ErrUnavailable)
		}
	}
	return fmt.Errorf("%s: %w", op, err)
}
