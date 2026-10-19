package bootstrap

import (
	"context"
	"errors"
)

type Lifecycle struct {
	onStart []func(context.Context) error
	onStop  []func(context.Context) error
}

func NewLifecycle() *Lifecycle {
	return &Lifecycle{}
}

func (l *Lifecycle) OnStart(fn func(context.Context) error) {
	l.onStart = append(l.onStart, fn)
}

func (l *Lifecycle) OnStop(fn func(context.Context) error) {
	l.onStop = append(l.onStop, fn)
}

func (l *Lifecycle) Start(ctx context.Context) error {
	for _, fn := range l.onStart {
		if err := fn(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (l *Lifecycle) Stop(ctx context.Context) error {
	var errs []error
	for i := len(l.onStop) - 1; i >= 0; i-- {
		if err := l.onStop[i](ctx); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
