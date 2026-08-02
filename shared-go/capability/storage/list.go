package storage

import "context"

// ListResult is one page of a listing. NextCursor empty means no more pages.
type ListResult struct {
	Objects    []ObjectInfo
	NextCursor string
}

// List enumerates objects under a prefix, page by page via cursor.
type List interface {
	List(ctx context.Context, prefix, cursor string, maxKeys int32) (*ListResult, error)
}
