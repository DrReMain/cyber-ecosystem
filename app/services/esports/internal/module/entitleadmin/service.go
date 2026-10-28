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
	"cyber-ecosystem/app/services/esports/internal/shared"
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
		log:       logger.With("module", "module/entitleadmin_service"),
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

func (s *EntitleAdminService) ListReportAccessPolicies(ctx context.Context, _ *esportspb.ListReportAccessPoliciesRequest) (*esportspb.ListReportAccessPoliciesResponse, error) {
	rows, err := s.entitleUC.ListReportAccessPolicies(ctx)
	if err != nil {
		return nil, err
	}
	policies := make([]*esportspb.ReportPolicyRow, 0, len(rows))
	for _, r := range rows {
		row := &esportspb.ReportPolicyRow{
			ReportType: reportTypeToPB(r.Family),
			Policy:     policyToPB(r.Policy),
			Overridden: r.Overridden,
			UpdatedAt:  utils.ToTimestamp(&r.UpdatedAt),
		}
		if r.Overridden {
			row.UpdatedBy = utils.StringW(r.UpdatedBy)
		}
		policies = append(policies, row)
	}
	return &esportspb.ListReportAccessPoliciesResponse{Policies: policies}, nil
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
		return shared.ReportCognitive
	case esportspb.ReportType_REPORT_TYPE_CAREER:
		return shared.ReportCareer
	case esportspb.ReportType_REPORT_TYPE_SCHOOL:
		return shared.ReportSchool
	}
	return ""
}

func policyFromPB(p esportspb.ReportAccessPolicy) string {
	switch p {
	case esportspb.ReportAccessPolicy_REPORT_ACCESS_POLICY_OPEN:
		return shared.PolicyOpen
	case esportspb.ReportAccessPolicy_REPORT_ACCESS_POLICY_MEMBERSHIP:
		return shared.PolicyMembership
	case esportspb.ReportAccessPolicy_REPORT_ACCESS_POLICY_TENANT:
		return shared.PolicyTenant
	case esportspb.ReportAccessPolicy_REPORT_ACCESS_POLICY_CLOSED:
		return shared.PolicyClosed
	}
	return ""
}

func policyToPB(p string) esportspb.ReportAccessPolicy {
	switch p {
	case shared.PolicyOpen:
		return esportspb.ReportAccessPolicy_REPORT_ACCESS_POLICY_OPEN
	case shared.PolicyMembership:
		return esportspb.ReportAccessPolicy_REPORT_ACCESS_POLICY_MEMBERSHIP
	case shared.PolicyTenant:
		return esportspb.ReportAccessPolicy_REPORT_ACCESS_POLICY_TENANT
	case shared.PolicyClosed:
		return esportspb.ReportAccessPolicy_REPORT_ACCESS_POLICY_CLOSED
	}
	return esportspb.ReportAccessPolicy_REPORT_ACCESS_POLICY_UNSPECIFIED
}

func reportTypeToPB(f string) esportspb.ReportType {
	switch f {
	case shared.ReportCognitive:
		return esportspb.ReportType_REPORT_TYPE_COGNITIVE
	case shared.ReportCareer:
		return esportspb.ReportType_REPORT_TYPE_CAREER
	case shared.ReportSchool:
		return esportspb.ReportType_REPORT_TYPE_SCHOOL
	}
	return esportspb.ReportType_REPORT_TYPE_UNSPECIFIED
}

func codePB(r *entitle.ActivationCode) *esportspb.ActivationCode {
	row := &esportspb.ActivationCode{
		// The DO carries the domain word ("career"), not the proto value
		// name — the enum lookup table keyed on proto names answered
		// UNSPECIFIED for every row; the inverse mapper is the right door.
		ReportType: reportTypeToPB(r.ReportType),
		Quantity:   utils.Int32W(utils.ConvNum[int32, int](r.Quantity)),
		UsedCount:  utils.Int32W(utils.ConvNum[int32, int](r.UsedCount)),
		ExpiresAt:  utils.ToTimestamp(&r.ExpiresAt),
		CreatedAt:  utils.ToTimestamp(&r.CreatedAt),
	}
	row.Code = utils.StringW(r.Code)
	return row
}
