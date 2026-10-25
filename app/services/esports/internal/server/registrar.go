package server

import (
	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	"cyber-ecosystem/shared-go/kratos/transport/connect"

	"cyber-ecosystem/app/services/esports/internal/module/entitle"
	"cyber-ecosystem/app/services/esports/internal/module/entitleadmin"
	"cyber-ecosystem/app/services/esports/internal/module/game"
	"cyber-ecosystem/app/services/esports/internal/module/match"
	"cyber-ecosystem/app/services/esports/internal/module/member"
	"cyber-ecosystem/app/services/esports/internal/module/pay"
	"cyber-ecosystem/app/services/esports/internal/module/profile"
	"cyber-ecosystem/app/services/esports/internal/module/test"
)

type Registrar interface {
	RegisterGRPC(*grpc.Server)
	RegisterHTTP(*http.Server)
	RegisterConnect(*connect.Server)
}

func NewRegistrarList(s1 *profile.ProfileService, s2 *game.GameService, s3 *member.MemberService, s4 *test.TestService, s5 *match.MatchService, s6 *pay.SkuService, s7 *pay.OrderService, s8 *entitle.EntitlementService, s9 *entitleadmin.EntitleAdminService) []Registrar {
	return []Registrar{s1, s2, s3, s4, s5, s6, s7, s8, s9}
}
