package pg

import "cyber-ecosystem/shared-go/mq"

// New assembles the PG-backed MQ container from a connected handle.
func New(h *handle) *mq.MQ {
	return &mq.MQ{Publisher: newPublisher(h), Consumer: newConsumer(h)}
}
