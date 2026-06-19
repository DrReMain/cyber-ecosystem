package pg

import (
	"context"
	"encoding/json"
	"fmt"

	"cyber-ecosystem/shared-go/mq"
)

type publisher struct{ h *handle }

func newPublisher(h *handle) mq.Publisher { return &publisher{h: h} }

// Publish inserts the message then fans out a delivery row to every subscriber
// group of the topic (ON CONFLICT dedupes vs a concurrent subscribe-backfill).
// Returns the message id (opaque bigserial) as the message ID.
func (p *publisher) Publish(ctx context.Context, topic string, msg *mq.Message) (string, error) {
	if err := mq.ValidateTopic(topic); err != nil {
		return "", err
	}
	headers, _ := json.Marshal(msg.Headers)
	tx, err := p.h.pool.Begin(ctx)
	if err != nil {
		return "", mapError(err, "begin")
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var id int64
	if err := tx.QueryRow(ctx,
		`INSERT INTO messages(topic,payload,headers) VALUES($1,$2,$3) RETURNING id`,
		topic, msg.Payload, headers).Scan(&id); err != nil {
		return "", mapError(err, "insert message")
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO deliveries(group_name,topic,message_id)
		 SELECT group_name, $1, $2 FROM subscribers WHERE topic=$1
		 ON CONFLICT (group_name, message_id) DO NOTHING`,
		topic, id); err != nil {
		return "", mapError(err, "fanout")
	}
	if err := tx.Commit(ctx); err != nil {
		return "", mapError(err, "commit")
	}
	return fmt.Sprintf("%d", id), nil
}
