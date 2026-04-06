package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/contracts/go/common"
	"cyber-ecosystem/shared-go/utils"
)

// Model ---------------------------------------------------------------------------------------------------------------

type User struct {
	ID             *string
	CreatedAt      *time.Time
	UpdatedAt      *time.Time
	PasswordPlain  *string
	PasswordCipher *string
	Email          *string
}

type UserQueryIn struct {
	*common.PageRequest
	OrderBy []*utils.OrderBy
	ID      *string
	Email   *string
}

type UserQueryOut struct {
	*common.PageResponse
	List []*User
}

// Port ----------------------------------------------------------------------------------------------------------------

type UserRP interface {
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, fieldsMask []string, user *User) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*User, error)
	Query(ctx context.Context, in *UserQueryIn) (*UserQueryOut, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
}

// UC ------------------------------------------------------------------------------------------------------------------

type UserUC struct {
	UC
	userRP   UserRP
	policyUC *PolicyUC
}

func NewUserUC(logger log.Logger, tm Transaction, userRP UserRP, policyUC *PolicyUC) *UserUC {
	return &UserUC{
		UC: UC{
			log: log.NewHelper(log.With(logger, "module", "biz/uc_user")),
			tm:  tm,
		},
		userRP:   userRP,
		policyUC: policyUC,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *UserUC) Create(ctx context.Context, user *User) error {
	user.PasswordCipher = utils.Ptr(utils.EncryptGenerate(*user.PasswordPlain))
	return uc.tm.InTx(ctx, func(ctx context.Context) error {
		return uc.userRP.Create(ctx, user)
	})
}

func (uc *UserUC) Update(ctx context.Context, fieldsMask []string, user *User) error {
	user.PasswordCipher = utils.Ptr(utils.EncryptGenerate(*user.PasswordPlain))
	return uc.tm.InTx(ctx, func(ctx context.Context) error {
		return uc.userRP.Update(ctx, fieldsMask, user)
	})
}

func (uc *UserUC) Delete(ctx context.Context, id string) error {
	return uc.tm.InTx(ctx, func(ctx context.Context) error {
		if err := uc.userRP.Delete(ctx, id); err != nil {
			return err
		}
		return uc.policyUC.CascadeUserDelete(ctx, id)
	})
}

func (uc *UserUC) Get(ctx context.Context, id string) (*User, error) {
	return uc.userRP.Get(ctx, id)
}

func (uc *UserUC) Query(ctx context.Context, in *UserQueryIn) (*UserQueryOut, error) {
	return uc.userRP.Query(ctx, in)
}
