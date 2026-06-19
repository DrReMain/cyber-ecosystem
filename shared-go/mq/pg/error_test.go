package pg

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"cyber-ecosystem/shared-go/mq"
)

func TestMapError(t *testing.T) {
	cases := []struct {
		name string
		in   error
		want error
	}{
		{"ctx deadline", context.DeadlineExceeded, mq.ErrTimeout},
		{"pgx conn closed", pgconn.ErrConnClosed, mq.ErrUnavailable},
		{"unknown", errors.New("boom"), mq.ErrUnavailable},
	}
	for _, c := range cases {
		if !errors.Is(mapError(c.in, "op"), c.want) {
			t.Errorf("%s: mapError(%v) not %v", c.name, c.in, c.want)
		}
	}
	if mapError(nil, "op") != nil {
		t.Error("nil should map to nil")
	}
}
