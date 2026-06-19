package nats

import "cyber-ecosystem/shared-go/mq"

// New assembles an *mq.MQ from a connected handle.
func New(h *handle) *mq.MQ {
	return &mq.MQ{Publisher: newPublisher(h), Consumer: newConsumer(h)}
}
