package data

import (
	"context"

	"github.com/DrReMain/cyber-ecosystem/examples/template3/internal/biz"
	"github.com/DrReMain/cyber-ecosystem/examples/template3/internal/data/ent/user"

	"github.com/go-kratos/kratos/v2/log"
)

type helloRP struct {
	log  *log.Helper
	data *Data
}

func NewHelloRP(logger log.Logger, data *Data) biz.HelloRP {
	return &helloRP{
		log:  log.NewHelper(log.With(logger, "module", "data/hello")),
		data: data,
	}
}

// ---------------------------------------------------------------------------------------------------------------------

func (rp *helloRP) SayHello(ctx context.Context, entity *biz.HelloEntity) (*biz.HelloEntity, error) {
	var result *biz.HelloEntity
	err := rp.data.InTx(ctx, func(ctx context.Context) error {
		client := rp.data.getClient(ctx)
		if en, err := client.User.Query().
			Where(user.NameEQ(*entity.Name)).
			Only(ctx); err != nil {
			return HandleError(err)
		} else {
			result = &biz.HelloEntity{
				Name: &en.Name,
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
