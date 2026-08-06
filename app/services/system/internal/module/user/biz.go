package user

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"cyber-ecosystem/shared-go/kratos/security"
	"cyber-ecosystem/shared-go/utils"

	commonpb "cyber-ecosystem/gen/go/cyber/shared/common/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"

	"cyber-ecosystem/app/services/system/internal/shared"
)

// DO ------------------------------------------------------------------------------------------------------------------

type User struct {
	ID           string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	TenantID     string
	DeptID       *string
	Email        *string
	PasswordHash *string // biz-only (never in proto); nil on Update = leave unchanged
}

type UserListIn struct {
	*commonpb.PageRequest
	OrderBy []string
	Email   *string
}

type UserListOut struct {
	*commonpb.PageResponse
	List []*User
}

// Port ----------------------------------------------------------------------------------------------------------------

type UserRP interface {
	Create(ctx context.Context, u *User) (*User, error)
	Update(ctx context.Context, fieldsMask []string, u *User) (*User, error)
	Delete(ctx context.Context, id string) (string, error)
	List(ctx context.Context, in *UserListIn) (*UserListOut, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
}

// UC ------------------------------------------------------------------------------------------------------------------

type UserUC struct {
	shared.UC
	userRP UserRP
}

func NewUserUC(logger *slog.Logger, tm shared.Transaction, userRP UserRP) *UserUC {
	return &UserUC{
		UC:     shared.NewUC(logger.With("module", "module/user"), tm),
		userRP: userRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *UserUC) Create(ctx context.Context, email, password *string, deptID *string) (out *User, err error) {
	hash, err := utils.Hash(*password)
	if err != nil {
		return
	}
	err = uc.Tm.InTx(ctx, func(ctx context.Context) error {
		out, err = uc.userRP.Create(ctx, &User{TenantID: shared.DefaultTenant, DeptID: deptID, Email: email, PasswordHash: &hash})
		return err
	})
	return
}

func (uc *UserUC) Update(ctx context.Context, fieldsMask []string, u *User, password *string) (out *User, err error) {
	if password != nil {
		hash, e := utils.Hash(*password)
		if e != nil {
			return nil, e
		}
		u.PasswordHash = &hash
	}
	err = uc.Tm.InTx(ctx, func(ctx context.Context) error {
		out, err = uc.userRP.Update(ctx, fieldsMask, u)
		return err
	})
	return
}

func (uc *UserUC) Delete(ctx context.Context, id string) (out string, err error) {
	err = uc.Tm.InTx(ctx, func(ctx context.Context) error {
		out, err = uc.userRP.Delete(ctx, id)
		return err
	})
	return
}

func (uc *UserUC) List(ctx context.Context, in *UserListIn) (*UserListOut, error) {
	return uc.userRP.List(ctx, in)
}

func (uc *UserUC) Get(ctx context.Context, id string) (*User, error) {
	return uc.userRP.FindByID(ctx, id)
}

func (uc *UserUC) GetCurrentUser(ctx context.Context) (*User, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(errors.New("no subject in context"))
	}
	return uc.userRP.FindByID(ctx, subject.UserID)
}
