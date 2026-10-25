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

// Struct --------------------------------------------------------------------------------------------------------------

type ActivationCodeService struct {
	esportspb.UnimplementedActivationCodeServiceServer

	log    *slog.Logger
	codeUC *CodeUC
}

func NewActivationCodeService(logger *slog.Logger, codeUC *CodeUC) *ActivationCodeService {
	return &ActivationCodeService{
		log:    logger.With("module", "module/entitle_code_service"),
		codeUC: codeUC,
	}
}

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

// Method --------------------------------------------------------------------------------------------------------------

func (s *ActivationCodeService) GenerateActivationCodes(ctx context.Context, req *esportspb.GenerateActivationCodesRequest) (*esportspb.GenerateActivationCodesResponse, error) {
	codes, err := s.codeUC.Generate(ctx, &GenerateIn{
		ReportType: mapReportTypeFromPB(req.GetReportType()),
		Count:      int(req.GetCount()),
		ExpireDays: int(req.GetExpireDays()),
	})
	if err != nil {
		return nil, err
	}
	return &esportspb.GenerateActivationCodesResponse{Codes: codes}, nil
}

func (s *ActivationCodeService) ListActivationCodes(ctx context.Context, req *esportspb.ListActivationCodesRequest) (*esportspb.ListActivationCodesResponse, error) {
	var reportType *string
	if req.GetReportType() != esportspb.ReportType_REPORT_TYPE_UNSPECIFIED {
		v := mapReportTypeFromPB(req.GetReportType())
		reportType = &v
	}
	out, err := s.codeUC.List(ctx, &ListCodesIn{PageRequest: req.GetPage(), ReportType: reportType})
	if err != nil {
		return nil, err
	}
	return &esportspb.ListActivationCodesResponse{
		Page:  out.PageResponse,
		Codes: utils.SliceMap(out.Codes, mapCodePB),
	}, nil
}

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

func (s *EntitlementService) SetReportAccessPolicy(ctx context.Context, req *esportspb.SetReportAccessPolicyRequest) (*esportspb.SetReportAccessPolicyResponse, error) {
	if err := s.entitleUC.SetReportAccessPolicy(ctx, mapReportTypeFromPB(req.GetReportType()), mapPolicyFromPB(req.GetPolicy())); err != nil {
		return nil, err
	}
	return &esportspb.SetReportAccessPolicyResponse{
		ReportType: req.GetReportType(),
		Policy:     req.GetPolicy(),
	}, nil
}

func (s *ActivationCodeService) RegisterGRPC(srv *grpc.Server) {
	esportspb.RegisterActivationCodeServiceServer(srv, s)
}

func (s *ActivationCodeService) RegisterHTTP(srv *http.Server) {
	esportspb.RegisterActivationCodeServiceHTTPServer(srv, s)
}

func (s *ActivationCodeService) RegisterConnect(srv *connecttransport.Server) {
	esportspb.RegisterActivationCodeServiceConnectServer(srv, s)
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

func mapCodePB(r *ActivationCode) *esportspb.ActivationCode {
	row := &esportspb.ActivationCode{
		ReportType: mapReportTypePB(r.ReportType),
		Quantity:   utils.Int32W(utils.ConvNum[int32, int](r.Quantity)),
		UsedCount:  utils.Int32W(utils.ConvNum[int32, int](r.UsedCount)),
		ExpiresAt:  utils.ToTimestamp(&r.ExpiresAt),
		CreatedAt:  utils.ToTimestamp(&r.CreatedAt),
	}
	row.Code = utils.StringW(r.Code)
	return row
}

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

func mapPolicyFromPB(p esportspb.ReportAccessPolicy) string {
	switch p {
	case esportspb.ReportAccessPolicy_REPORT_ACCESS_POLICY_OPEN:
		return policyOpen
	case esportspb.ReportAccessPolicy_REPORT_ACCESS_POLICY_MEMBERSHIP:
		return policyMembership
	case esportspb.ReportAccessPolicy_REPORT_ACCESS_POLICY_TENANT:
		return policyTenant
	case esportspb.ReportAccessPolicy_REPORT_ACCESS_POLICY_CLOSED:
		return policyClosed
	}
	return ""
}
