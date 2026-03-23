package service

import (
	"github.com/DrReMain/cyber-ecosystem/shared-go/kratos/transport/connect"

	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	"github.com/google/wire"
)

type Registrar interface {
	RegisterGRPC(*grpc.Server)
	RegisterHTTP(*http.Server)
	RegisterConnect(*connect.Server)
}

var ProviderSet = wire.NewSet(
	NewRegistrarList,
	NewReadingService,
)

func NewRegistrarList(
	s1 *ReadingService,
) []Registrar {
	return []Registrar{
		s1,
	}
}
