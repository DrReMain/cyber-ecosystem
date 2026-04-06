package biz

import (
	"context"

	"github.com/google/wire"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/apps/singleton/services/backend/internal/pkg/condition"
	"cyber-ecosystem/apps/singleton/services/backend/internal/pkg/datascope"
)

type Transaction interface {
	InTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type UC struct {
	log *log.Helper
	tm  Transaction
}

// ScopeCacheInvalidator abstracts scope cache invalidation for cross-domain coordination.
type ScopeCacheInvalidator interface {
	InvalidateUser(ctx context.Context, userID string) error
}

// NewScopeCacheInvalidator provides ScopeCacheInvalidator from DataScopeUC.
func NewScopeCacheInvalidator(uc *DataScopeUC) ScopeCacheInvalidator {
	return uc
}

var ProviderSet = wire.NewSet(
	NewUserUC,
	NewRoleUC,
	NewDepartmentUC,
	NewUserAttributeUC,
	NewDepartmentBindingUC,
	NewPolicyUC,
	NewConditionUC,
	NewDataScopeUC,
	NewResourceUC,
	NewWorkReportUC,
	NewAccountUC,
	NewSessionValidator,
	NewAuthorizer,
	NewConditionChecker,
	NewScopeResolver,
	NewScopeCacheInvalidator,
	NewRoleCascadeRegistry,
	condition.NewBuiltinConditionRegistry,
	datascope.NewBuiltinScopePluginRegistry,
)
