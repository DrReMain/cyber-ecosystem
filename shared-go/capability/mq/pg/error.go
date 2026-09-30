package pg

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	"cyber-ecosystem/shared-go/capability/mq"
)

// mapError translates a pgx/PG error into an mq sentinel. Consumer handler business
// errors do not go through here.
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
