package entitleadmin

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"

	"cyber-ecosystem/app/services/esports/internal/module/entitle"
)

// Struct --------------------------------------------------------------------------------------------------------------

type EntitleAdminService struct {
	esportspb.UnimplementedEntitleAdminServiceServer

	log       *slog.Logger
	codeUC    *entitle.CodeUC
	entitleUC *entitle.EntitleUC
}

func NewEntitleAdminService(logger *slog.Logger, codeUC *entitle.CodeUC, entitleUC *entitle.EntitleUC) *EntitleAdminService {
	return &EntitleAdminService{
		log:       logger.With("module", "module/entitle_admin_service"),
		codeUC:    codeUC,
		entitleUC: entitleUC,
	}
}

// Handler ------------------------------------------------------------------------------------------------------------

func (s *EntitleAdminService) GenerateActivationCodes(ctx context.Context, req *esportspb.GenerateActivationCodesRequest) (*esportspb.GenerateActivationCodesResponse, error) {
	codes, err := s.codeUC.Generate(ctx, &entitle.GenerateIn{
		ReportType: reportTypeFromPB(req.GetReportType()),
		Count:      int(req.GetCount()),
		ExpireDays: int(req.GetExpireDays()),
	})
	if err != nil {
		return nil, err
	}
	return &esportspb.GenerateActivationCodesResponse{Codes: codes}, nil
}

func (s *EntitleAdminService) ListActivationCodes(ctx context.Context, req *esportspb.ListActivationCodesRequest) (*esportspb.ListActivationCodesResponse, error) {
	var reportType *string
	if req.GetReportType() != esportspb.ReportType_REPORT_TYPE_UNSPECIFIED {
		v := reportTypeFromPB(req.GetReportType())
		reportType = &v
	}
	out, err := s.codeUC.List(ctx, &entitle.ListCodesIn{PageRequest: req.GetPage(), ReportType: reportType})
	if err != nil {
		return nil, err
	}
	return &esportspb.ListActivationCodesResponse{
		Page:  out.PageResponse,
		Codes: utils.SliceMap(out.Codes, codePB),
	}, nil
}

func (s *EntitleAdminService) SetReportAccessPolicy(ctx context.Context, req *esportspb.SetReportAccessPolicyRequest) (*esportspb.SetReportAccessPolicyResponse, error) {
	if err := s.entitleUC.SetReportAccessPolicy(ctx, reportTypeFromPB(req.GetReportType()), policyFromPB(req.GetPolicy())); err != nil {
		return nil, err
	}
	return &esportspb.SetReportAccessPolicyResponse{
		ReportType: req.GetReportType(),
		Policy:     req.GetPolicy(),
	}, nil
}

func (s *EntitleAdminService) RegisterGRPC(srv *grpc.Server) {
	esportspb.RegisterEntitleAdminServiceServer(srv, s)
}

func (s *EntitleAdminService) RegisterHTTP(srv *http.Server) {
	esportspb.RegisterEntitleAdminServiceHTTPServer(srv, s)
}

func (s *EntitleAdminService) RegisterConnect(srv *connecttransport.Server) {
	esportspb.RegisterEntitleAdminServiceConnectServer(srv, s)
}

// Private -------------------------------------------------------------------------------------------------------------

func reportTypeFromPB(t esportspb.ReportType) string {
	switch t {
	case esportspb.ReportType_REPORT_TYPE_COGNITIVE:
		return "cognitive"
	case esportspb.ReportType_REPORT_TYPE_CAREER:
		return "career"
	case esportspb.ReportType_REPORT_TYPE_SCHOOL:
		return "school"
	}
	return ""
}

func policyFromPB(p esportspb.ReportAccessPolicy) string {
	switch p {
	case esportspb.ReportAccessPolicy_REPORT_ACCESS_POLICY_OPEN:
		return "open"
	case esportspb.ReportAccessPolicy_REPORT_ACCESS_POLICY_MEMBERSHIP:
		return "membership"
	case esportspb.ReportAccessPolicy_REPORT_ACCESS_POLICY_TENANT:
		return "tenant"
	case esportspb.ReportAccessPolicy_REPORT_ACCESS_POLICY_CLOSED:
		return "closed"
	}
	return ""
}

func codePB(r *entitle.ActivationCode) *esportspb.ActivationCode {
	row := &esportspb.ActivationCode{
		ReportType: esportspb.ReportType(esportspb.ReportType_value[r.ReportType]),
		Quantity:   utils.Int32W(utils.ConvNum[int32, int](r.Quantity)),
		UsedCount:  utils.Int32W(utils.ConvNum[int32, int](r.UsedCount)),
		ExpiresAt:  utils.ToTimestamp(&r.ExpiresAt),
		CreatedAt:  utils.ToTimestamp(&r.CreatedAt),
	}
	row.Code = utils.StringW(r.Code)
	return row
}
