package service

import (
	"github.com/google/wire"

	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	"cyber-ecosystem/shared-go/kratos/transport/connect"
)

type Registrar interface {
	RegisterGRPC(*grpc.Server)
	RegisterHTTP(*http.Server)
	RegisterConnect(*connect.Server)
}

var ProviderSet = wire.NewSet(
	NewRegistrarList,
	NewResourceService,
	NewUserService,
	NewAccountAuthService,
	NewAccountDetailService,
	NewRoleService,
	NewRoleBindingService,
	NewPermissionBindingService,
	NewDataScopeService,
	NewDepartmentService,
	NewDepartmentBindingService,
	NewUserAttributeService,
	NewConditionService,
	NewWorkReportService,
)

func NewRegistrarList(
	s1 *ResourceService,
	s2 *UserService,
	s3 *AccountAuthService,
	s4 *AccountDetailService,
	s5 *RoleService,
	s6 *RoleBindingService,
	s7 *PermissionBindingService,
	s8 *DataScopeService,
	s9 *DepartmentService,
	s10 *DepartmentBindingService,
	s11 *UserAttributeService,
	s12 *ConditionService,
	s13 *WorkReportService,
) []Registrar {
	return []Registrar{
		s1,
		s2,
		s3,
		s4,
		s5,
		s6,
		s7,
		s8,
		s9,
		s10,
		s11,
		s12,
		s13,
	}
}
