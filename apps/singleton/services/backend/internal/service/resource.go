package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	"cyber-ecosystem/shared-go/kratos/transport/connect"

	singletonV1 "cyber-ecosystem/apps/singleton/gen/go/v1"
	singletonV1connect "cyber-ecosystem/apps/singleton/gen/go/v1/singletonV1connect"
	"cyber-ecosystem/apps/singleton/services/backend/internal/biz"
)

type ResourceService struct {
	singletonV1.UnimplementedResourceServiceServer
	log        *log.Helper
	resourceUC *biz.ResourceUC
}

func NewResourceService(logger log.Logger, resourceUC *biz.ResourceUC) *ResourceService {
	return &ResourceService{
		log:        log.NewHelper(log.With(logger, "module", "service/resource")),
		resourceUC: resourceUC,
	}
}
func (s *ResourceService) RegisterGRPC(srv *grpc.Server) {
	singletonV1.RegisterResourceServiceServer(srv, s)
}
func (s *ResourceService) RegisterHTTP(srv *http.Server) {
	singletonV1.RegisterResourceServiceHTTPServer(srv, s)
}
func (s *ResourceService) RegisterConnect(srv *connect.Server) {
	srv.Register(singletonV1connect.NewResourceServiceHandler(s, srv.HandlerOptions()...))
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *ResourceService) ListResource(ctx context.Context, in *singletonV1.ListResourceRequest) (*singletonV1.ListResourceResponse, error) {
	services, err := s.resourceUC.ListResource(ctx)
	if err != nil {
		return nil, err
	}

	return &singletonV1.ListResourceResponse{
		List: func() []*singletonV1.Service {
			r1 := make([]*singletonV1.Service, 0, len(services))
			for _, s := range services {
				r1 = append(r1, &singletonV1.Service{
					Name:       s.Name,
					FullName:   s.FullName,
					Package:    s.Package,
					SourceFile: s.SourceFile,
					Comment:    s.Comment,
					Methods: func() []*singletonV1.Method {
						r2 := make([]*singletonV1.Method, 0, len(s.Methods))
						for _, m := range s.Methods {
							r2 = append(r2, &singletonV1.Method{
								Name:             m.Name,
								FullName:         m.FullName,
								RequestName:      m.RequestName,
								RequestFullName:  m.RequestFullName,
								ResponseName:     m.ResponseName,
								ResponseFullName: m.ResponseFullName,
								HttpMethod:       m.HttpMethod,
								HttpPath:         m.HttpPath,
								Comment:          m.Comment,
							})
						}
						return r2
					}(),
				})
			}
			return r1
		}(),
	}, nil
}
