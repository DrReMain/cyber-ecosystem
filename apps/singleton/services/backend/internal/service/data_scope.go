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

type DataScopeService struct {
	singletonV1.UnimplementedDataScopeServiceServer
	log         *log.Helper
	dataScopeUC *biz.DataScopeUC
}

func NewDataScopeService(logger log.Logger, dataScopeUC *biz.DataScopeUC) *DataScopeService {
	return &DataScopeService{
		log:         log.NewHelper(log.With(logger, "module", "service/data_scope")),
		dataScopeUC: dataScopeUC,
	}
}
func (s *DataScopeService) RegisterGRPC(srv *grpc.Server) {
	singletonV1.RegisterDataScopeServiceServer(srv, s)
}
func (s *DataScopeService) RegisterHTTP(srv *http.Server) {
	singletonV1.RegisterDataScopeServiceHTTPServer(srv, s)
}
func (s *DataScopeService) RegisterConnect(srv *connect.Server) {
	srv.Register(singletonV1connect.NewDataScopeServiceHandler(s, srv.HandlerOptions()...))
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *DataScopeService) CreateDataScope(ctx context.Context, in *singletonV1.CreateDataScopeRequest) (*singletonV1.CreateDataScopeResponse, error) {
	if err := s.dataScopeUC.CreateDataScope(ctx, &biz.DataScope{
		RoleCode:       in.RoleCode,
		ScopeType:      in.ScopeType,
		ScopeConfig:    in.ScopeConfig,
		TargetResource: in.TargetResource,
	}); err != nil {
		return nil, err
	}
	return &singletonV1.CreateDataScopeResponse{}, nil
}

func (s *DataScopeService) UpdateDataScope(ctx context.Context, in *singletonV1.UpdateDataScopeRequest) (*singletonV1.UpdateDataScopeResponse, error) {
	scope := &biz.DataScope{
		ID:             &in.Id,
		ScopeType:      in.ScopeType,
		ScopeConfig:    in.ScopeConfig,
		TargetResource: in.TargetResource,
	}
	if err := s.dataScopeUC.UpdateDataScope(ctx, in.FieldsMask, scope); err != nil {
		return nil, err
	}
	return &singletonV1.UpdateDataScopeResponse{}, nil
}

func (s *DataScopeService) DeleteDataScope(ctx context.Context, in *singletonV1.DeleteDataScopeRequest) (*singletonV1.DeleteDataScopeResponse, error) {
	if err := s.dataScopeUC.DeleteDataScope(ctx, in.Id); err != nil {
		return nil, err
	}
	return &singletonV1.DeleteDataScopeResponse{}, nil
}

func (s *DataScopeService) GetDataScope(ctx context.Context, in *singletonV1.GetDataScopeRequest) (*singletonV1.GetDataScopeResponse, error) {
	scope, err := s.dataScopeUC.GetDataScope(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return s.dataScopeToProto(scope), nil
}

func (s *DataScopeService) QueryDataScopes(ctx context.Context, in *singletonV1.QueryDataScopesRequest) (*singletonV1.QueryDataScopesResponse, error) {
	out, err := s.dataScopeUC.QueryDataScopes(ctx, &biz.DataScopeQueryIn{
		PageRequest:    utils.EnsurePageRequest(in.Page),
		OrderBy:        utils.ParseOrderBy(in.OrderBy),
		RoleCode:       in.RoleCode,
		ScopeType:      in.ScopeType,
		TargetResource: in.TargetResource,
	})
	if err != nil {
		return nil, err
	}
	return &singletonV1.QueryDataScopesResponse{
		Page: out.PageResponse,
		List: utils.SliceMap(out.List, s.dataScopeToProto),
	}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func (s *DataScopeService) dataScopeToProto(e *biz.DataScope) *singletonV1.GetDataScopeResponse {
	return &singletonV1.GetDataScopeResponse{
		Id:             *e.ID,
		CreatedAt:      utils.ToTimestamp(e.CreatedAt),
		UpdatedAt:      utils.ToTimestamp(e.UpdatedAt),
		RoleCode:       utils.Wrap(e.RoleCode, utils.StringW),
		ScopeType:      utils.Wrap(e.ScopeType, utils.StringW),
		ScopeConfig:    utils.Wrap(e.ScopeConfig, utils.StringW),
		TargetResource: utils.Wrap(e.TargetResource, utils.StringW),
	}
}
