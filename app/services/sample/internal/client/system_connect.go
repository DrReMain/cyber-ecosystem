package client

import (
	"context"
	"log/slog"

	connectrpc "connectrpc.com/connect"

	connect "cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	systempb "cyber-ecosystem/gen/go/cyber/system/v1"
	systemv1connect "cyber-ecosystem/gen/go/cyber/system/v1/v1connect"

	"cyber-ecosystem/app/services/sample/internal/biz"
	"cyber-ecosystem/app/services/sample/internal/conf"
)

// Adapter -------------------------------------------------------------------------------------------------------------

type systemConnectClient struct {
	log   *slog.Logger
	items systemv1connect.ItemServiceClient
}

func NewSystemConnectClient(c *conf.Remote, logger *slog.Logger) (biz.SystemConnectRP, func(), error) {
	conn, err := connect.DialInsecure(context.Background(),
		connect.WithEndpoint(c.GetSystemConnect()),
		connect.WithMiddleware(standardMiddleware(logger)...),
	)
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() { _ = conn.Close() }
	return &systemConnectClient{
		log:   logger.With("module", "client/system_connect"),
		items: systemv1connect.NewItemServiceClient(conn.HTTPClient(), conn.BaseURL(), conn.ClientOptions()...),
	}, cleanup, nil
}

// Method --------------------------------------------------------------------------------------------------------------

func (c *systemConnectClient) CreateItem(ctx context.Context, item *biz.Item) (*biz.Item, error) {
	resp, err := c.items.CreateItem(ctx, connectrpc.NewRequest(&systempb.CreateItemRequest{Name: item.Name}))
	if err != nil {
		return nil, err
	}
	id := utils.Unwrap[string](resp.Msg.GetId())
	return &biz.Item{ID: utils.Deref(id, "")}, nil
}
