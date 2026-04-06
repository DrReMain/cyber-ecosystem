package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	"cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	singletonV1 "cyber-ecosystem/apps/singleton/gen/go/v1"
	singletonV1connect "cyber-ecosystem/apps/singleton/gen/go/v1/singletonV1connect"
	"cyber-ecosystem/apps/singleton/services/backend/internal/biz"
)

type PermissionBindingService struct {
	singletonV1.UnimplementedPermissionBindingServiceServer
	log      *log.Helper
	policyUC *biz.PolicyUC
}

func NewPermissionBindingService(logger log.Logger, policyUC *biz.PolicyUC) *PermissionBindingService {
	return &PermissionBindingService{
		log:      log.NewHelper(log.With(logger, "module", "service/permission_binding")),
		policyUC: policyUC,
	}
}
func (s *PermissionBindingService) RegisterGRPC(srv *grpc.Server) {
	singletonV1.RegisterPermissionBindingServiceServer(srv, s)
}
func (s *PermissionBindingService) RegisterHTTP(srv *http.Server) {
	singletonV1.RegisterPermissionBindingServiceHTTPServer(srv, s)
}
func (s *PermissionBindingService) RegisterConnect(srv *connect.Server) {
	srv.Register(singletonV1connect.NewPermissionBindingServiceHandler(s, srv.HandlerOptions()...))
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *PermissionBindingService) CreatePermissionBinding(ctx context.Context, in *singletonV1.CreatePermissionBindingRequest) (*singletonV1.CreatePermissionBindingResponse, error) {
	effect := utils.Deref(in.Effect, "allow")
	if err := s.policyUC.AssignPermission(ctx, in.RoleCode, in.Object, effect); err != nil {
		return nil, err
	}
	return &singletonV1.CreatePermissionBindingResponse{}, nil
}

func (s *PermissionBindingService) DeletePermissionBinding(ctx context.Context, in *singletonV1.DeletePermissionBindingRequest) (*singletonV1.DeletePermissionBindingResponse, error) {
	effect := utils.Deref(in.Effect, "allow")
	if err := s.policyUC.RemovePermission(ctx, in.RoleCode, in.Object, effect); err != nil {
		return nil, err
	}
	return &singletonV1.DeletePermissionBindingResponse{}, nil
}

func (s *PermissionBindingService) ListPermissionBindings(ctx context.Context, in *singletonV1.ListPermissionBindingsRequest) (*singletonV1.ListPermissionBindingsResponse, error) {
	permissions, err := s.policyUC.QueryRolePermissions(ctx, in.RoleCode)
	if err != nil {
		return nil, err
	}
	return &singletonV1.ListPermissionBindingsResponse{
		List: utils.SliceMap(permissions, func(p *biz.PermissionBinding) *singletonV1.PermissionBinding {
			return &singletonV1.PermissionBinding{Object: p.Object, Effect: p.Effect}
		}),
	}, nil
}
