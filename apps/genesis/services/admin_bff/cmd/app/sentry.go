package main

import (
	"time"

	"github.com/getsentry/sentry-go"

	"cyber-ecosystem/apps/genesis/services/admin_bff/internal/conf"
)

func initSentry(c *conf.ErrorReport) (func(), error) {
	if c == nil || !c.Enabled || c.Dsn == "" {
		return func() {}, nil
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              c.Dsn,
		Environment:      c.Environment,
		SampleRate:       float64(c.SampleRate),
		AttachStacktrace: true,
	})
	if err != nil {
		return nil, err
	}

	return func() {
		sentry.Flush(5 * time.Second)
	}, nil
}
