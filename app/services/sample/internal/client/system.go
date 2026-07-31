package client

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"

	"cyber-ecosystem/shared-go/utils"

	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/sample/internal/biz"
	"cyber-ecosystem/app/services/sample/internal/conf"
)

// Adapter -------------------------------------------------------------------------------------------------------------

type systemClient struct {
	log   *slog.Logger
	items systempb.ItemServiceClient
}

func NewSystemClient(c *conf.Remote, logger *slog.Logger) (biz.SystemRP, func(), error) {
	conn, err := grpc.NewClient(context.Background(),
		grpc.WithEndpoint(c.GetSystem()),
		grpc.WithMiddleware(standardMiddleware(logger)...),
	)
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() { _ = conn.Close() }
	return &systemClient{
		log:   logger.With("module", "client/system"),
		items: systempb.NewItemServiceClient(conn),
	}, cleanup, nil
}

// Method --------------------------------------------------------------------------------------------------------------

func (c *systemClient) ListItems(ctx context.Context, in *biz.ItemListIn) (*biz.ItemListOut, error) {
	resp, err := c.items.ListItems(ctx, &systempb.ListItemsRequest{Page: in.PageRequest})
	if err != nil {
		return nil, err
	}
	return &biz.ItemListOut{
		PageResponse: resp.GetPage(),
		List:         utils.SliceMap(resp.GetList(), mapItem),
	}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mapItem(it *systempb.Item) *biz.Item {
	id := utils.Unwrap[string](it.GetId())
	return &biz.Item{
		ID:        utils.Deref(id, ""),
		CreatedAt: utils.ToTime(it.GetCreatedAt()),
		Name:      utils.Unwrap[string](it.GetName()),
	}
}
