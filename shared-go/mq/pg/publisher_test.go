package pg

import (
	"context"
	"errors"
	"testing"

	"cyber-ecosystem/shared-go/mq"
)

func TestPublishInsertsAndFansOut(t *testing.T) {
	h, cleanup := newTestMQ(t)
	defer cleanup()
	ctx := context.Background()
	topic := uniqTopic(t, "pub")
	if _, err := h.pool.Exec(ctx, `INSERT INTO subscribers(group_name,topic) VALUES('g',$1) ON CONFLICT DO NOTHING`, topic); err != nil {
		t.Fatalf("seed sub: %v", err)
	}
	id, err := newPublisher(h).Publish(ctx, topic, &mq.Message{Payload: []byte("hi"), Headers: map[string]string{"k": "v"}})
	if err != nil || id == "" {
		t.Fatalf("Publish: id=%q err=%v", id, err)
	}
	var payload []byte
	if err := h.pool.QueryRow(ctx, `SELECT payload FROM messages WHERE id=$1`, id).Scan(&payload); err != nil {
		t.Fatalf("message: %v", err)
	}
	if string(payload) != "hi" {
		t.Fatalf("payload=%q", payload)
	}
	var dcount int
	if err := h.pool.QueryRow(ctx, `SELECT count(*) FROM deliveries WHERE message_id=$1 AND group_name='g'`, id).Scan(&dcount); err != nil {
		t.Fatalf("deliveries: %v", err)
	}
	if dcount != 1 {
		t.Fatalf("deliveries=%d, want 1", dcount)
	}
}

func TestPublishInvalidTopic(t *testing.T) {
	h, cleanup := newTestMQ(t)
	defer cleanup()
	_, err := newPublisher(h).Publish(context.Background(), "", &mq.Message{Payload: []byte("x")})
	if !errors.Is(err, mq.ErrInvalidArgument) {
		t.Fatalf("empty topic: got %v, want ErrInvalidArgument", err)
	}
}
