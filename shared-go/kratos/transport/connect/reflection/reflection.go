// Package reflection provides opt-in gRPC reflection for a Connect server.
package reflection

import (
	"connectrpc.com/grpcreflect"

	"cyber-ecosystem/shared-go/kratos/transport/connect"
)

// Register registers gRPC reflection (v1 + v1alpha) for the services listed by
// srv.RegisteredServices().
func Register(srv *connect.Server) error {
	reflector := grpcreflect.NewStaticReflector(srv.RegisteredServices()...)
	p1, h1 := grpcreflect.NewHandlerV1(reflector)
	srv.Register(p1, h1)
	p2, h2 := grpcreflect.NewHandlerV1Alpha(reflector)
	srv.Register(p2, h2)
	return nil
}
