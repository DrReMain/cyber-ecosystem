package connect

import (
	"net/http"

	"connectrpc.com/grpcreflect"
)

// ReflectionHandler creates a v1 gRPC reflection handler for Connect.
// This enables tools like grpcurl to discover services.
// The services parameter should be fully-qualified Protobuf service names
// (for example, "acme.user.v1.UserService").
func ReflectionHandler(services ...string) (string, http.Handler) {
	reflector := grpcreflect.NewStaticReflector(services...)
	return grpcreflect.NewHandlerV1(reflector)
}

// ReflectionHandlerV1Alpha creates a v1alpha gRPC reflection handler for Connect.
func ReflectionHandlerV1Alpha(services ...string) (string, http.Handler) {
	reflector := grpcreflect.NewStaticReflector(services...)
	return grpcreflect.NewHandlerV1Alpha(reflector)
}

// ReflectionHandlers returns both v1 and v1alpha handlers.
func ReflectionHandlers(services ...string) []struct {
	Path    string
	Handler http.Handler
} {
	v1Path, v1Handler := ReflectionHandler(services...)
	v1AlphaPath, v1AlphaHandler := ReflectionHandlerV1Alpha(services...)
	return []struct {
		Path    string
		Handler http.Handler
	}{
		{Path: v1Path, Handler: v1Handler},
		{Path: v1AlphaPath, Handler: v1AlphaHandler},
	}
}

// RegisterReflection registers gRPC reflection on the Connect server.
// This enables tools like grpcurl, grpcui, and other gRPC tools to work with Connect services.
// The services parameter should be fully-qualified Protobuf service names.
func RegisterReflection(srv *Server, services ...string) {
	for _, entry := range ReflectionHandlers(services...) {
		srv.Register(entry.Path, entry.Handler)
	}
}
