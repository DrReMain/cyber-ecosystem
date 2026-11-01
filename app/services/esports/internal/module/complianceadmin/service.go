package complianceadmin

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"

	"cyber-ecosystem/app/services/esports/internal/module/compliance"
)

// Struct --------------------------------------------------------------------------------------------------------------

type ComplianceAdminService struct {
	esportspb.UnimplementedComplianceAdminServiceServer

	log          *slog.Logger
	complianceUC *compliance.ComplianceUC
}

func NewComplianceAdminService(logger *slog.Logger, complianceUC *compliance.ComplianceUC) *ComplianceAdminService {
	return &ComplianceAdminService{
		log:          logger.With("module", "module/complianceadmin_service"),
		complianceUC: complianceUC,
	}
}

func (s *ComplianceAdminService) RegisterGRPC(srv *grpc.Server) {
	esportspb.RegisterComplianceAdminServiceServer(srv, s)
}

func (s *ComplianceAdminService) RegisterHTTP(srv *http.Server) {
	esportspb.RegisterComplianceAdminServiceHTTPServer(srv, s)
}

func (s *ComplianceAdminService) RegisterConnect(srv *connecttransport.Server) {
	esportspb.RegisterComplianceAdminServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *ComplianceAdminService) SetBirthCollection(ctx context.Context, in *esportspb.SetBirthCollectionRequest) (*esportspb.SetBirthCollectionResponse, error) {
	config, err := s.complianceUC.SetBirthCollection(ctx, in.GetCollect())
	if err != nil {
		return nil, err
	}
	return &esportspb.SetBirthCollectionResponse{Config: configPB(config)}, nil
}

func (s *ComplianceAdminService) SetCompliancePrompts(ctx context.Context, in *esportspb.SetCompliancePromptsRequest) (*esportspb.SetCompliancePromptsResponse, error) {
	config, err := s.complianceUC.SetCompliancePrompts(ctx, in.MinorPrompt, in.MissingBirthPrompt)
	if err != nil {
		return nil, err
	}
	return &esportspb.SetCompliancePromptsResponse{Config: configPB(config)}, nil
}

func (s *ComplianceAdminService) SetPayAgeLine(ctx context.Context, in *esportspb.SetPayAgeLineRequest) (*esportspb.SetPayAgeLineResponse, error) {
	config, err := s.complianceUC.SetPayAgeLine(ctx, in.GetAge())
	if err != nil {
		return nil, err
	}
	return &esportspb.SetPayAgeLineResponse{Config: configPB(config)}, nil
}

func (s *ComplianceAdminService) GetComplianceConfig(ctx context.Context, _ *esportspb.GetComplianceConfigRequest) (*esportspb.GetComplianceConfigResponse, error) {
	config, err := s.complianceUC.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &esportspb.GetComplianceConfigResponse{Config: configPB(config)}, nil
}

// Private ------------------------------------------------------------------------------------------------------------

func configPB(config *compliance.ComplianceConfig) *esportspb.ComplianceConfig {
	return &esportspb.ComplianceConfig{
		BirthCollect:       utils.BoolW(config.BirthCollect),
		PayAgeLine:         utils.Int32W(config.PayAgeLine),
		MinorPrompt:        utils.Wrap(config.MinorPrompt, utils.StringW),
		MissingBirthPrompt: utils.Wrap(config.MissingBirthPrompt, utils.StringW),
	}
}
