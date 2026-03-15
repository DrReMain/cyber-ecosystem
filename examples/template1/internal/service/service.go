package service

import (
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	"github.com/google/wire"
)

type Registrar interface {
	RegisterGRPC(*grpc.Server)
	RegisterHTTP(*http.Server)
}

var ProviderSet = wire.NewSet(
	NewRegistrarList,
	NewBlogService,
)

func NewRegistrarList(
	s1 *BlogService,
) []Registrar {
	return []Registrar{
		s1,
	}
}
