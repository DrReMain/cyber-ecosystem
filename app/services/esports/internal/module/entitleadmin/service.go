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

	log          *slog.Logger
	codeUC       *entitle.CodeUC
	entitleUC    *entitle.EntitleUC
	membershipUC *MembershipUC
}

func NewEntitleAdminService(logger *slog.Logger, codeUC *entitle.CodeUC, entitleUC *entitle.EntitleUC, membershipUC *MembershipUC) *EntitleAdminService {
	return &EntitleAdminService{
		log:          logger.With("module", "module/entitleadmin_service"),
		codeUC:       codeUC,
		entitleUC:    entitleUC,
		membershipUC: membershipUC,
	}
}

// Handler ------------------------------------------------------------------------------------------------------------

func (s *EntitleAdminService) GenerateActivationCodes(ctx context.Context, req *esportspb.GenerateActivationCodesRequest) (*esportspb.GenerateActivationCodesResponse, error) {
	out, err := s.codeUC.Generate(ctx, &entitle.GenerateIn{
		ReportType: reportTypeFromPB(req.GetReportType()),
		Count:      int(req.GetCount()),
		ExpireDays: int(req.GetExpireDays()),
	})
	if err != nil {
		return nil, err
	}
	return &esportspb.GenerateActivationCodesResponse{Codes: out.Codes, BatchNo: out.BatchNo}, nil
}

func (s *EntitleAdminService) SetActivationBatchStatus(ctx context.Context, req *esportspb.SetActivationBatchStatusRequest) (*esportspb.SetActivationBatchStatusResponse, error) {
	batch, err := s.codeUC.SetBatchStatus(ctx, req.GetBatchNo(), req.GetDisabled())
	if err != nil {
		return nil, err
	}
	return &esportspb.SetActivationBatchStatusResponse{Batch: batchPB(batch)}, nil
}

func (s *EntitleAdminService) ListActivationCodes(ctx context.Context, req *esportspb.ListActivationCodesRequest) (*esportspb.ListActivationCodesResponse, error) {
	out, err := s.codeUC.List(ctx, &entitle.ListCodesIn{
		PageRequest: req.GetPage(),
		ReportType:  reportTypePtrFromPB(req.GetReportType()),
		BatchNo:     req.BatchNo,
		Status:      codeStatusPtrFromPB(req.GetCodeStatus()),
		Code:        req.Code,
		User:        req.UserId,
	})
	if err != nil {
		return nil, err
	}
	return &esportspb.ListActivationCodesResponse{
		Page:  out.PageResponse,
		Codes: utils.SliceMap(out.Codes, codePB),
	}, nil
}

func (s *EntitleAdminService) ListActivationBatches(ctx context.Context, req *esportspb.ListActivationBatchesRequest) (*esportspb.ListActivationBatchesResponse, error) {
	out, err := s.codeUC.ListBatches(ctx, &entitle.ListBatchesIn{
		PageRequest: req.GetPage(),
		ReportType:  reportTypePtrFromPB(req.GetReportType()),
		Disabled:    req.Disabled,
		BatchNo:     req.BatchNo,
	})
	if err != nil {
		return nil, err
	}
	return &esportspb.ListActivationBatchesResponse{
		Page:         out.PageResponse,
		Batches:      utils.SliceMap(out.Batches, batchPB),
		TotalBatches: utils.Int32W(utils.ConvNum[int32, int](out.TotalBatches)),
		TotalCodes:   utils.Int32W(utils.ConvNum[int32, int](out.TotalCodes)),
		UsedCodes:    utils.Int32W(utils.ConvNum[int32, int](out.UsedCodes)),
	}, nil
}

func (s *EntitleAdminService) ExportActivationCodes(ctx context.Context, req *esportspb.ExportActivationCodesRequest) (*esportspb.ExportActivationCodesResponse, error) {
	codes, err := s.codeUC.Export(ctx, &entitle.ListCodesIn{
		BatchNo: req.BatchNo,
		Status:  codeStatusPtrFromPB(req.GetCodeStatus()),
		Code:    req.Code,
	})
	if err != nil {
		return nil, err
	}
	return &esportspb.ExportActivationCodesResponse{Codes: utils.SliceMap(codes, codePB)}, nil
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

func (s *EntitleAdminService) GrantMembership(ctx context.Context, req *esportspb.GrantMembershipRequest) (*esportspb.GrantMembershipResponse, error) {
	facts, err := s.membershipUC.Grant(ctx, req.GetUserId(), req.GetSkuCode())
	if err != nil {
		return nil, err
	}
	return &esportspb.GrantMembershipResponse{Membership: membershipSummaryPB(facts)}, nil
}

func (s *EntitleAdminService) TerminateMembership(ctx context.Context, req *esportspb.TerminateMembershipRequest) (*esportspb.TerminateMembershipResponse, error) {
	facts, err := s.membershipUC.Terminate(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}
	return &esportspb.TerminateMembershipResponse{Membership: membershipSummaryPB(facts)}, nil
}

func (s *EntitleAdminService) ListMemberships(ctx context.Context, req *esportspb.ListMembershipsRequest) (*esportspb.ListMembershipsResponse, error) {
	in := &ListMembershipsIn{PageRequest: req.GetPage()}
	if req.GetStatus() != esportspb.MembershipStatus_MEMBERSHIP_STATUS_UNSPECIFIED {
		v := statusFromPB(req.GetStatus())
		in.Status = &v
	}
	if req.GetSource() != esportspb.MembershipSource_MEMBERSHIP_SOURCE_UNSPECIFIED {
		v := sourceFromPB(req.GetSource())
		in.Source = &v
	}
	out, err := s.membershipUC.List(ctx, in)
	if err != nil {
		return nil, err
	}
	resp := &esportspb.ListMembershipsResponse{
		Page:              out.PageResponse,
		Memberships:       make([]*esportspb.MembershipSummary, 0, len(out.Rows)),
		TotalCount:        utils.Int32W(utils.ConvNum[int32, int](out.Counts.Total)),
		ActiveCount:       utils.Int32W(utils.ConvNum[int32, int](out.Counts.Active)),
		ExpiringSoonCount: utils.Int32W(utils.ConvNum[int32, int](out.Counts.ExpiringSoon)),
		ExpiredCount:      utils.Int32W(utils.ConvNum[int32, int](out.Counts.Expired)),
	}
	for _, r := range out.Rows {
		summary := membershipSummaryPB(&r.MembershipFacts)
		summary.Nickname = utils.Wrap(r.Nickname, utils.StringW)
		summary.PhoneMasked = utils.Wrap(r.PhoneMasked, utils.StringW)
		resp.Memberships = append(resp.Memberships, summary)
	}
	return resp, nil
}

func (s *EntitleAdminService) GetMembership(ctx context.Context, req *esportspb.GetMembershipRequest) (*esportspb.GetMembershipResponse, error) {
	detail, err := s.membershipUC.Get(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}
	resp := &esportspb.GetMembershipResponse{
		Membership:  membershipSummaryPB(detail.Facts),
		Periods:     make([]*esportspb.MembershipPeriod, 0, len(detail.History)),
		AgeBand:     utils.Wrap(detail.Identity.AgeBand, utils.StringW),
		Sex:         sexToPB(detail.Identity.Sex),
		WechatBound: utils.BoolW(detail.Identity.WechatBound),
	}
	// The header's identity facts ride the summary too — the list and the
	// detail read the same shape, one consumer each side.
	resp.Membership.Nickname = utils.Wrap(detail.Identity.Nickname, utils.StringW)
	resp.Membership.PhoneMasked = utils.Wrap(detail.Identity.PhoneMasked, utils.StringW)
	for _, h := range detail.History {
		resp.Periods = append(resp.Periods, &esportspb.MembershipPeriod{
			GrantedAt: utils.ToTimestamp(&h.GrantedAt),
			ExpiresAt: utils.ToTimestamp(&h.ExpiresAt),
			Source:    sourceToPB(h.Source),
			SourceRef: utils.StringW(h.SourceRef),
			Status:    statusToPB(h.Status),
		})
	}
	return resp, nil
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
		BatchNo:    utils.Wrap(r.BatchNo, utils.StringW),
		RedeemedAt: utils.ToTimestamp(r.RedeemedAt),
		UserId:     utils.Wrap(r.RedeemedBy, utils.StringW),
		Nickname:   utils.Wrap(r.Nickname, utils.StringW),
	}
	row.Code = utils.StringW(r.Code)
	return row
}

func batchPB(b *entitle.ActivationBatch) *esportspb.ActivationBatch {
	return &esportspb.ActivationBatch{
		BatchNo:    utils.StringW(b.BatchNo),
		ReportType: reportTypeToPB(b.ReportType),
		Count:      utils.Int32W(utils.ConvNum[int32, int](b.Count)),
		UsedCount:  utils.Int32W(utils.ConvNum[int32, int](b.UsedCount)),
		ExpiresAt:  utils.ToTimestamp(&b.ExpiresAt),
		CreatedAt:  utils.ToTimestamp(&b.CreatedAt),
		DisabledAt: utils.ToTimestamp(b.DisabledAt),
	}
}

func reportTypePtrFromPB(t esportspb.ReportType) *string {
	// The filters' direction: UNSPECIFIED folds to "no filter" — the
	// caller omits rather than asks for it.
	if t == esportspb.ReportType_REPORT_TYPE_UNSPECIFIED {
		return nil
	}
	v := reportTypeFromPB(t)
	return &v
}

func codeStatusPtrFromPB(t esportspb.ActivationCodeStatus) *string {
	switch t {
	case esportspb.ActivationCodeStatus_ACTIVATION_CODE_STATUS_UNUSED:
		v := entitle.CodeStatusUnused
		return &v
	case esportspb.ActivationCodeStatus_ACTIVATION_CODE_STATUS_REDEEMED:
		v := entitle.CodeStatusRedeemed
		return &v
	}
	return nil
}

func membershipSummaryPB(f *MembershipFacts) *esportspb.MembershipSummary {
	summary := &esportspb.MembershipSummary{
		UserId:      utils.StringW(f.UserID),
		GrantedAt:   utils.ToTimestamp(&f.GrantedAt),
		ExpiresAt:   utils.ToTimestamp(&f.ExpiresAt),
		Status:      statusToPB(f.Status),
		Source:      sourceToPB(f.Source),
		PeriodCount: utils.Int32W(utils.ConvNum[int32, int](f.PeriodCount)),
		UnlockCount: utils.Int32W(utils.ConvNum[int32, int](f.UnlockCount)),
	}
	if f.DaysLeft != nil {
		summary.DaysLeft = utils.Int32W(utils.ConvNum[int32, int](*f.DaysLeft))
	}
	return summary
}

func statusToPB(s string) esportspb.MembershipStatus {
	switch s {
	case displayActive:
		return esportspb.MembershipStatus_MEMBERSHIP_STATUS_ACTIVE
	case displayExpiringSoon:
		return esportspb.MembershipStatus_MEMBERSHIP_STATUS_EXPIRING_SOON
	case displayExpired:
		return esportspb.MembershipStatus_MEMBERSHIP_STATUS_EXPIRED
	case displayTerminated:
		return esportspb.MembershipStatus_MEMBERSHIP_STATUS_TERMINATED
	}
	return esportspb.MembershipStatus_MEMBERSHIP_STATUS_UNSPECIFIED
}

func statusFromPB(s esportspb.MembershipStatus) string {
	switch s {
	case esportspb.MembershipStatus_MEMBERSHIP_STATUS_ACTIVE:
		return displayActive
	case esportspb.MembershipStatus_MEMBERSHIP_STATUS_EXPIRING_SOON:
		return displayExpiringSoon
	case esportspb.MembershipStatus_MEMBERSHIP_STATUS_EXPIRED:
		return displayExpired
	case esportspb.MembershipStatus_MEMBERSHIP_STATUS_TERMINATED:
		return displayTerminated
	}
	return ""
}

func sourceToPB(s string) esportspb.MembershipSource {
	switch s {
	case sourceManual:
		return esportspb.MembershipSource_MEMBERSHIP_SOURCE_MANUAL
	case sourceLegacyCode:
		return esportspb.MembershipSource_MEMBERSHIP_SOURCE_LEGACY_CODE
	case sourceOrder:
		return esportspb.MembershipSource_MEMBERSHIP_SOURCE_ORDER
	}
	return esportspb.MembershipSource_MEMBERSHIP_SOURCE_UNSPECIFIED
}

func sourceFromPB(s esportspb.MembershipSource) string {
	switch s {
	case esportspb.MembershipSource_MEMBERSHIP_SOURCE_MANUAL:
		return sourceManual
	case esportspb.MembershipSource_MEMBERSHIP_SOURCE_LEGACY_CODE:
		return sourceLegacyCode
	case esportspb.MembershipSource_MEMBERSHIP_SOURCE_ORDER:
		return sourceOrder
	}
	return ""
}

func sexToPB(sex *string) esportspb.ProfileSex {
	if sex == nil {
		return esportspb.ProfileSex_PROFILE_SEX_UNSPECIFIED
	}
	switch *sex {
	case shared.SexMale:
		return esportspb.ProfileSex_PROFILE_SEX_MALE
	case shared.SexFemale:
		return esportspb.ProfileSex_PROFILE_SEX_FEMALE
	case shared.SexUndetermined:
		return esportspb.ProfileSex_PROFILE_SEX_UNDETERMINED
	}
	return esportspb.ProfileSex_PROFILE_SEX_UNSPECIFIED
}
