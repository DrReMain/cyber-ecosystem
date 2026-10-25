package entitle

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
)

// Struct ------------------------------------------------------------------------------------------------------

type EntitlementService struct {
	esportspb.UnimplementedEntitlementServiceServer

	log       *slog.Logger
	entitleUC *EntitleUC
}

func NewEntitlementService(logger *slog.Logger, entitleUC *EntitleUC) *EntitlementService {
	return &EntitlementService{
		log:       logger.With("module", "module/entitle_service"),
		entitleUC: entitleUC,
	}
}

// Handler --------------------------------------------------------------------------------------------------------------

func (s *EntitlementService) RedeemActivationCode(ctx context.Context, req *esportspb.RedeemActivationCodeRequest) (*esportspb.RedeemActivationCodeResponse, error) {
	out, err := s.entitleUC.Redeem(ctx, req.GetCode())
	if err != nil {
		return nil, err
	}
	resp := &esportspb.RedeemActivationCodeResponse{ReportType: mapReportTypePB(out.ReportType)}
	if out.ExpiresAt != nil {
		resp.ExpiresAt = utils.ToTimestamp(out.ExpiresAt)
	}
	resp.GrantedAt = utils.ToTimestamp(&out.GrantedAt)
	return resp, nil
}

func (s *EntitlementService) GetReportAccess(ctx context.Context, req *esportspb.GetReportAccessRequest) (*esportspb.GetReportAccessResponse, error) {
	out, err := s.entitleUC.GetReportAccess(ctx, mapReportTypeFromPB(req.GetReportType()))
	if err != nil {
		return nil, err
	}
	return &esportspb.GetReportAccessResponse{Allowed: out.Allowed, Reason: out.Reason}, nil
}

func (s *EntitlementService) RegisterGRPC(srv *grpc.Server) {
	esportspb.RegisterEntitlementServiceServer(srv, s)
}

func (s *EntitlementService) RegisterHTTP(srv *http.Server) {
	esportspb.RegisterEntitlementServiceHTTPServer(srv, s)
}

func (s *EntitlementService) RegisterConnect(srv *connecttransport.Server) {
	esportspb.RegisterEntitlementServiceConnectServer(srv, s)
}

// Private -------------------------------------------------------------------------------------------------------------

func mapReportTypePB(family string) esportspb.ReportType {
	switch family {
	case reportCognitive:
		return esportspb.ReportType_REPORT_TYPE_COGNITIVE
	case reportCareer:
		return esportspb.ReportType_REPORT_TYPE_CAREER
	case reportSchool:
		return esportspb.ReportType_REPORT_TYPE_SCHOOL
	}
	return esportspb.ReportType_REPORT_TYPE_UNSPECIFIED
}

func mapReportTypeFromPB(t esportspb.ReportType) string {
	switch t {
	case esportspb.ReportType_REPORT_TYPE_COGNITIVE:
		return reportCognitive
	case esportspb.ReportType_REPORT_TYPE_CAREER:
		return reportCareer
	case esportspb.ReportType_REPORT_TYPE_SCHOOL:
		return reportSchool
	}
	return ""
}
