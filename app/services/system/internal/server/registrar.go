package server

import (
	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	"cyber-ecosystem/shared-go/kratos/transport/connect"

	"cyber-ecosystem/app/services/system/internal/module/auth"
	"cyber-ecosystem/app/services/system/internal/module/dept"
	"cyber-ecosystem/app/services/system/internal/module/item"
	"cyber-ecosystem/app/services/system/internal/module/resource"
	"cyber-ecosystem/app/services/system/internal/module/user"
)

type Registrar interface {
	RegisterGRPC(*grpc.Server)
	RegisterHTTP(*http.Server)
	RegisterConnect(*connect.Server)
}

func NewRegistrarList(
	s1 *auth.AuthService,
	s2 *dept.DeptService,
	s3 *item.ItemService,
	s4 *resource.ResourceService,
	s5 *user.UserService,
) []Registrar {
	return []Registrar{s1, s2, s3, s4, s5}
}
