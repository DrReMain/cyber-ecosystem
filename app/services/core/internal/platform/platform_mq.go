package platform

import (
	"fmt"

	"cyber-ecosystem/shared-go/kratos/observability"
	"cyber-ecosystem/shared-go/mq"
	mqnats "cyber-ecosystem/shared-go/mq/nats"
	mqpg "cyber-ecosystem/shared-go/mq/pg"

	"cyber-ecosystem/app/services/core/internal/conf"
)

func NewMQ(c *conf.Data) (*mq.MQ, func(), error) {
	mc := c.GetMq()
	if mc == nil {
		return nil, nil, fmt.Errorf("mq config is required")
	}
	if n := mc.GetNats(); n != nil && n.GetEndpoint() != "" {
		h, closeFn, err := mqnats.NewClient(toMQConfig(n))
		if err != nil {
			return nil, nil, err
		}
		return observability.InstrumentMQ(mqnats.New(h)), closeFn, nil
	}
	if p := mc.GetPg(); p != nil && p.GetDsn() != "" {
		h, closeFn, err := mqpg.NewClient(toPGConfig(p))
		if err != nil {
			return nil, nil, err
		}
		return observability.InstrumentMQ(mqpg.New(h)), closeFn, nil
	}
	return nil, nil, fmt.Errorf("mq: configure either nats or pg")
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

func toPGConfig(p *conf.Data_MQ_PG) *mqpg.Config {
	cfg := &mqpg.Config{
		DSN:        p.GetDsn(),
		MaxRetries: int(p.GetMaxRetries()),
		BatchSize:  int(p.GetBatchSize()),
	}
	if p.PollInterval != nil {
		cfg.PollInterval = p.GetPollInterval().AsDuration()
	}
	if p.VisibilityTimeout != nil {
		cfg.VisibilityTimeout = p.GetVisibilityTimeout().AsDuration()
	}
	if p.Retention != nil {
		cfg.Retention = p.GetRetention().AsDuration()
	}
	return cfg
}
