package agentconfig

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	agentpb "cyber-ecosystem/gen/go/cyber/agent/v1"
)

// Struct --------------------------------------------------------------------------------------------------------------

type AgentConfigService struct {
	agentpb.UnimplementedAgentConfigServiceServer

	log           *slog.Logger
	agentConfigUC *AgentConfigUC
}

func NewAgentConfigService(logger *slog.Logger, agentConfigUC *AgentConfigUC) *AgentConfigService {
	return &AgentConfigService{
		log:           logger.With("module", "module/agentconfig_service"),
		agentConfigUC: agentConfigUC,
	}
}

func (s *AgentConfigService) RegisterGRPC(srv *grpc.Server) {
	agentpb.RegisterAgentConfigServiceServer(srv, s)
}

func (s *AgentConfigService) RegisterHTTP(srv *http.Server) {
	agentpb.RegisterAgentConfigServiceHTTPServer(srv, s)
}

func (s *AgentConfigService) RegisterConnect(srv *connecttransport.Server) {
	agentpb.RegisterAgentConfigServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *AgentConfigService) UpdateMyAgentConfig(ctx context.Context, in *agentpb.UpdateMyAgentConfigRequest) (*agentpb.UpdateMyAgentConfigResponse, error) {
	if err := s.agentConfigUC.Update(ctx, &UpdateIn{BaseURL: *in.BaseUrl, APIKey: in.ApiKey}); err != nil {
		return nil, err
	}
	return &agentpb.UpdateMyAgentConfigResponse{}, nil
}

func (s *AgentConfigService) DeleteMyAgentConfig(ctx context.Context, in *agentpb.DeleteMyAgentConfigRequest) (*agentpb.DeleteMyAgentConfigResponse, error) {
	if err := s.agentConfigUC.Delete(ctx); err != nil {
		return nil, err
	}
	return &agentpb.DeleteMyAgentConfigResponse{}, nil
}

func (s *AgentConfigService) GetMyAgentConfig(ctx context.Context, in *agentpb.GetMyAgentConfigRequest) (*agentpb.GetMyAgentConfigResponse, error) {
	mc, err := s.agentConfigUC.Get(ctx)
	if err != nil {
		return nil, err
	}
	if mc == nil {
		return &agentpb.GetMyAgentConfigResponse{}, nil
	}
	return &agentpb.GetMyAgentConfigResponse{AgentConfig: mapAgentConfigPB(mc)}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mapAgentConfigPB(mc *AgentConfig) *agentpb.AgentConfig {
	return &agentpb.AgentConfig{
		Id:        utils.StringW(mc.ID),
		CreatedAt: utils.ToTimestamp(&mc.CreatedAt),
		UpdatedAt: utils.ToTimestamp(&mc.UpdatedAt),
		UserId:    utils.StringW(mc.UserID),
		BaseUrl:   utils.StringW(mc.BaseURL),
		ApiKeySet: utils.BoolW(mc.APIKeySet),
	}
}
