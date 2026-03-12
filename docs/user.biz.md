```go
package biz

import (
	"context"
	"time"

	v1 "github.com/DrReMain/cyber-ecosystem/kratos/system-service/gen/v1"

	"github.com/go-kratos/kratos/v2/log"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound = v1.ErrorUserNotFound("", "").WithMetadata(map[string]string{})
)

type User struct {
	ID        string
	Username  string
	Email     string
	Age       int32
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserQueryOption struct {
	PageNo   int32
	PageSize int32
	Email    string
	Fields   []string
}

type UserRP interface {
	Create(context.Context, *User) (*User, error)
	Update(context.Context, *User, []string) (*User, error)
	DeleteBatch(context.Context, []string) error
	Delete(context.Context, string) error
	Get(context.Context, string) (*User, error)
	Query(context.Context, *UserQueryOption) ([]*User, int64, error)
	List(context.Context, *UserQueryOption) ([]*User, error)
}

type UserUC struct {
	userRP UserRP
	log    *log.Helper
}

func NewUserUC(userRP UserRP, logger log.Logger) *UserUC {
	return &UserUC{userRP: userRP, log: log.NewHelper(logger)}
}

func (uc *UserUC) CreateUser(ctx context.Context, u *User) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u.Password = string(hashedPassword)
	return uc.userRP.Create(ctx, u)
}

func (uc *UserUC) UpdateUser(ctx context.Context, u *User, fields []string) (*User, error) {
	return uc.userRP.Update(ctx, u, fields)
}

func (uc *UserUC) DeleteBatchUser(ctx context.Context, ids []string) error {
	return uc.userRP.DeleteBatch(ctx, ids)
}

func (uc *UserUC) DeleteUser(ctx context.Context, id string) error {
	return uc.userRP.Delete(ctx, id)
}

func (uc *UserUC) GetUser(ctx context.Context, id string) (*User, error) {
	return uc.userRP.Get(ctx, id)
}

func (uc *UserUC) QueryUser(ctx context.Context, opt *UserQueryOption) ([]*User, int64, error) {
	return uc.userRP.Query(ctx, opt)
}

func (uc *UserUC) ListUser(ctx context.Context, opt *UserQueryOption) ([]*User, error) {
	return uc.userRP.List(ctx, opt)
}

```
