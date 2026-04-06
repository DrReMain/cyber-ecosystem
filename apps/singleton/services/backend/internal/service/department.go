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

type DepartmentService struct {
	singletonV1.UnimplementedDepartmentServiceServer
	log    *log.Helper
	deptUC *biz.DepartmentUC
}

func NewDepartmentService(logger log.Logger, deptUC *biz.DepartmentUC) *DepartmentService {
	return &DepartmentService{
		log:    log.NewHelper(log.With(logger, "module", "service/department")),
		deptUC: deptUC,
	}
}
func (s *DepartmentService) RegisterGRPC(srv *grpc.Server) {
	singletonV1.RegisterDepartmentServiceServer(srv, s)
}
func (s *DepartmentService) RegisterHTTP(srv *http.Server) {
	singletonV1.RegisterDepartmentServiceHTTPServer(srv, s)
}
func (s *DepartmentService) RegisterConnect(srv *connect.Server) {
	srv.Register(singletonV1connect.NewDepartmentServiceHandler(s, srv.HandlerOptions()...))
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *DepartmentService) CreateDepartment(ctx context.Context, in *singletonV1.CreateDepartmentRequest) (*singletonV1.CreateDepartmentResponse, error) {
	dept := &biz.Department{
		Name:     in.Name,
		Code:     in.Code,
		ParentID: in.ParentId,
	}
	if err := s.deptUC.Create(ctx, dept); err != nil {
		return nil, err
	}
	return &singletonV1.CreateDepartmentResponse{}, nil
}

func (s *DepartmentService) UpdateDepartment(ctx context.Context, in *singletonV1.UpdateDepartmentRequest) (*singletonV1.UpdateDepartmentResponse, error) {
	dept := &biz.Department{
		ID:     &in.Id,
		Name:   in.Name,
		Code:   in.Code,
		Status: utils.ConvPtr[uint8](in.Status),
	}
	if err := s.deptUC.Update(ctx, in.FieldsMask, dept); err != nil {
		return nil, err
	}
	return &singletonV1.UpdateDepartmentResponse{}, nil
}

func (s *DepartmentService) DeleteDepartment(ctx context.Context, in *singletonV1.DeleteDepartmentRequest) (*singletonV1.DeleteDepartmentResponse, error) {
	if err := s.deptUC.Delete(ctx, in.Id); err != nil {
		return nil, err
	}
	return &singletonV1.DeleteDepartmentResponse{}, nil
}

func (s *DepartmentService) GetDepartment(ctx context.Context, in *singletonV1.GetDepartmentRequest) (*singletonV1.GetDepartmentResponse, error) {
	dept, err := s.deptUC.Get(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return s.deptToProto(dept), nil
}

func (s *DepartmentService) QueryDepartment(ctx context.Context, in *singletonV1.QueryDepartmentRequest) (*singletonV1.QueryDepartmentResponse, error) {
	out, err := s.deptUC.Query(ctx, &biz.DepartmentQueryIn{
		PageRequest: utils.EnsurePageRequest(in.Page),
		OrderBy:     utils.ParseOrderBy(in.OrderBy),
		Name:        in.Name,
		Code:        in.Code,
		ParentID:    in.ParentId,
	})
	if err != nil {
		return nil, err
	}
	return &singletonV1.QueryDepartmentResponse{
		Page: out.PageResponse,
		List: utils.SliceMap(out.List, s.deptToProto),
	}, nil
}

func (s *DepartmentService) MoveDepartment(ctx context.Context, in *singletonV1.MoveDepartmentRequest) (*singletonV1.MoveDepartmentResponse, error) {
	if err := s.deptUC.Move(ctx, in.Id, in.ParentId); err != nil {
		return nil, err
	}
	return &singletonV1.MoveDepartmentResponse{}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func (s *DepartmentService) deptToProto(e *biz.Department) *singletonV1.GetDepartmentResponse {
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
