package service

import (
	"cyber-ecosystem/shared-go/kratos/transport/connect"

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
	NewBlogService,
	NewAuthorService,
)

func NewRegistrarList(
	s1 *BlogService,
	s2 *AuthorService,
) []Registrar {
	return []Registrar{
		s1,
		s2,
	}
}
