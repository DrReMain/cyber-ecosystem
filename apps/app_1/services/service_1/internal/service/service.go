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
	NewBlogService,
	NewAuthorService,
)

func NewRegistrarList(
	blog *BlogService,
	author *AuthorService,
) []Registrar {
	return []Registrar{
		blog,
		author,
	}
}
