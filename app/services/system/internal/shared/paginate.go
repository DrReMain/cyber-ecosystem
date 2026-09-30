package shared

import (
	"context"

	"cyber-ecosystem/shared-go/helper"

	commonpb "cyber-ecosystem/gen/go/cyber/shared/common/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
)

func Paginate[Q helper.QueryPaginator[Q]](ctx context.Context, query Q, req *commonpb.PageRequest, maxSize int) (total, offset, limit int, err error) {
	return helper.ApplyPagination(ctx, query, req,
		helper.NewPageConfig(helper.DefaultPageSize, maxSize),
		errorspb.ErrorGeneralErrorPaginationInvalidArgument(""),
	)
}
