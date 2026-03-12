package biz

import (
	"context"

	v1 "github.com/DrReMain/cyber-ecosystem/kratos/system-service/gen/v1"

	"github.com/go-kratos/kratos/v2/log"
)

var (
	ErrUserNotFound = v1.ErrorUserNotFound("", "").WithMetadata(map[string]string{})
)

type UserRP interface {
	Create(context.Context) error
}

type UserUC struct {
	userRP UserRP
	log    *log.Helper
}

func NewUserUC(userRP UserRP, logger log.Logger) *UserUC {
	return &UserUC{userRP: userRP, log: log.NewHelper(logger)}
}

func (uc *UserUC) CreateUser(ctx context.Context) error {
	return uc.userRP.Create(ctx)
}
