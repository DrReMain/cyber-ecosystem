package service

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"

	pb "cyber-ecosystem/gen/go/cyber/mobile/v1"
	resourcev1 "cyber-ecosystem/gen/go/cyber/resource/v1"

	"cyber-ecosystem/app/services/mobile_bff/internal/biz"
)

// Struct ----------------------------------------------------------------------------------------------------------------

type ResourceService struct {
	pb.UnimplementedMobileResourceServiceServer

	log        *slog.Logger
	resourceUC *biz.ResourceUC
}

func NewResourceService(logger *slog.Logger, resourceUC *biz.ResourceUC) *ResourceService {
	return &ResourceService{
		log:        logger.With("module", "service/resource"),
		resourceUC: resourceUC,
	}
}

func (s *ResourceService) RegisterGRPC(srv *grpc.Server) {
	pb.RegisterMobileResourceServiceServer(srv, s)
}

func (s *ResourceService) RegisterHTTP(srv *http.Server) {
	pb.RegisterMobileResourceServiceHTTPServer(srv, s)
}

func (s *ResourceService) RegisterConnect(srv *connecttransport.Server) {
	pb.RegisterMobileResourceServiceConnectServer(srv, s)
}

// Handler ---------------------------------------------------------------------------------------------------------------

func (s *ResourceService) ListResource(ctx context.Context, _ *resourcev1.ListResourceRequest) (*resourcev1.ListResourceResponse, error) {
	services, err := s.resourceUC.ListResource(ctx)
	if err != nil {
		return nil, err
	}
	return &resourcev1.ListResourceResponse{
		List: toProtoServices(services),
	}, nil
}

// Private ---------------------------------------------------------------------------------------------------------------

func toProtoServices(services []*biz.ResourceService) []*resourcev1.Service {
	out := make([]*resourcev1.Service, len(services))
	for i, svc := range services {
		out[i] = &resourcev1.Service{
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

func toProtoMethods(methods []*biz.ResourceMethod) []*resourcev1.Method {
	out := make([]*resourcev1.Method, len(methods))
	for i, m := range methods {
		out[i] = &resourcev1.Method{
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
