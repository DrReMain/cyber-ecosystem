package resource

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"

	systempb "cyber-ecosystem/gen/go/cyber/system/v1"
)

// Struct --------------------------------------------------------------------------------------------------------------

type ResourceService struct {
	systempb.UnimplementedResourceServiceServer

	log        *slog.Logger
	resourceUC *ResourceUC
}

func NewResourceService(logger *slog.Logger, resourceUC *ResourceUC) *ResourceService {
	return &ResourceService{
		log:        logger.With("module", "module/resource_service"),
		resourceUC: resourceUC,
	}
}

func (s *ResourceService) RegisterGRPC(srv *grpc.Server) {
	systempb.RegisterResourceServiceServer(srv, s)
}

func (s *ResourceService) RegisterHTTP(srv *http.Server) {
	systempb.RegisterResourceServiceHTTPServer(srv, s)
}

func (s *ResourceService) RegisterConnect(srv *connecttransport.Server) {
	systempb.RegisterResourceServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *ResourceService) ListResource(ctx context.Context, in *systempb.ListResourceRequest) (*systempb.ListResourceResponse, error) {
	services, err := s.resourceUC.ListResource(ctx)
	if err != nil {
		return nil, err
	}
	return &systempb.ListResourceResponse{
		List: toProtoServices(services),
	}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func toProtoServices(services []*ServiceMeta) []*systempb.Service {
	out := make([]*systempb.Service, len(services))
	for i, svc := range services {
		out[i] = &systempb.Service{
			Name:       svc.Name,
			FullName:   svc.FullName,
			Package:    svc.Package,
			SourceFile: svc.SourceFile,
			Comment:    svc.Comment,
			Methods:    toProtoMethods(svc.Methods),
		}
	}
	return out
}

func toProtoMethods(methods []*ResourceMethod) []*systempb.Method {
	out := make([]*systempb.Method, len(methods))
	for i, m := range methods {
		out[i] = &systempb.Method{
			Name:             m.Name,
			FullName:         m.FullName,
			RequestName:      m.RequestName,
			RequestFullName:  m.RequestFullName,
			ResponseName:     m.ResponseName,
			ResponseFullName: m.ResponseFullName,
			HttpMethod:       m.HttpMethod,
			HttpPath:         m.HttpPath,
			Comment:          m.Comment,
		}
	}
	return out
}
