package server

import (
	"github.com/DrReMain/cyber-ecosystem/kratos/system-service/internal/conf"
	"github.com/DrReMain/cyber-ecosystem/kratos/system-service/internal/service"

	"github.com/DrReMain/cyber-ecosystem/go-shared/kratos/middleware/validate"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

func NewGRPCServer(c *conf.Server, services []service.Registrar, logger log.Logger) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
			validate.ProtoValidate(validate.UseProtoMessage),
		),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)
	for _, svc := range services {
		svc.RegisterGRPC(srv)
	}
	return srv
}
