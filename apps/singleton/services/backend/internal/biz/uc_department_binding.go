package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// Port ----------------------------------------------------------------------------------------------------------------

// DepartmentBindingRP abstracts user-department binding persistence.
type DepartmentBindingRP interface {
	SetUserDepartments(ctx context.Context, userID string, departmentIDs []string) error
	ListUserDepartments(ctx context.Context, userID string) ([]*Department, error)
	ListDepartmentUsers(ctx context.Context, departmentID string) ([]string, error)
}

// UC ------------------------------------------------------------------------------------------------------------------

// DepartmentBindingUC handles user-department binding with scope cache invalidation.
type DepartmentBindingUC struct {
	UC
	deptBindingRP DepartmentBindingRP
	deptRP        DepartmentRP
	invalidator   ScopeCacheInvalidator
}

// NewDepartmentBindingUC creates a new DepartmentBindingUC.
func NewDepartmentBindingUC(
	logger log.Logger,
	tm Transaction,
	deptBindingRP DepartmentBindingRP,
	deptRP DepartmentRP,
	invalidator ScopeCacheInvalidator,
) *DepartmentBindingUC {
	return &DepartmentBindingUC{
		UC: UC{
			log: log.NewHelper(log.With(logger, "module", "biz/uc_department_binding")),
			tm:  tm,
		},
		deptBindingRP: deptBindingRP,
		deptRP:        deptRP,
		invalidator:   invalidator,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

// SetUserDepartments sets the departments for a user, validating each department exists.
func (uc *DepartmentBindingUC) SetUserDepartments(ctx context.Context, userID string, departmentIDs []string) error {
	return uc.tm.InTx(ctx, func(txCtx context.Context) error {
		for _, deptID := range departmentIDs {
			if _, err := uc.deptRP.Get(txCtx, deptID); err != nil {
				return err
			}
		}
		if err := uc.deptBindingRP.SetUserDepartments(txCtx, userID, departmentIDs); err != nil {
			return err
		}
		if err := uc.invalidator.InvalidateUser(txCtx, userID); err != nil {
			uc.log.Warnf("failed to invalidate scope cache for user %s: %v", userID, err)
		}
		return nil
	})
}

// ListUserDepartments returns the departments bound to a user.
func (uc *DepartmentBindingUC) ListUserDepartments(ctx context.Context, userID string) ([]*Department, error) {
	return uc.deptBindingRP.ListUserDepartments(ctx, userID)
}

// ListDepartmentUsers returns user IDs bound to a department.
func (uc *DepartmentBindingUC) ListDepartmentUsers(ctx context.Context, departmentID string) ([]string, error) {
	return uc.deptBindingRP.ListDepartmentUsers(ctx, departmentID)
}
