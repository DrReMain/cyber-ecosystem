package biz

import (
	"time"

	"cyber-ecosystem/contracts/go/common"
	"cyber-ecosystem/shared-go/utils"
)

// region[rgba(239,83,80,0.15)] 🔴 Model -------------------------------------------------------------------------------

type Article struct {
	ID        *string
	CreatedAt *time.Time
	UpdatedAt *time.Time
	Title     *string
	Content   *string
	Status    *string
}

type ArticleQueryIn struct {
	*common.PageRequest
	OrderBy []*utils.OrderBy
	ID      *string
	Title   *string
	Status  *string
}

type ArticleQueryOut struct {
	*common.PageResponse
	List []*Article
}
