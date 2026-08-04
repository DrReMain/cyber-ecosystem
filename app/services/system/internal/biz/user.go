package biz

import (
	"context"
	"log/slog"
	"time"

	"cyber-ecosystem/shared-go/utils"
)

// DO ------------------------------------------------------------------------------------------------------------------

type User struct {
	ID           string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	TenantID     string
	DeptID       *string
	Email        string
	PasswordHash string // stays in biz — never reaches proto, so login verifies without leaking it
}

// Port ----------------------------------------------------------------------------------------------------------------

type UserRP interface {
	Create(ctx context.Context, u *User) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
}

// UC ------------------------------------------------------------------------------------------------------------------

type UserUC struct {
	UC
	userRP UserRP
}

func NewUserUC(logger *slog.Logger, tm Transaction, userRP UserRP) *UserUC {
	return &UserUC{
		UC:     UC{log: logger.With("module", "biz/user"), tm: tm},
		userRP: userRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *UserUC) Create(ctx context.Context, email, password string, deptID *string) (out *User, err error) {
	hash, err := utils.Hash(password)
	if err != nil {
		return
	}
	err = uc.tm.InTx(ctx, func(ctx context.Context) error {
		out, err = uc.userRP.Create(ctx, &User{TenantID: defaultTenant, DeptID: deptID, Email: email, PasswordHash: hash})
		return err
	})
	return
}

func (uc *UserUC) Get(ctx context.Context, id string) (*User, error) {
	return uc.userRP.FindByID(ctx, id)
}
