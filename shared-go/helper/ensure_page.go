package helper

import (
	commonv1 "cyber-ecosystem/gen/go/cyber/shared/common/v1"
)

// EnsurePageRequest returns a default PageRequest if request is nil.
func EnsurePageRequest(request *commonv1.PageRequest) *commonv1.PageRequest {
	if request == nil {
		return &commonv1.PageRequest{}
	}
	return request
}
