package data

import (
	"github.com/google/wire"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/apps/singleton/services/backend/internal/biz"
)

type RP struct {
	log   *log.Helper
	store *Store
}

var ProviderSet = wire.NewSet(
	NewStore,
	NewCache,
	NewEntClient,
	wire.Bind(new(biz.Transaction), new(*Store)),
	// -----------------------------------------------------------------------------------------------------------------
	NewResourceRP,
	NewSessionRP,
	NewPolicyRP,
	NewDataScopeRP,
	NewDataScopeSnapshotRP,
	NewUserRP,
	NewRoleRP,
	NewDepartmentRP,
	NewDepartmentBindingRP,
	NewUserAttributeRP,
	NewConditionRP,
	NewWorkReportRP,
)
