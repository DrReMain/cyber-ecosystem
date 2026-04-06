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

type RoleBindingService struct {
	singletonV1.UnimplementedRoleBindingServiceServer
	log              *log.Helper
	policyUC         *biz.PolicyUC
	cacheInvalidator biz.ScopeCacheInvalidator
}

func NewRoleBindingService(logger log.Logger, policyUC *biz.PolicyUC, cacheInvalidator biz.ScopeCacheInvalidator) *RoleBindingService {
	return &RoleBindingService{
		log:              log.NewHelper(log.With(logger, "module", "service/role_binding")),
		policyUC:         policyUC,
		cacheInvalidator: cacheInvalidator,
	}
}
func (s *RoleBindingService) RegisterGRPC(srv *grpc.Server) {
	singletonV1.RegisterRoleBindingServiceServer(srv, s)
}
func (s *RoleBindingService) RegisterHTTP(srv *http.Server) {
	singletonV1.RegisterRoleBindingServiceHTTPServer(srv, s)
}
func (s *RoleBindingService) RegisterConnect(srv *connect.Server) {
	srv.Register(singletonV1connect.NewRoleBindingServiceHandler(s, srv.HandlerOptions()...))
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *RoleBindingService) CreateRoleBinding(ctx context.Context, in *singletonV1.CreateRoleBindingRequest) (*singletonV1.CreateRoleBindingResponse, error) {
	if err := s.policyUC.GrantRole(ctx, in.UserId, in.RoleCode); err != nil {
		return nil, err
	}
	if err := s.cacheInvalidator.InvalidateUser(ctx, in.UserId); err != nil {
		s.log.Warnf("failed to invalidate scope cache for user %s: %v", in.UserId, err)
	}
	return &singletonV1.CreateRoleBindingResponse{}, nil
}

func (s *RoleBindingService) DeleteRoleBinding(ctx context.Context, in *singletonV1.DeleteRoleBindingRequest) (*singletonV1.DeleteRoleBindingResponse, error) {
	if err := s.policyUC.RevokeRole(ctx, in.UserId, in.RoleCode); err != nil {
		return nil, err
	}
	if err := s.cacheInvalidator.InvalidateUser(ctx, in.UserId); err != nil {
		s.log.Warnf("failed to invalidate scope cache for user %s: %v", in.UserId, err)
	}
	return &singletonV1.DeleteRoleBindingResponse{}, nil
}

func (s *RoleBindingService) ListRoleBindings(ctx context.Context, in *singletonV1.ListRoleBindingsRequest) (*singletonV1.ListRoleBindingsResponse, error) {
	roles, err := s.policyUC.QueryUserRoles(ctx, in.UserId)
	if err != nil {
		return nil, err
	}
	return &singletonV1.ListRoleBindingsResponse{
		List: utils.SliceMap(roles, s.roleToProto),
	}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func (s *RoleBindingService) roleToProto(e *biz.Role) *singletonV1.GetRoleResponse {
	return &singletonV1.GetRoleResponse{
		Id:          *e.ID,
		CreatedAt:   utils.ToTimestamp(e.CreatedAt),
		UpdatedAt:   utils.ToTimestamp(e.UpdatedAt),
		Name:        utils.Wrap(e.Name, utils.StringW),
		Code:        utils.Wrap(e.Code, utils.StringW),
		Description: utils.Wrap(e.Description, utils.StringW),
		Status:      utils.Wrap(e.Status, utils.UInt32FromUint8),
		Sort:        *e.Sort,
	}
}
