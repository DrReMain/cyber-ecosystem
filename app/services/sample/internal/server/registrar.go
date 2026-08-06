package server

import (
	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	"cyber-ecosystem/shared-go/kratos/transport/connect"

	"cyber-ecosystem/app/services/sample/internal/module/sample"
)

// Registrar is implemented by every domain service (duck-typed via
// RegisterGRPC/HTTP/Connect) to register its RPC handlers on all transports.
type Registrar interface {
	RegisterGRPC(*grpc.Server)
	RegisterHTTP(*http.Server)
	RegisterConnect(*connect.Server)
}

// NewRegistrarList aggregates every domain service into the []Registrar the
// transport servers loop over.
func NewRegistrarList(s1 *sample.SampleService) []Registrar {
	return []Registrar{s1}
}
