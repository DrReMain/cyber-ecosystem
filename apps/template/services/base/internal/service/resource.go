package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	"cyber-ecosystem/shared-go/kratos/transport/connect"

	templateV1 "cyber-ecosystem/apps/template/gen/go/v1"
	templateV1connect "cyber-ecosystem/apps/template/gen/go/v1/templateV1connect"
	"cyber-ecosystem/apps/template/services/base/internal/biz"
)

// region[rgba(236,64,122,0.15)] 🩷 Struct -------------------------------------------------------------------------------

type ResourceService struct {
	templateV1.UnimplementedResourceServiceServer
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
	templateV1.RegisterResourceServiceServer(srv, s)
}
func (s *ResourceService) RegisterHTTP(srv *http.Server) {
	templateV1.RegisterResourceServiceHTTPServer(srv, s)
}
func (s *ResourceService) RegisterConnect(srv *connect.Server) {
	srv.Register(templateV1connect.NewResourceServiceHandler(s, srv.HandlerOptions()...))
}

// region[rgba(255,167,38,0.15)] 🟠 Handler -----------------------------------------------------------------------------

func (s *ResourceService) ListResource(ctx context.Context, in *templateV1.ListResourceRequest) (*templateV1.ListResourceResponse, error) {
	services, err := s.resourceUC.ListResource(ctx)
	if err != nil {
		return nil, err
	}

	return &templateV1.ListResourceResponse{
		List: func() []*templateV1.Service {
			r1 := make([]*templateV1.Service, 0, len(services))
			for _, svc := range services {
				r1 = append(r1, &templateV1.Service{
					Name:       svc.Name,
					FullName:   svc.FullName,
					Package:    svc.Package,
					SourceFile: svc.SourceFile,
					Comment:    svc.Comment,
					Methods: func() []*templateV1.Method {
						r2 := make([]*templateV1.Method, 0, len(svc.Methods))
						for _, m := range svc.Methods {
							r2 = append(r2, &templateV1.Method{
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
