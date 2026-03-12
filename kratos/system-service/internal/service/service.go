package service

import (
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	NewUserService,
)

type Registrar interface {
	RegisterGRPC(*grpc.Server)
	RegisterHTTP(*http.Server)
}
