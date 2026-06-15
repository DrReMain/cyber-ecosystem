package platform

import "context"

type Transaction interface {
	InTx(ctx context.Context, fn func(context.Context) error) error
}
