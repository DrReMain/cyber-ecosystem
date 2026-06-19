package pg

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	"cyber-ecosystem/shared-go/mq"
)

// mapError 把 pgx/PG 错误翻成 mq sentinel。consumer handler 业务错误不进这里。
func mapError(err error, op string) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return fmt.Errorf("%s: %w", op, mq.ErrTimeout)
	case errors.Is(err, context.Canceled):
		return fmt.Errorf("%s: %w", op, err)
	case errors.Is(err, pgconn.ErrConnClosed):
		return fmt.Errorf("%s: %w", op, mq.ErrUnavailable)
	default:
		return fmt.Errorf("%s: %w", op, mq.ErrUnavailable)
	}
}
