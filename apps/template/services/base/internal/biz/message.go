package biz

import (
	"time"

	"cyber-ecosystem/contracts/go/common"
	"cyber-ecosystem/shared-go/utils"
)

// region[rgba(239,83,80,0.15)] 🔴 Model -------------------------------------------------------------------------------

type Message struct {
	ID        *string
	CreatedAt *time.Time
	UpdatedAt *time.Time
	Title     *string
	Content   *string
	Status    *string
}

type MessageQueryIn struct {
	*common.PageRequest
	OrderBy []*utils.OrderBy
	ID      *string
	Title   *string
	Status  *string
}

type MessageQueryOut struct {
	*common.PageResponse
	List []*Message
}
