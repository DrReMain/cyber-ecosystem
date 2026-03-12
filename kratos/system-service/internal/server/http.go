package server

import (
	"github.com/DrReMain/cyber-ecosystem/kratos/system-service/internal/conf"
	"github.com/DrReMain/cyber-ecosystem/kratos/system-service/internal/service"

	"github.com/DrReMain/cyber-ecosystem/go-shared/kratos/encoder"
	"github.com/DrReMain/cyber-ecosystem/go-shared/kratos/middleware/validate"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

func NewHTTPServer(c *conf.Server, services []service.Registrar, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			validate.ProtoValidate(validate.UseProtoMessage),
		),
		http.ResponseEncoder(encoder.ResponseEncoder),
		http.ErrorEncoder(encoder.ErrorEncoder),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	for _, svc := range services {
		svc.RegisterHTTP(srv)
	}
	return srv
}
