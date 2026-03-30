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

// ProviderSet aggregates all service providers.
// When adding a new service: add its constructor here AND add it to NewRegistrarList.
var ProviderSet = wire.NewSet(
	NewRegistrarList,
	NewBlogService,
	NewAuthorService,
)

// NewRegistrarList collects all services that need to register with transport servers.
// Add new services as parameters here and append them to the returned slice.
func NewRegistrarList(
	blog *BlogService,
	author *AuthorService,
) []Registrar {
	return []Registrar{
		blog,
		author,
	}
}
