package agentconfigadmin

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	agentpb "cyber-ecosystem/gen/go/cyber/agent/v1"

	"cyber-ecosystem/app/services/agent/internal/module/agentconfig"
)

// Struct --------------------------------------------------------------------------------------------------------------

type AgentConfigAdminService struct {
	agentpb.UnimplementedAgentConfigAdminServiceServer

	log           *slog.Logger
	agentConfigUC *agentconfig.AgentConfigUC
}

func NewAgentConfigAdminService(logger *slog.Logger, agentConfigUC *agentconfig.AgentConfigUC) *AgentConfigAdminService {
	return &AgentConfigAdminService{
		log:           logger.With("module", "module/agentconfigadmin_service"),
		agentConfigUC: agentConfigUC,
	}
}

func (s *AgentConfigAdminService) RegisterGRPC(srv *grpc.Server) {
	agentpb.RegisterAgentConfigAdminServiceServer(srv, s)
}

func (s *AgentConfigAdminService) RegisterHTTP(srv *http.Server) {
	agentpb.RegisterAgentConfigAdminServiceHTTPServer(srv, s)
}

func (s *AgentConfigAdminService) RegisterConnect(srv *connecttransport.Server) {
	agentpb.RegisterAgentConfigAdminServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *AgentConfigAdminService) ListAgentConfigs(ctx context.Context, in *agentpb.ListAgentConfigsRequest) (*agentpb.ListAgentConfigsResponse, error) {
	out, err := s.agentConfigUC.List(ctx, &agentconfig.ListIn{PageRequest: in.GetPage()})
	if err != nil {
		return nil, err
	}
	list := make([]*agentpb.AgentConfigView, 0, len(out.List))
	for _, it := range out.List {
		v := &agentpb.AgentConfigView{
			Id:        utils.StringW(it.Config.ID),
			CreatedAt: utils.ToTimestamp(&it.Config.CreatedAt),
			UpdatedAt: utils.ToTimestamp(&it.Config.UpdatedAt),
			UserId:    utils.StringW(it.Config.UserID),
			BaseUrl:   utils.StringW(it.Config.BaseURL),
			ApiKeySet: utils.BoolW(it.Config.APIKeySet),
		}
		if it.User != nil {
			v.UserEmail = utils.Wrap(it.User.Email, utils.StringW)
			v.UserAvatar = utils.Wrap(it.User.Avatar, utils.StringW)
		}
		list = append(list, v)
	}
	return &agentpb.ListAgentConfigsResponse{
		Page: out.PageResponse,
		List: list,
	}, nil
}
