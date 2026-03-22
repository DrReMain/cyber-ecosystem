package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

type HelloEntity struct {
	Name *string
}

type HelloRP interface {
	SayHello(ctx context.Context, entity *HelloEntity) (*HelloEntity, error)
}

type HelloUC struct {
	log *log.Helper
	tm  Transaction

	helloRP HelloRP
}

func NewHelloUC(logger log.Logger, tm Transaction, helloRP HelloRP) *HelloUC {
	return &HelloUC{
		log:     log.NewHelper(log.With(logger, "module", "biz/hello")),
		tm:      tm,
		helloRP: helloRP,
	}
}

// ---------------------------------------------------------------------------------------------------------------------

func (uc *HelloUC) SayHello(ctx context.Context, entity *HelloEntity) (*HelloEntity, error) {
	var result *HelloEntity
	err := uc.tm.InTx(ctx, func(ctx context.Context) error {
		if en, err := uc.helloRP.SayHello(ctx, entity); err != nil {
			return err
		} else {
			result = en
		}
		return nil
	})

	return result, err
}
