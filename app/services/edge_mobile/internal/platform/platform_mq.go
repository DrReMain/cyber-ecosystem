package platform

import (
	"fmt"

	"cyber-ecosystem/shared-go/mq"
	mqnats "cyber-ecosystem/shared-go/mq/nats"

	"cyber-ecosystem/app/services/edge_mobile/internal/conf"
)

// NewMQ builds the NATS-backed MQ container for edge_mobile. The (T, func(), error)
// shape lets wire chain the cleanup (NATS conn drain) for graceful shutdown and
// partial-injection rollback.
func NewMQ(c *conf.Data) (*mq.MQ, func(), error) {
	mc := c.GetMq()
	if mc == nil {
		return nil, nil, fmt.Errorf("mq config is required")
	}
	cfg := toMQConfig(mc.GetNats())
	h, closeFn, err := mqnats.NewClient(cfg)
	if err != nil {
		return nil, nil, err
	}
	return mqnats.New(h), closeFn, nil
}

func toMQConfig(n *conf.Data_MQ_NATS) *mqnats.Config {
	cfg := &mqnats.Config{
		Endpoint:      n.GetEndpoint(),
		Creds:         n.GetCreds(),
		MaxBytes:      n.GetMaxBytes(),
		MaxRetries:    int(n.GetMaxRetries()),
		MaxAckPending: int(n.GetMaxAckPending()),
		DLQMaxBytes:   n.GetDlqMaxBytes(),
	}
	// Duration fields are *durationpb.Duration pointers (nil = unset, so the
	// backend applies its default); scalars are read via Get* and fall through to
	// the backend's *OrDefault helpers on a zero value.
	if n.MaxAge != nil {
		cfg.MaxAge = n.GetMaxAge().AsDuration()
	}
	if n.AckWait != nil {
		cfg.AckWait = n.GetAckWait().AsDuration()
	}
	if n.DlqMaxAge != nil {
		cfg.DLQMaxAge = n.GetDlqMaxAge().AsDuration()
	}
	if n.NakBackoffStep != nil {
		cfg.NakBackoffStep = n.GetNakBackoffStep().AsDuration()
	}
	return cfg
}
