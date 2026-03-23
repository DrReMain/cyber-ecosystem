package service

import (
	"context"

	"github.com/DrReMain/cyber-ecosystem/examples/template3/internal/biz"

	template3V1 "github.com/DrReMain/cyber-ecosystem/gen/go/template3/v1"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
)

type HelloService struct {
	template3V1.UnimplementedHelloServiceServer

	log     *log.Helper
	helloUC *biz.HelloUC
}

func NewHelloService(logger log.Logger, helloUC *biz.HelloUC) *HelloService {
	return &HelloService{
		log:     log.NewHelper(log.With(logger, "module", "service/hello")),
		helloUC: helloUC,
	}
}
func (s *HelloService) RegisterGRPC(srv *grpc.Server) {
	template3V1.RegisterHelloServiceServer(srv, s)
}
func (s *HelloService) RegisterHTTP(srv *http.Server) {
	template3V1.RegisterHelloServiceHTTPServer(srv, s)
}

// ---------------------------------------------------------------------------------------------------------------------

func (s *HelloService) SayHello(ctx context.Context, in *template3V1.SayHelloRequest) (*template3V1.SayHelloResponse, error) {
	entity := &biz.HelloEntity{
		Name: in.Name,
	}
	e, err := s.helloUC.SayHello(ctx, entity)
	if err != nil {
		return nil, err
	}
	return &template3V1.SayHelloResponse{
		Message: "Hello, " + *e.Name + "!",
	}, nil
}
