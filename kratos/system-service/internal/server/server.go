package server

import (
	"github.com/DrReMain/cyber-ecosystem/kratos/system-service/internal/service"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer, NewRegistrarList)

func NewRegistrarList(
	s1 *service.UserService,
) []service.Registrar {
	return []service.Registrar{
		s1,
	}
}
