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

type RoleService struct {
	singletonV1.UnimplementedRoleServiceServer
	log    *log.Helper
	roleUC *biz.RoleUC
}

func NewRoleService(logger log.Logger, roleUC *biz.RoleUC) *RoleService {
	return &RoleService{
		log:    log.NewHelper(log.With(logger, "module", "service/role")),
		roleUC: roleUC,
	}
}
func (s *RoleService) RegisterGRPC(srv *grpc.Server) {
	singletonV1.RegisterRoleServiceServer(srv, s)
}
func (s *RoleService) RegisterHTTP(srv *http.Server) {
	singletonV1.RegisterRoleServiceHTTPServer(srv, s)
}
func (s *RoleService) RegisterConnect(srv *connect.Server) {
	srv.Register(singletonV1connect.NewRoleServiceHandler(s, srv.HandlerOptions()...))
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *RoleService) CreateRole(ctx context.Context, in *singletonV1.CreateRoleRequest) (*singletonV1.CreateRoleResponse, error) {
	if err := s.roleUC.Create(ctx, &biz.Role{
		Name:        in.Name,
		Code:        in.Code,
		Description: in.Description,
	}); err != nil {
		return nil, err
	}
	return &singletonV1.CreateRoleResponse{}, nil
}

func (s *RoleService) UpdateRole(ctx context.Context, in *singletonV1.UpdateRoleRequest) (*singletonV1.UpdateRoleResponse, error) {
	role := &biz.Role{
		ID:          &in.Id,
		Name:        in.Name,
		Description: in.Description,
		Status:      utils.ConvPtr[uint8](in.Status),
	}
	if err := s.roleUC.Update(ctx, in.FieldsMask, role); err != nil {
		return nil, err
	}
	return &singletonV1.UpdateRoleResponse{}, nil
}

func (s *RoleService) DeleteRole(ctx context.Context, in *singletonV1.DeleteRoleRequest) (*singletonV1.DeleteRoleResponse, error) {
	if err := s.roleUC.Delete(ctx, in.Id); err != nil {
		return nil, err
	}
	return &singletonV1.DeleteRoleResponse{}, nil
}

func (s *RoleService) GetRole(ctx context.Context, in *singletonV1.GetRoleRequest) (*singletonV1.GetRoleResponse, error) {
	role, err := s.roleUC.Get(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return s.roleToProto(role), nil
}

func (s *RoleService) QueryRole(ctx context.Context, in *singletonV1.QueryRoleRequest) (*singletonV1.QueryRoleResponse, error) {
	out, err := s.roleUC.Query(ctx, &biz.RoleQueryIn{
		PageRequest: utils.EnsurePageRequest(in.Page),
		OrderBy:     utils.ParseOrderBy(in.OrderBy),
		Name:        in.Name,
		Code:        in.Code,
	})
	if err != nil {
		return nil, err
	}
	return &singletonV1.QueryRoleResponse{
		Page: out.PageResponse,
		List: utils.SliceMap(out.List, s.roleToProto),
	}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func (s *RoleService) roleToProto(e *biz.Role) *singletonV1.GetRoleResponse {
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
