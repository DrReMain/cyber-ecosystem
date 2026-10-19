package server

import (
	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	"cyber-ecosystem/shared-go/kratos/transport/connect"

	"cyber-ecosystem/app/services/esports/internal/module/game"
	"cyber-ecosystem/app/services/esports/internal/module/profile"
)

type Registrar interface {
	RegisterGRPC(*grpc.Server)
	RegisterHTTP(*http.Server)
	RegisterConnect(*connect.Server)
}

func NewRegistrarList(s1 *profile.ProfileService, s2 *game.GameService) []Registrar {
	return []Registrar{s1, s2}
}
