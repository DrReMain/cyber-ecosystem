package server

import (
	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	"cyber-ecosystem/shared-go/kratos/transport/connect"

	"cyber-ecosystem/app/services/agent/internal/module/agentconfig"
	"cyber-ecosystem/app/services/agent/internal/module/agentconfigadmin"
)

type Registrar interface {
	RegisterGRPC(*grpc.Server)
	RegisterHTTP(*http.Server)
	RegisterConnect(*connect.Server)
}

func NewRegistrarList(s1 *agentconfig.AgentConfigService, s2 *agentconfigadmin.AgentConfigAdminService) []Registrar {
	return []Registrar{s1, s2}
}
