package service

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	"cyber-ecosystem/shared-go/helper"
	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/biz"
)

// Struct --------------------------------------------------------------------------------------------------------------

type ItemService struct {
	systempb.UnimplementedItemServiceServer

	log    *slog.Logger
	itemUC *biz.ItemUC
}

func NewItemService(logger *slog.Logger, itemUC *biz.ItemUC) *ItemService {
	return &ItemService{
		log:    logger.With("module", "service/item"),
		itemUC: itemUC,
	}
}

func (s *ItemService) RegisterGRPC(srv *grpc.Server) {
	systempb.RegisterItemServiceServer(srv, s)
}

func (s *ItemService) RegisterHTTP(srv *http.Server) {
	systempb.RegisterItemServiceHTTPServer(srv, s)
}

func (s *ItemService) RegisterConnect(srv *connecttransport.Server) {
	systempb.RegisterItemServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *ItemService) CreateItem(ctx context.Context, in *systempb.CreateItemRequest) (*systempb.CreateItemResponse, error) {
	created, err := s.itemUC.Create(ctx, &biz.Item{
		Name:        in.Name,
		Description: in.Description,
	})
	if err != nil {
		return nil, err
	}
	return &systempb.CreateItemResponse{
		Id: utils.StringW(created.ID),
	}, nil
}

func (s *ItemService) UpdateItem(ctx context.Context, in *systempb.UpdateItemRequest) (*systempb.UpdateItemResponse, error) {
	if _, err := s.itemUC.Update(ctx, in.FieldsMask, &biz.Item{
		ID:          *in.Id,
		Name:        in.Name,
		Description: in.Description,
	}); err != nil {
		return nil, err
	}
	return &systempb.UpdateItemResponse{}, nil
}

func (s *ItemService) UpdateItemStatus(ctx context.Context, in *systempb.UpdateItemStatusRequest) (*systempb.UpdateItemStatusResponse, error) {
	if _, err := s.itemUC.UpdateStatus(ctx, *in.Id, *in.Status); err != nil {
		return nil, err
	}
	return &systempb.UpdateItemStatusResponse{}, nil
}

func (s *ItemService) DeleteItem(ctx context.Context, in *systempb.DeleteItemRequest) (*systempb.DeleteItemResponse, error) {
	if _, err := s.itemUC.Delete(ctx, *in.Id); err != nil {
		return nil, err
	}
	return &systempb.DeleteItemResponse{}, nil
}

func (s *ItemService) ListItems(ctx context.Context, in *systempb.ListItemsRequest) (*systempb.ListItemsResponse, error) {
	out, err := s.itemUC.List(ctx, &biz.ItemListIn{
		PageRequest: helper.EnsurePageRequest(in.Page),
		OrderBy:     in.OrderBy,
		Name:        in.Name,
		Status:      in.Status,
	})
	if err != nil {
		return nil, err
	}
	return &systempb.ListItemsResponse{
		Page: out.PageResponse,
		List: utils.SliceMap(out.List, s.itemToProto),
	}, nil
}

func (s *ItemService) GetItem(ctx context.Context, in *systempb.GetItemRequest) (*systempb.GetItemResponse, error) {
	a, err := s.itemUC.Get(ctx, *in.Id)
	if err != nil {
		return nil, err
	}
	return &systempb.GetItemResponse{
		Item: s.itemToProto(a),
	}, nil
}

func (s *ItemService) SortItem(ctx context.Context, in *systempb.SortItemRequest) (*systempb.SortItemResponse, error) {
	if _, err := s.itemUC.Sort(ctx, *in.Id, in.PrevId, in.NextId); err != nil {
		return nil, err
	}
	return &systempb.SortItemResponse{}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func (s *ItemService) itemToProto(a *biz.Item) *systempb.Item {
	return &systempb.Item{
		Id:          utils.StringW(a.ID),
		CreatedAt:   utils.ToTimestamp(&a.CreatedAt),
		UpdatedAt:   utils.ToTimestamp(&a.UpdatedAt),
		Name:        utils.Wrap(a.Name, utils.StringW),
		Description: utils.Wrap(a.Description, utils.StringW),
		Status:      utils.Wrap(a.Status, utils.StringW),
	}
}
