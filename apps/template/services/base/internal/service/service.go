package service

import (
	"github.com/google/wire"

	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	"cyber-ecosystem/shared-go/kratos/transport/connect"
)

type Registrar interface {
	RegisterGRPC(*grpc.Server)
	RegisterHTTP(*http.Server)
	RegisterConnect(*connect.Server)
}

var ProviderSet = wire.NewSet(
	NewRegistrarList,
	NewResourceService,
	NewMessageService,
)

func NewRegistrarList(
	s1 *ResourceService,
	s2 *MessageService,
) []Registrar {
	return []Registrar{
		s1,
		s2,
	}
}
