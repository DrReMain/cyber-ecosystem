package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

const Domain = "self"

type PermissionBinding struct {
	Object string
	Effect string
}

// Port ----------------------------------------------------------------------------------------------------------------

type PolicyRP interface {
	Enforce(sub, dom, obj string) (bool, error)
	GetRolesForUser(userID string) []string
	GetUsersForRole(role string) []string
	GetPermissionsForRole(role string) ([][]string, error)
	AddRoleForUser(ctx context.Context, userID, role string) (bool, func(), error)
	RemoveRoleForUser(ctx context.Context, userID, role string) (bool, func(), error)
	AddPermissionForRole(ctx context.Context, role, object, effect string) (bool, func(), error)
	RemovePermissionForRole(ctx context.Context, role, object, effect string) (bool, func(), error)
	RemoveRoleGroupings(ctx context.Context, roleCode string) error
	RemoveUserGroupings(ctx context.Context, userID string) error
	RemoveRolePermissions(ctx context.Context, roleCode string) error
}

// UC ------------------------------------------------------------------------------------------------------------------

type PolicyUC struct {
	UC
	domain   string
	policyRP PolicyRP
	roleRP   RoleRP
	userRP   UserRP
}

func NewPolicyUC(logger log.Logger, tm Transaction, policyRP PolicyRP, roleRP RoleRP, userRP UserRP) *PolicyUC {
	return &PolicyUC{
		UC: UC{
			log: log.NewHelper(log.With(logger, "module", "biz/uc_policy")),
			tm:  tm,
		},
		domain:   Domain,
		policyRP: policyRP,
		roleRP:   roleRP,
		userRP:   userRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

// Enforce checks whether a subject is allowed to perform an operation.
func (uc *PolicyUC) Enforce(ctx context.Context, sub, operation string) (bool, error) {
	return uc.policyRP.Enforce(sub, uc.domain, operation)
}

// GrantRole assigns a role to a user.
func (uc *PolicyUC) GrantRole(ctx context.Context, userID, roleCode string) error {
	var syncFn func()
	err := uc.tm.InTx(ctx, func(txCtx context.Context) error {
		// TODO: 如果传入不存在的role code结果需要进一步DEBUG
		_, err := uc.roleRP.FindByCodes(txCtx, []string{roleCode})
		if err != nil {
			return err
		}
		if _, err := uc.userRP.Get(txCtx, userID); err != nil {
			return err
		}
		_, sync, err := uc.policyRP.AddRoleForUser(txCtx, userID, roleCode)
		if err != nil {
			return err
		}
		syncFn = sync
		return nil
	})
	if err != nil {
		return err
	}
	safeSync(syncFn)
	return nil
}

// RevokeRole removes a role from a user.
func (uc *PolicyUC) RevokeRole(ctx context.Context, userID, roleCode string) error {
	var syncFn func()
	err := uc.tm.InTx(ctx, func(txCtx context.Context) error {
		_, sync, err := uc.policyRP.RemoveRoleForUser(txCtx, userID, roleCode)
		if err != nil {
			return err
		}
		syncFn = sync
		return nil
	})
	if err != nil {
		return err
	}
	safeSync(syncFn)
	return nil
}

// AssignPermission adds a permission policy for a role.
func (uc *PolicyUC) AssignPermission(ctx context.Context, roleCode, object, effect string) error {
	var syncFn func()
	err := uc.tm.InTx(ctx, func(txCtx context.Context) error {
		_, sync, err := uc.policyRP.AddPermissionForRole(txCtx, roleCode, object, effect)
		if err != nil {
			return err
		}
		syncFn = sync
		return nil
	})
	if err != nil {
		return err
	}
	safeSync(syncFn)
	return nil
}

// RemovePermission removes a permission policy from a role.
func (uc *PolicyUC) RemovePermission(ctx context.Context, roleCode, object, effect string) error {
	var syncFn func()
	err := uc.tm.InTx(ctx, func(txCtx context.Context) error {
		_, sync, err := uc.policyRP.RemovePermissionForRole(txCtx, roleCode, object, effect)
		if err != nil {
			return err
		}
		syncFn = sync
		return nil
	})
	if err != nil {
		return err
	}
	safeSync(syncFn)
	return nil
}

// QueryUserRoles returns the roles assigned to a user.
func (uc *PolicyUC) QueryUserRoles(ctx context.Context, userID string) ([]*Role, error) {
	codes := uc.policyRP.GetRolesForUser(userID)
	return uc.roleRP.FindByCodes(ctx, codes)
}

// QueryRolePermissions returns the permissions assigned to a role.
func (uc *PolicyUC) QueryRolePermissions(ctx context.Context, roleCode string) ([]*PermissionBinding, error) {
	policies, err := uc.policyRP.GetPermissionsForRole(roleCode)
	if err != nil {
		return nil, err
	}
	list := make([]*PermissionBinding, 0, len(policies))
	for _, p := range policies {
		// p = [role, domain, object, effect]
		if len(p) >= 4 {
			list = append(list, &PermissionBinding{Object: p[2], Effect: p[3]})
		}
	}
	return list, nil
}

// GetRolesForUser returns the role codes assigned to a user.
func (uc *PolicyUC) GetRolesForUser(userID string) []string {
	return uc.policyRP.GetRolesForUser(userID)
}

// GetUsersForRole returns the user IDs that have the given role.
func (uc *PolicyUC) GetUsersForRole(role string) []string {
	return uc.policyRP.GetUsersForRole(role)
}

// CascadeRoleDelete removes all groupings and policies for a role.
func (uc *PolicyUC) CascadeRoleDelete(ctx context.Context, roleCode string) error {
	if err := uc.policyRP.RemoveRoleGroupings(ctx, roleCode); err != nil {
		return err
	}
	return uc.policyRP.RemoveRolePermissions(ctx, roleCode)
}

// CascadeUserDelete removes all groupings for a user.
func (uc *PolicyUC) CascadeUserDelete(ctx context.Context, userID string) error {
	return uc.policyRP.RemoveUserGroupings(ctx, userID)
}

// ---------------------------------------------------------------------------------------------------------------------

func safeSync(fn func()) {
	if fn == nil {
		return
	}
	defer func() {
		recover()
	}()
	fn()
}
