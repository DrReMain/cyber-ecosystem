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
	NewRegistrarList,
	NewSampleService,
)

func NewRegistrarList(s1 *SampleService) []Registrar {
	return []Registrar{s1}
}
