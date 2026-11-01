package agreement

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

type AgreementService struct {
	esportspb.UnimplementedAgreementServiceServer

	log         *slog.Logger
	agreementUC *AgreementUC
}

func NewAgreementService(logger *slog.Logger, agreementUC *AgreementUC) *AgreementService {
	return &AgreementService{
		log:         logger.With("module", "module/agreement_service"),
		agreementUC: agreementUC,
	}
}

func (s *AgreementService) RegisterGRPC(srv *grpc.Server) {
	esportspb.RegisterAgreementServiceServer(srv, s)
}

func (s *AgreementService) RegisterHTTP(srv *http.Server) {
	esportspb.RegisterAgreementServiceHTTPServer(srv, s)
}

func (s *AgreementService) RegisterConnect(srv *connecttransport.Server) {
	esportspb.RegisterAgreementServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *AgreementService) GetAgreementCurrent(ctx context.Context, _ *esportspb.GetAgreementCurrentRequest) (*esportspb.GetAgreementCurrentResponse, error) {
	current, err := s.agreementUC.Current(ctx)
	if err != nil {
		return nil, err
	}
	return &esportspb.GetAgreementCurrentResponse{
		Docs: utils.SliceMap(current, mapCurrentPB),
	}, nil
}

func (s *AgreementService) RecordAgreementConsent(ctx context.Context, in *esportspb.RecordAgreementConsentRequest) (*esportspb.RecordAgreementConsentResponse, error) {
	if err := s.agreementUC.Record(ctx, LoginMethodFromPB(in.GetLoginMethod())); err != nil {
		return nil, err
	}
	return &esportspb.RecordAgreementConsentResponse{}, nil
}

func DocPB(doc string) esportspb.AgreementDoc {
	// The doc/login-method vocabulary is this package's; the admin face
	// consumes these folds instead of re-spelling the wire values.
	if doc == docPrivacy {
		return esportspb.AgreementDoc_AGREEMENT_DOC_PRIVACY
	}
	return esportspb.AgreementDoc_AGREEMENT_DOC_USER
}

func DocFromPB(doc esportspb.AgreementDoc) string {
	if doc == esportspb.AgreementDoc_AGREEMENT_DOC_PRIVACY {
		return docPrivacy
	}
	return docUser
}

func LoginMethodFromPB(method esportspb.AgreementLoginMethod) string {
	switch method {
	case esportspb.AgreementLoginMethod_AGREEMENT_LOGIN_METHOD_PASSWORD:
		return loginPassword
	case esportspb.AgreementLoginMethod_AGREEMENT_LOGIN_METHOD_WECHAT:
		return loginWechat
	default:
		return loginSMSCode
	}
}

func LoginMethodPB(method string) esportspb.AgreementLoginMethod {
	switch method {
	case loginPassword:
		return esportspb.AgreementLoginMethod_AGREEMENT_LOGIN_METHOD_PASSWORD
	case loginWechat:
		return esportspb.AgreementLoginMethod_AGREEMENT_LOGIN_METHOD_WECHAT
	default:
		return esportspb.AgreementLoginMethod_AGREEMENT_LOGIN_METHOD_SMS_CODE
	}
}

// Private -------------------------------------------------------------------------------------------------------------

func mapCurrentPB(v *AgreementVersion) *esportspb.AgreementCurrent {
	return &esportspb.AgreementCurrent{
		Doc:         DocPB(v.Doc),
		No:          utils.StringW(v.No),
		EffectiveAt: utils.ToTimestamp(&v.EffectiveAt),
	}
}
