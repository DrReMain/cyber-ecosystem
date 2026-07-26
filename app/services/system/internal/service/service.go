package service

import (
	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"
	"github.com/google/wire"

	"cyber-ecosystem/shared-go/kratos/transport/connect"
)

type Registrar interface {
	RegisterGRPC(*grpc.Server)
	RegisterHTTP(*http.Server)
	RegisterConnect(*connect.Server)
}

var ProviderSet = wire.NewSet(
	NewItemService,
	NewRegistrarList,
	NewResourceService,
)

func NewRegistrarList(
	s1 *ItemService,
	s2 *ResourceService,
) []Registrar {
	return []Registrar{s1, s2}
}
