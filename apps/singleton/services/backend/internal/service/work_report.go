package service

import (
	"context"

	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	"cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	singletonV1 "cyber-ecosystem/apps/singleton/gen/go/v1"
	singletonV1connect "cyber-ecosystem/apps/singleton/gen/go/v1/singletonV1connect"
	"cyber-ecosystem/apps/singleton/services/backend/internal/biz"
)

type WorkReportService struct {
	singletonV1.UnimplementedWorkReportServiceServer
	log          *log.Helper
	workReportUC *biz.WorkReportUC
}

func NewWorkReportService(logger log.Logger, workReportUC *biz.WorkReportUC) *WorkReportService {
	return &WorkReportService{
		log:          log.NewHelper(log.With(logger, "module", "service/work_report")),
		workReportUC: workReportUC,
	}
}
func (s *WorkReportService) RegisterGRPC(srv *grpc.Server) {
	singletonV1.RegisterWorkReportServiceServer(srv, s)
}
func (s *WorkReportService) RegisterHTTP(srv *http.Server) {
	singletonV1.RegisterWorkReportServiceHTTPServer(srv, s)
}
func (s *WorkReportService) RegisterConnect(srv *connect.Server) {
	srv.Register(singletonV1connect.NewWorkReportServiceHandler(s, srv.HandlerOptions()...))
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *WorkReportService) CreateWorkReport(ctx context.Context, in *singletonV1.CreateWorkReportRequest) (*singletonV1.CreateWorkReportResponse, error) {
	report := &biz.WorkReport{
		Title:        in.Title,
		Content:      in.Content,
		Type:         in.Type,
		DepartmentID: in.DepartmentId,
		AccessLevel:  utils.ConvPtr[int](in.AccessLevel),
		Region:       in.Region,
	}
	result, err := s.workReportUC.Create(ctx, report)
	if err != nil {
		return nil, err
	}
	return &singletonV1.CreateWorkReportResponse{Id: *result.ID}, nil
}

func (s *WorkReportService) UpdateWorkReport(ctx context.Context, in *singletonV1.UpdateWorkReportRequest) (*singletonV1.UpdateWorkReportResponse, error) {
	report := &biz.WorkReport{
		ID:          &in.Id,
		Title:       in.Title,
		Content:     in.Content,
		Type:        in.Type,
		AccessLevel: utils.ConvPtr[int](in.AccessLevel),
		Region:      in.Region,
		Status:      in.Status,
	}
	if err := s.workReportUC.Update(ctx, report, in.FieldsMask); err != nil {
		return nil, err
	}
	return &singletonV1.UpdateWorkReportResponse{}, nil
}

func (s *WorkReportService) DeleteWorkReport(ctx context.Context, in *singletonV1.DeleteWorkReportRequest) (*singletonV1.DeleteWorkReportResponse, error) {
	if err := s.workReportUC.Delete(ctx, in.Id); err != nil {
		return nil, err
	}
	return &singletonV1.DeleteWorkReportResponse{}, nil
}

func (s *WorkReportService) GetWorkReport(ctx context.Context, in *singletonV1.GetWorkReportRequest) (*singletonV1.GetWorkReportResponse, error) {
	result, err := s.workReportUC.Get(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return s.workReportToProto(result), nil
}

func (s *WorkReportService) QueryWorkReport(ctx context.Context, in *singletonV1.QueryWorkReportRequest) (*singletonV1.QueryWorkReportResponse, error) {
	result, err := s.workReportUC.Query(ctx, &biz.WorkReportQueryIn{
		PageRequest: utils.EnsurePageRequest(in.Page),
		Type:        in.Type,
		Status:      in.Status,
	})
	if err != nil {
		return nil, err
	}
	return &singletonV1.QueryWorkReportResponse{
		Page: result.PageResponse,
		List: utils.SliceMap(result.List, s.workReportToProto),
	}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func (s *WorkReportService) workReportToProto(e *biz.WorkReport) *singletonV1.GetWorkReportResponse {
	return &singletonV1.GetWorkReportResponse{
		Id:           *e.ID,
		CreatedAt:    utils.ToTimestamp(e.CreatedAt),
		UpdatedAt:    utils.ToTimestamp(e.UpdatedAt),
		Title:        utils.Wrap(e.Title, utils.StringW),
		Content:      utils.Wrap(e.Content, utils.StringW),
		Type:         utils.Wrap(e.Type, utils.StringW),
		DepartmentId: utils.Wrap(e.DepartmentID, utils.StringW),
		AccessLevel:  utils.Wrap(e.AccessLevel, func(v int) *wrapperspb.Int32Value { return wrapperspb.Int32(int32(v)) }),
		Region:       utils.Wrap(e.Region, utils.StringW),
		CreatedBy:    utils.Wrap(e.CreatedBy, utils.StringW),
		Status:       utils.Wrap(e.Status, utils.StringW),
	}
}
