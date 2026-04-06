package utils

import (
	"cyber-ecosystem/contracts/go/common"
)

// EnsurePageRequest returns a default PageRequest if request is nil.
func EnsurePageRequest(request *common.PageRequest) *common.PageRequest {
	if request == nil {
		return &common.PageRequest{}
	}
	return request
}
