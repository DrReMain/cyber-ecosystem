package mq

// MQ is the messaging capability root. Backend constructors (nats.New /
// pg.New) populate both fields — a nil field nil-derefs deep in a request
// path, not at startup; the underlying connection lifecycle belongs to the
// backend provider's wire-registered cleanup, not to this container.
type MQ struct {
	Publisher Publisher
	Consumer  Consumer
}
