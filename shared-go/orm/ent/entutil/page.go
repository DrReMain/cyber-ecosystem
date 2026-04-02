package entutil

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"

	"github.com/go-kratos/kratos/v2/errors"

	"cyber-ecosystem/contracts/go/common"
)

var (
	DefaultPageSize        = 10
	DefaultPageSizeMax     = 100
	DefaultPageSizeUnlimit = 0
)

type PageConfig struct {
	DefaultSize int
	MaxSize     int // if 0, unlimit
}

func NewPageConfig(defaultSize, maxSize int) PageConfig {
	return PageConfig{DefaultSize: defaultSize, MaxSize: maxSize}
}

type QueryPaginator[Q any] interface {
	Offset(int) Q
	Limit(int) Q
	Clone() Q
	Count(context.Context) (int, error)
}

func ApplyPagination[Q QueryPaginator[Q]](ctx context.Context, query Q, req *common.PageRequest, config PageConfig, ce *errors.Error) (total, offset, limit int, err error) {
	if ce == nil {
		return 0, 0, 0, fmt.Errorf("badRequestReason is required")
	}
	if req == nil {
		return 0, 0, 0, ce
	}
	reqPageNo := req.GetPageNo()
	reqPageSize := req.GetPageSize()

	pageNo := 1
	if reqPageNo > 0 {
		pageNo = int(reqPageNo)
	}
	pageSize := config.DefaultSize
	if reqPageSize > 0 {
		pageSize = int(reqPageSize)
	}

	if req.GetAll() {
		if config.MaxSize > 0 {
			return 0, 0, 0, ce
		}
		// Count may trigger query interceptors that append predicates in-place.
		// Use a clone so the original query builder can still be reused safely.
		total, err = query.Clone().Count(ctx)
		return total, 0, 0, err
	}

	if config.MaxSize > 0 && pageSize > config.MaxSize {
		pageSize = config.MaxSize
	}

	offset = (pageNo - 1) * pageSize
	limit = pageSize

	// Count may mutate the query builder through interceptors.
	// Clone before counting to avoid duplicating predicates on the main query.
	total, err = query.Clone().Count(ctx)
	if err != nil {
		return 0, 0, 0, err
	}

	query.Offset(offset).Limit(limit)
	return total, offset, limit, nil
}

// ---------------------------------------------------------------------------------------------------------------------

func WherePtr[T any, Q interface{ Where(...P) Q }, P any](query Q, ptr *T, fn func(T) P) {
	if ptr != nil {
		query.Where(fn(*ptr))
	}
}

func Where[Q interface{ Where(...P) Q }, P any](query Q, condition bool, fn func() P) {
	if condition {
		query.Where(fn())
	}
}

// ---------------------------------------------------------------------------------------------------------------------

type SQLSelector func(string) func(*sql.Selector)

type FOMapping map[string]func(SQLSelector)

func ApplyOrderBy[O interface {
	FieldString() string
	OrderString() string
	ASC() string
	DESC() string
}](ob []O, ascFunc, descFunc func(...string) func(*sql.Selector), mapping FOMapping) {
	if len(ob) > 0 {
		for _, rule := range ob {
			if action, ok := mapping[rule.FieldString()]; ok {
				if rule.OrderString() == rule.ASC() {
					action(func(s string) func(*sql.Selector) {
						return ascFunc(s)
					})
				}
				if rule.OrderString() == rule.DESC() {
					action(func(s string) func(*sql.Selector) {
						return descFunc(s)
					})
				}
			}
		}
	}
}

// ---------------------------------------------------------------------------------------------------------------------

func BuildPageResponse(total, offset, limit int) *common.PageResponse {
	if limit == 0 {
		return &common.PageResponse{
			PageNo:   1,
			PageSize: int32(total),
			Total:    int32(total),
			More:     false,
		}
	}
	return &common.PageResponse{
		PageNo:   int32(offset/limit + 1),
		PageSize: int32(limit),
		Total:    int32(total),
		More:     total > offset+limit,
	}
}
