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

type ConditionService struct {
	singletonV1.UnimplementedConditionServiceServer
	log    *log.Helper
	condUC *biz.ConditionUC
}

func NewConditionService(logger log.Logger, condUC *biz.ConditionUC) *ConditionService {
	return &ConditionService{
		log:    log.NewHelper(log.With(logger, "module", "service/condition")),
		condUC: condUC,
	}
}

func (s *ConditionService) RegisterGRPC(srv *grpc.Server) {
	singletonV1.RegisterConditionServiceServer(srv, s)
}

func (s *ConditionService) RegisterHTTP(srv *http.Server) {
	singletonV1.RegisterConditionServiceHTTPServer(srv, s)
}

func (s *ConditionService) RegisterConnect(srv *connect.Server) {
	srv.Register(singletonV1connect.NewConditionServiceHandler(s, srv.HandlerOptions()...))
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *ConditionService) CreateCondition(ctx context.Context, in *singletonV1.CreateConditionRequest) (*singletonV1.CreateConditionResponse, error) {
	if err := s.condUC.CreateCondition(ctx, &biz.Condition{
		RoleCode:      in.RoleCode,
		Operation:     in.Operation,
		ConditionType: in.ConditionType,
		Config:        in.Config,
		GroupID:       in.GroupId,
	}); err != nil {
		return nil, err
	}
	return &singletonV1.CreateConditionResponse{}, nil
}

func (s *ConditionService) UpdateCondition(ctx context.Context, in *singletonV1.UpdateConditionRequest) (*singletonV1.UpdateConditionResponse, error) {
	cond := &biz.Condition{
		ID:            &in.Id,
		Operation:     in.Operation,
		ConditionType: in.ConditionType,
		Config:        in.Config,
		GroupID:       in.GroupId,
	}
	if err := s.condUC.UpdateCondition(ctx, in.FieldsMask, cond); err != nil {
		return nil, err
	}
	return &singletonV1.UpdateConditionResponse{}, nil
}

func (s *ConditionService) DeleteCondition(ctx context.Context, in *singletonV1.DeleteConditionRequest) (*singletonV1.DeleteConditionResponse, error) {
	if err := s.condUC.DeleteCondition(ctx, in.Id); err != nil {
		return nil, err
	}
	return &singletonV1.DeleteConditionResponse{}, nil
}

func (s *ConditionService) GetCondition(ctx context.Context, in *singletonV1.GetConditionRequest) (*singletonV1.GetConditionResponse, error) {
	cond, err := s.condUC.GetCondition(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return s.conditionToProto(cond), nil
}

func (s *ConditionService) QueryConditions(ctx context.Context, in *singletonV1.QueryConditionsRequest) (*singletonV1.QueryConditionsResponse, error) {
	out, err := s.condUC.QueryConditions(ctx, &biz.ConditionQueryIn{
		PageRequest:   utils.EnsurePageRequest(in.Page),
		OrderBy:       utils.ParseOrderBy(in.OrderBy),
		RoleCode:      in.RoleCode,
		Operation:     in.Operation,
		ConditionType: in.ConditionType,
		GroupID:       in.GroupId,
	})
	if err != nil {
		return nil, err
	}
	return &singletonV1.QueryConditionsResponse{
		Page: out.PageResponse,
		List: utils.SliceMap(out.List, s.conditionToProto),
	}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func (s *ConditionService) conditionToProto(e *biz.Condition) *singletonV1.GetConditionResponse {
	return &singletonV1.GetConditionResponse{
		Id:            *e.ID,
		CreatedAt:     utils.ToTimestamp(e.CreatedAt),
		UpdatedAt:     utils.ToTimestamp(e.UpdatedAt),
		RoleCode:      utils.Wrap(e.RoleCode, utils.StringW),
		Operation:     utils.Wrap(e.Operation, utils.StringW),
		ConditionType: utils.Wrap(e.ConditionType, utils.StringW),
		Config:        utils.Wrap(e.Config, utils.StringW),
		GroupId:       utils.Wrap(e.GroupID, utils.StringW),
	}
}
