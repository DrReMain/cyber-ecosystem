package nats

import (
	"context"
	"errors"
	"fmt"

	natsclient "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"cyber-ecosystem/shared-go/capability/mq"
)

// mapError translates a nats/jetstream error into an mq sentinel. Consumer
// handler BUSINESS errors never reach here (they are retried/DLQ'd upstream).
func mapError(err error, op string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, natsclient.ErrTimeout) {
		return fmt.Errorf("%s: %w", op, mq.ErrTimeout)
	}
	// JetStream API errors carry an HTTP status code; 504 is a server-side timeout.
	var apiErr *jetstream.APIError
	if errors.As(err, &apiErr) && apiErr.Code == 504 {
		return fmt.Errorf("%s: %w", op, mq.ErrTimeout)
	}
	if errors.Is(err, context.Canceled) {
		return fmt.Errorf("%s: %w", op, err)
	}
	if errors.Is(err, natsclient.ErrNoResponders) || errors.Is(err, natsclient.ErrConnectionClosed) {
		return fmt.Errorf("%s: %w", op, mq.ErrUnavailable)
	}
	return fmt.Errorf("%s: %w", op, err)
}
