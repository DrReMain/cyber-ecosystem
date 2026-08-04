package service

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	"cyber-ecosystem/shared-go/helper"
	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/biz"
)

// Struct --------------------------------------------------------------------------------------------------------------

type DeptService struct {
	systempb.UnimplementedDeptServiceServer

	log    *slog.Logger
	deptUC *biz.DeptUC
}

func NewDeptService(logger *slog.Logger, deptUC *biz.DeptUC) *DeptService {
	return &DeptService{
		log:    logger.With("module", "service/dept"),
		deptUC: deptUC,
	}
}

func (s *DeptService) RegisterGRPC(srv *grpc.Server) {
	systempb.RegisterDeptServiceServer(srv, s)
}

func (s *DeptService) RegisterHTTP(srv *http.Server) {
	systempb.RegisterDeptServiceHTTPServer(srv, s)
}

func (s *DeptService) RegisterConnect(srv *connecttransport.Server) {
	systempb.RegisterDeptServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *DeptService) CreateDept(ctx context.Context, in *systempb.CreateDeptRequest) (*systempb.CreateDeptResponse, error) {
	created, err := s.deptUC.Create(ctx, &biz.Dept{
		Name: in.Name,
	})
	if err != nil {
		return nil, err
	}
	return &systempb.CreateDeptResponse{
		Id: utils.StringW(created.ID),
	}, nil
}

func (s *DeptService) UpdateDept(ctx context.Context, in *systempb.UpdateDeptRequest) (*systempb.UpdateDeptResponse, error) {
	if _, err := s.deptUC.Update(ctx, in.FieldsMask, &biz.Dept{
		ID:   in.Id,
		Name: in.Name,
	}); err != nil {
		return nil, err
	}
	return &systempb.UpdateDeptResponse{}, nil
}

func (s *DeptService) DeleteDept(ctx context.Context, in *systempb.DeleteDeptRequest) (*systempb.DeleteDeptResponse, error) {
	if _, err := s.deptUC.Delete(ctx, in.Id); err != nil {
		return nil, err
	}
	return &systempb.DeleteDeptResponse{}, nil
}

func (s *DeptService) ListDepts(ctx context.Context, in *systempb.ListDeptsRequest) (*systempb.ListDeptsResponse, error) {
	out, err := s.deptUC.List(ctx, &biz.DeptListIn{
		PageRequest: helper.EnsurePageRequest(in.Page),
		OrderBy:     in.OrderBy,
		Name:        in.Name,
	})
	if err != nil {
		return nil, err
	}
	return &systempb.ListDeptsResponse{
		Page: out.PageResponse,
		List: utils.SliceMap(out.List, s.deptToProto),
	}, nil
}

func (s *DeptService) GetDept(ctx context.Context, in *systempb.GetDeptRequest) (*systempb.GetDeptResponse, error) {
	d, err := s.deptUC.Get(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return &systempb.GetDeptResponse{
		Dept: s.deptToProto(d),
	}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func (s *DeptService) deptToProto(d *biz.Dept) *systempb.Dept {
	return &systempb.Dept{
		Id:        utils.StringW(d.ID),
		CreatedAt: utils.ToTimestamp(&d.CreatedAt),
		UpdatedAt: utils.ToTimestamp(&d.UpdatedAt),
		Name:      utils.Wrap(d.Name, utils.StringW),
	}
}
