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

type DepartmentBindingService struct {
	singletonV1.UnimplementedDepartmentBindingServiceServer
	log           *log.Helper
	deptBindingUC *biz.DepartmentBindingUC
}

func NewDepartmentBindingService(logger log.Logger, deptBindingUC *biz.DepartmentBindingUC) *DepartmentBindingService {
	return &DepartmentBindingService{
		log:           log.NewHelper(log.With(logger, "module", "service/department_binding")),
		deptBindingUC: deptBindingUC,
	}
}
func (s *DepartmentBindingService) RegisterGRPC(srv *grpc.Server) {
	singletonV1.RegisterDepartmentBindingServiceServer(srv, s)
}
func (s *DepartmentBindingService) RegisterHTTP(srv *http.Server) {
	singletonV1.RegisterDepartmentBindingServiceHTTPServer(srv, s)
}
func (s *DepartmentBindingService) RegisterConnect(srv *connect.Server) {
	srv.Register(singletonV1connect.NewDepartmentBindingServiceHandler(s, srv.HandlerOptions()...))
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *DepartmentBindingService) SetUserDepartments(ctx context.Context, in *singletonV1.SetUserDepartmentsRequest) (*singletonV1.SetUserDepartmentsResponse, error) {
	if err := s.deptBindingUC.SetUserDepartments(ctx, in.UserId, in.DepartmentIds); err != nil {
		return nil, err
	}
	return &singletonV1.SetUserDepartmentsResponse{}, nil
}

func (s *DepartmentBindingService) ListUserDepartments(ctx context.Context, in *singletonV1.ListUserDepartmentsRequest) (*singletonV1.ListUserDepartmentsResponse, error) {
	depts, err := s.deptBindingUC.ListUserDepartments(ctx, in.UserId)
	if err != nil {
		return nil, err
	}
	return &singletonV1.ListUserDepartmentsResponse{
		List: utils.SliceMap(depts, s.deptToProto),
	}, nil
}

func (s *DepartmentBindingService) ListDepartmentUsers(ctx context.Context, in *singletonV1.ListDepartmentUsersRequest) (*singletonV1.ListDepartmentUsersResponse, error) {
	userIDs, err := s.deptBindingUC.ListDepartmentUsers(ctx, in.DepartmentId)
	if err != nil {
		return nil, err
	}
	return &singletonV1.ListDepartmentUsersResponse{
		UserIds: userIDs,
	}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func (s *DepartmentBindingService) deptToProto(e *biz.Department) *singletonV1.GetDepartmentResponse {
	return &singletonV1.GetDepartmentResponse{
		Id:        *e.ID,
		CreatedAt: utils.ToTimestamp(e.CreatedAt),
		UpdatedAt: utils.ToTimestamp(e.UpdatedAt),
		Name:      utils.Wrap(e.Name, utils.StringW),
		Code:      utils.Wrap(e.Code, utils.StringW),
		ParentId:  utils.Wrap(e.ParentID, utils.StringW),
		Path:      utils.Wrap(e.Path, utils.StringW),
		Status:    utils.Wrap(e.Status, utils.UInt32FromUint8),
		Sort:      *e.Sort,
	}
}
