package data

import (
	"context"

	"github.com/DrReMain/cyber-ecosystem/kratos/system-service/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type userRP struct {
	data *Data
	log  *log.Helper
}

func NewUserRP(data *Data, logger log.Logger) biz.UserRP {
	return &userRP{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (rp *userRP) Create(ctx context.Context) error {
	if err := rp.data.db.User.Create().Exec(ctx); err != nil {
		return err
	}
	return nil
}
