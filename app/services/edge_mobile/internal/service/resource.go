package service

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"

	pb "cyber-ecosystem/gen/go/cyber/mobile/v1"

	"cyber-ecosystem/app/services/edge_mobile/internal/biz"
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

func (s *ResourceService) ListResource(ctx context.Context, _ *pb.ListResourceRequest) (*pb.ListResourceResponse, error) {
	services, err := s.resourceUC.ListResource(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.ListResourceResponse{
		List: toProtoServices(services),
	}, nil
}

// Private ---------------------------------------------------------------------------------------------------------------

func toProtoServices(services []*biz.ResourceService) []*pb.Service {
	out := make([]*pb.Service, len(services))
	for i, svc := range services {
		out[i] = &pb.Service{
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

func toProtoMethods(methods []*biz.ResourceMethod) []*pb.Method {
	out := make([]*pb.Method, len(methods))
	for i, m := range methods {
		out[i] = &pb.Method{
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
