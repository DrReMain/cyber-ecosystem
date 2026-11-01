package agreementadmin

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	"cyber-ecosystem/shared-go/helper"
	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
	commonv1 "cyber-ecosystem/gen/go/cyber/shared/common/v1"

	"cyber-ecosystem/app/services/esports/internal/module/agreement"
)

// Struct --------------------------------------------------------------------------------------------------------------

type AgreementAdminService struct {
	esportspb.UnimplementedAgreementAdminServiceServer

	log              *slog.Logger
	agreementAdminUC *agreement.AgreementAdminUC
}

func NewAgreementAdminService(logger *slog.Logger, agreementAdminUC *agreement.AgreementAdminUC) *AgreementAdminService {
	return &AgreementAdminService{
		log:              logger.With("module", "module/agreementadmin_service"),
		agreementAdminUC: agreementAdminUC,
	}
}

func (s *AgreementAdminService) RegisterGRPC(srv *grpc.Server) {
	esportspb.RegisterAgreementAdminServiceServer(srv, s)
}

func (s *AgreementAdminService) RegisterHTTP(srv *http.Server) {
	esportspb.RegisterAgreementAdminServiceHTTPServer(srv, s)
}

func (s *AgreementAdminService) RegisterConnect(srv *connecttransport.Server) {
	esportspb.RegisterAgreementAdminServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *AgreementAdminService) CreateAgreementVersion(ctx context.Context, in *esportspb.CreateAgreementVersionRequest) (*esportspb.CreateAgreementVersionResponse, error) {
	version, err := s.agreementAdminUC.CreateVersion(ctx,
		agreement.DocFromPB(in.GetDoc()), in.GetNo(), in.GetEffectiveAt().AsTime(), in.GetBody())
	if err != nil {
		return nil, err
	}
	return &esportspb.CreateAgreementVersionResponse{Version: versionPB(version)}, nil
}

func (s *AgreementAdminService) SetAgreementCurrent(ctx context.Context, in *esportspb.SetAgreementCurrentRequest) (*esportspb.SetAgreementCurrentResponse, error) {
	version, err := s.agreementAdminUC.SetCurrent(ctx, in.GetVersionId())
	if err != nil {
		return nil, err
	}
	return &esportspb.SetAgreementCurrentResponse{Version: versionPB(version)}, nil
}

func (s *AgreementAdminService) DeleteAgreementVersion(ctx context.Context, in *esportspb.DeleteAgreementVersionRequest) (*esportspb.DeleteAgreementVersionResponse, error) {
	if err := s.agreementAdminUC.DeleteVersion(ctx, in.GetVersionId()); err != nil {
		return nil, err
	}
	return &esportspb.DeleteAgreementVersionResponse{}, nil
}

func (s *AgreementAdminService) ListAgreementVersions(ctx context.Context, in *esportspb.ListAgreementVersionsRequest) (*esportspb.ListAgreementVersionsResponse, error) {
	var doc *string
	if in.Doc != nil {
		folded := agreement.DocFromPB(*in.Doc)
		doc = &folded
	}
	versions, err := s.agreementAdminUC.ListVersions(ctx, doc)
	if err != nil {
		return nil, err
	}
	return &esportspb.ListAgreementVersionsResponse{
		Versions: utils.SliceMap(versions, versionPB),
	}, nil
}

func (s *AgreementAdminService) ListAgreementConsents(ctx context.Context, in *esportspb.ListAgreementConsentsRequest) (*esportspb.ListAgreementConsentsResponse, error) {
	out, err := s.agreementAdminUC.ListConsents(ctx,
		consentsIn(in.GetPage(), in.GetDoc(), in.GetVersionNo(), in.GetKeyword()))
	if err != nil {
		return nil, err
	}
	return &esportspb.ListAgreementConsentsResponse{
		Page:     out.PageResponse,
		Consents: utils.SliceMap(out.Rows, consentPB),
	}, nil
}

func (s *AgreementAdminService) ExportAgreementConsents(ctx context.Context, in *esportspb.ExportAgreementConsentsRequest) (*esportspb.ExportAgreementConsentsResponse, error) {
	// The export answers the visible filter, never a wider set — the
	// same filters the list applies, on the unpaged read the port gives
	// the export.
	rows, err := s.agreementAdminUC.ExportConsents(ctx,
		consentsIn(&commonv1.PageRequest{CreatedAtA: in.CreatedAtA, CreatedAtZ: in.CreatedAtZ},
			in.GetDoc(), in.GetVersionNo(), in.GetKeyword()))
	if err != nil {
		return nil, err
	}
	return &esportspb.ExportAgreementConsentsResponse{
		Consents: utils.SliceMap(rows, consentPB),
	}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func consentsIn(page *commonv1.PageRequest, doc esportspb.AgreementDoc, versionNo, keyword string) *agreement.ListConsentsIn {
	out := &agreement.ListConsentsIn{PageRequest: helper.EnsurePageRequest(page)}
	if doc != esportspb.AgreementDoc_AGREEMENT_DOC_UNSPECIFIED {
		folded := agreement.DocFromPB(doc)
		out.Doc = &folded
	}
	if versionNo != "" {
		out.VersionNo = &versionNo
	}
	if keyword != "" {
		out.Keyword = &keyword
	}
	return out
}

func versionPB(v *agreement.AgreementVersion) *esportspb.AgreementVersion {
	return &esportspb.AgreementVersion{
		Id:          utils.StringW(v.ID),
		Doc:         agreement.DocPB(v.Doc),
		No:          utils.StringW(v.No),
		Title:       utils.StringW(v.Title),
		EffectiveAt: utils.ToTimestamp(&v.EffectiveAt),
		AgreedCount: utils.Int32W(int32(v.AgreedCount)),
		IsCurrent:   utils.BoolW(v.IsCurrent),
		Body:        utils.StringW(v.Body),
	}
}

func consentPB(row *agreement.AdminConsentRow) *esportspb.AgreementConsent {
	return &esportspb.AgreementConsent{
		UserId:      utils.StringW(row.UserID),
		Nickname:    utils.Wrap(row.Nickname, utils.StringW),
		PhoneMasked: utils.Wrap(row.PhoneMasked, utils.StringW),
		Doc:         agreement.DocPB(row.Doc),
		VersionNo:   utils.StringW(row.VersionNo),
		ConsentedAt: utils.ToTimestamp(&row.ConsentAt),
		LoginMethod: agreement.LoginMethodPB(row.LoginMethod),
	}
}
