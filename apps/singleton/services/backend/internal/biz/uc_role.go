package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/contracts/go/common"
	"cyber-ecosystem/shared-go/utils"
)

// Model ---------------------------------------------------------------------------------------------------------------

type Role struct {
	ID          *string
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
	Name        *string
	Code        *string
	Description *string
	Status      *uint8
	Sort        *string
}

type RoleQueryIn struct {
	*common.PageRequest
	OrderBy []*utils.OrderBy
	Name    *string
	Code    *string
}

type RoleQueryOut struct {
	*common.PageResponse
	List []*Role
}

// Port ----------------------------------------------------------------------------------------------------------------

type RoleRP interface {
	Create(ctx context.Context, role *Role) error
	Update(ctx context.Context, fieldsMask []string, role *Role) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*Role, error)
	Query(ctx context.Context, in *RoleQueryIn) (*RoleQueryOut, error)
	FindByCodes(ctx context.Context, codes []string) ([]*Role, error)
}

// UC ------------------------------------------------------------------------------------------------------------------

type RoleUC struct {
	UC
	cascade *RoleCascadeRegistry
	roleRP  RoleRP
}

func NewRoleUC(logger log.Logger, tm Transaction, cascade *RoleCascadeRegistry, roleRP RoleRP) *RoleUC {
	return &RoleUC{
		UC: UC{
			log: log.NewHelper(log.With(logger, "module", "biz/uc_role")),
			tm:  tm,
		},
		cascade: cascade,
		roleRP:  roleRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *RoleUC) Create(ctx context.Context, role *Role) error {
	return uc.tm.InTx(ctx, func(ctx context.Context) error {
		return uc.roleRP.Create(ctx, role)
	})
}

func (uc *RoleUC) Update(ctx context.Context, fieldsMask []string, role *Role) error {
	return uc.tm.InTx(ctx, func(ctx context.Context) error {
		return uc.roleRP.Update(ctx, fieldsMask, role)
	})
}

func (uc *RoleUC) Delete(ctx context.Context, id string) error {
	return uc.tm.InTx(ctx, func(ctx context.Context) error {
		role, err := uc.roleRP.Get(ctx, id)
		if err != nil {
			return err
		}
		roleCode := *role.Code
		if err := uc.cascade.CascadeRoleDelete(ctx, roleCode); err != nil {
			return err
		}
		return uc.roleRP.Delete(ctx, id)
	})
}

func (uc *RoleUC) Get(ctx context.Context, id string) (*Role, error) {
	return uc.roleRP.Get(ctx, id)
}

func (uc *RoleUC) Query(ctx context.Context, in *RoleQueryIn) (*RoleQueryOut, error) {
	return uc.roleRP.Query(ctx, in)
}

// ---------------------------------------------------------------------------------------------------------------------

// RoleCascadeHandler is implemented by any UC that needs cleanup on role deletion.
type RoleCascadeHandler interface {
	CascadeRoleDelete(ctx context.Context, roleCode string) error
}

// RoleCascadeRegistry coordinates cascade cleanup across subsystems on role deletion.
type RoleCascadeRegistry struct {
	handlers []RoleCascadeHandler
}

func NewRoleCascadeRegistry(dataScopeUC *DataScopeUC, condUC *ConditionUC, policyUC *PolicyUC) *RoleCascadeRegistry {
	return &RoleCascadeRegistry{
		handlers: []RoleCascadeHandler{dataScopeUC, condUC, policyUC},
	}
}

func (r *RoleCascadeRegistry) CascadeRoleDelete(ctx context.Context, roleCode string) error {
	for _, h := range r.handlers {
		if err := h.CascadeRoleDelete(ctx, roleCode); err != nil {
			return err
		}
	}
	return nil
}
