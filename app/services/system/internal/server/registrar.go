package server

import (
	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	"cyber-ecosystem/shared-go/kratos/transport/connect"

	"cyber-ecosystem/app/services/system/internal/module/audit"
	"cyber-ecosystem/app/services/system/internal/module/auth"
	"cyber-ecosystem/app/services/system/internal/module/dept"
	"cyber-ecosystem/app/services/system/internal/module/file"
	"cyber-ecosystem/app/services/system/internal/module/filepresign"
	"cyber-ecosystem/app/services/system/internal/module/fileproxy"
	"cyber-ecosystem/app/services/system/internal/module/introspect"
	"cyber-ecosystem/app/services/system/internal/module/policy"
	"cyber-ecosystem/app/services/system/internal/module/resource"
	"cyber-ecosystem/app/services/system/internal/module/role"
	"cyber-ecosystem/app/services/system/internal/module/transfer"
	"cyber-ecosystem/app/services/system/internal/module/user"
)

type Registrar interface {
	RegisterGRPC(*grpc.Server)
	RegisterHTTP(*http.Server)
	RegisterConnect(*connect.Server)
}

func NewRegistrarList(
	s1 *dept.DeptService,
	s2 *resource.ResourceService,
	s3 *user.UserService,
	s4 *transfer.TransferService,
	s5 *auth.AuthService,
	s6 *role.RoleService,
	s7 *policy.PolicyService,
	s8 *audit.AuditService,
	s9 *file.FileService,
	s10 *fileproxy.FileProxyService,
	s11 *filepresign.FilePresignService,
	s12 *introspect.IntrospectService,
) []Registrar {
	return []Registrar{s1, s2, s3, s4, s5, s6, s7, s8, s9, s10, s11, s12}
}
