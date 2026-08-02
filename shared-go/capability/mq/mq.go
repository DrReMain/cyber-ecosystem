package mq

// MQ holds the messaging capability. Backends populate both fields.
type MQ struct {
	Publisher Publisher
	Consumer  Consumer
}
