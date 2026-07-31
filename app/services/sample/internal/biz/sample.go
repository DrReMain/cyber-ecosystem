package biz

import (
	"context"
	"log/slog"
	"time"

	commonpb "cyber-ecosystem/gen/go/cyber/shared/common/v1"
)

// DO ------------------------------------------------------------------------------------------------------------------

type Item struct {
	ID        string
	CreatedAt time.Time
	Name      *string
}

type ItemListIn struct {
	*commonpb.PageRequest
}

type ItemListOut struct {
	*commonpb.PageResponse
	List []*Item
}

// Port ----------------------------------------------------------------------------------------------------------------

type SystemRP interface {
	ListItems(ctx context.Context, in *ItemListIn) (*ItemListOut, error)
}

type SystemConnectRP interface {
	CreateItem(ctx context.Context, item *Item) (*Item, error)
}

// UC ------------------------------------------------------------------------------------------------------------------

type SampleUC struct {
	UC
	systemRP        SystemRP
	systemConnectRP SystemConnectRP
}

func NewSampleUC(logger *slog.Logger, systemRP SystemRP, systemConnectRP SystemConnectRP) *SampleUC {
	return &SampleUC{
		UC:              UC{log: logger.With("module", "biz/sample")},
		systemRP:        systemRP,
		systemConnectRP: systemConnectRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *SampleUC) Create(ctx context.Context, item *Item) (*Item, error) {
	return uc.systemConnectRP.CreateItem(ctx, item)
}

func (uc *SampleUC) List(ctx context.Context, in *ItemListIn) (*ItemListOut, error) {
	return uc.systemRP.ListItems(ctx, in)
}
