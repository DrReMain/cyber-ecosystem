package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	"cyber-ecosystem/shared-go/kratos/transport/connect"

	genesisV1 "cyber-ecosystem/apps/genesis/gen/go/v1"
	"cyber-ecosystem/apps/genesis/services/base/internal/biz"
)

// region[rgba(236,64,122,0.15)] 🩷 Struct -------------------------------------------------------------------------------

type ResourceService struct {
	genesisV1.UnimplementedResourceServiceServer
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
	genesisV1.RegisterResourceServiceServer(srv, s)
}
func (s *ResourceService) RegisterHTTP(_ *http.Server)       {}
func (s *ResourceService) RegisterConnect(_ *connect.Server) {}

// region[rgba(255,167,38,0.15)] 🟠 Handler -----------------------------------------------------------------------------

func (s *ResourceService) ListResource(ctx context.Context, _ *genesisV1.ListResourceRequest) (*genesisV1.ListResourceResponse, error) {
	services, err := s.resourceUC.ListResource(ctx)
	if err != nil {
		return nil, err
	}

	return &genesisV1.ListResourceResponse{
		List: func() []*genesisV1.Service {
			r1 := make([]*genesisV1.Service, 0, len(services))
			for _, svc := range services {
				r1 = append(r1, &genesisV1.Service{
					Name:       svc.Name,
					FullName:   svc.FullName,
					Package:    svc.Package,
					SourceFile: svc.SourceFile,
					Comment:    svc.Comment,
					Methods: func() []*genesisV1.Method {
						r2 := make([]*genesisV1.Method, 0, len(svc.Methods))
						for _, m := range svc.Methods {
							r2 = append(r2, &genesisV1.Method{
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
