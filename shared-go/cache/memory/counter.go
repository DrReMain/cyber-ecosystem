package memory

import (
	"context"
	"strconv"
	"time"

	"cyber-ecosystem/shared-go/cache"
)

type counter struct {
	m *Memory
}

func (c *counter) Incr(ctx context.Context, key string, delta int64) (int64, error) {
	opStart := time.Now()
	if err := cache.ValidateKey(key); err != nil {
		c.m.logOperation(ctx, "incrby", "key", key, "delta", delta, "latency", time.Since(opStart), "error", err)
		return 0, err
	}

	var newVal int64
	var err error

	c.m.traceOperationWithArgs(ctx, "INCR", &cache.OperationArgs{Key: key, DeltaInt: delta}, func(ctx context.Context) error {
		c.m.mu.Lock()
		defer c.m.mu.Unlock()

		expiresAt := time.Time{}
		e, ok := c.m.data[key]
		if !ok || e == nil || e.isExpired() {
			newVal = delta
		} else {
			expiresAt = e.expiresAt
			v, parseErr := strconv.ParseInt(string(e.value), 10, 64)
			if parseErr != nil {
				err = parseErr
				return nil
			}
			newVal = v + delta
		}

		c.m.data[key] = &entry{
			value:     []byte(strconv.FormatInt(newVal, 10)),
			expiresAt: expiresAt,
		}
		return nil
	})

	c.m.logOperation(ctx, "incrby", "key", key, "delta", delta, "newVal", newVal, "latency", time.Since(opStart), "error", err)
	return newVal, err
}

func (c *counter) Decr(ctx context.Context, key string, delta int64) (int64, error) {
	opStart := time.Now()
	if err := cache.ValidateKey(key); err != nil {
		c.m.logOperation(ctx, "decrby", "key", key, "delta", delta, "latency", time.Since(opStart), "error", err)
		return 0, err
	}

	var newVal int64
	var err error

	c.m.traceOperationWithArgs(ctx, "DECR", &cache.OperationArgs{Key: key, DeltaInt: delta}, func(ctx context.Context) error {
		c.m.mu.Lock()
		defer c.m.mu.Unlock()

		expiresAt := time.Time{}
		e, ok := c.m.data[key]
		if !ok || e == nil || e.isExpired() {
			newVal = -delta
		} else {
			expiresAt = e.expiresAt
			v, parseErr := strconv.ParseInt(string(e.value), 10, 64)
			if parseErr != nil {
				err = parseErr
				return nil
			}
			newVal = v - delta
		}

		c.m.data[key] = &entry{
			value:     []byte(strconv.FormatInt(newVal, 10)),
			expiresAt: expiresAt,
		}
		return nil
	})

	c.m.logOperation(ctx, "decrby", "key", key, "delta", delta, "newVal", newVal, "latency", time.Since(opStart), "error", err)
	return newVal, err
}

func (c *counter) Get(ctx context.Context, key string) (int64, error) {
	opStart := time.Now()
	if err := cache.ValidateKey(key); err != nil {
		c.m.logOperation(ctx, "get", "key", key, "latency", time.Since(opStart), "error", err)
		return 0, err
	}

	var val int64
	var err error

	c.m.traceOperationWithArgs(ctx, "GET", &cache.OperationArgs{Key: key}, func(ctx context.Context) error {
		c.m.mu.RLock()
		defer c.m.mu.RUnlock()

		e, ok := c.m.data[key]
		if !ok || e.isExpired() {
			err = cache.ErrKeyNotFound
			return nil
		}
		val, err = strconv.ParseInt(string(e.value), 10, 64)
		return nil
	})

	c.m.logOperation(ctx, "get", "key", key, "val", val, "latency", time.Since(opStart), "error", err)
	return val, err
}

func (c *counter) Set(ctx context.Context, key string, value int64) error {
	opStart := time.Now()
	if err := cache.ValidateKey(key); err != nil {
		c.m.logOperation(ctx, "set", "key", key, "value", value, "latency", time.Since(opStart), "error", err)
		return err
	}

	var err error

	c.m.traceOperationWithArgs(ctx, "SET", &cache.OperationArgs{Key: key, DeltaInt: value}, func(ctx context.Context) error {
		c.m.mu.Lock()
		defer c.m.mu.Unlock()

		c.m.data[key] = &entry{
			value:     []byte(strconv.FormatInt(value, 10)),
			expiresAt: time.Time{},
		}
		return nil
	})

	c.m.logOperation(ctx, "set", "key", key, "value", value, "latency", time.Since(opStart), "error", err)
	return err
}

func (c *counter) Expire(ctx context.Context, key string, ttl time.Duration) error {
	opStart := time.Now()
	if err := cache.ValidateKey(key); err != nil {
		c.m.logOperation(ctx, "expire", "key", key, "ttl", ttl, "latency", time.Since(opStart), "error", err)
		return err
	}

	var err error

	c.m.traceOperationWithArgs(ctx, "EXPIRE", &cache.OperationArgs{Key: key}, func(ctx context.Context) error {
		c.m.mu.Lock()
		defer c.m.mu.Unlock()

		e, ok := c.m.data[key]
		if !ok || e.isExpired() {
			err = cache.ErrKeyNotFound
			return nil
		}

		if ttl > 0 {
			e.expiresAt = time.Now().Add(ttl)
		} else {
			e.expiresAt = time.Time{}
		}
		return nil
	})

	c.m.logOperation(ctx, "expire", "key", key, "ttl", ttl, "latency", time.Since(opStart), "error", err)
	return err
}
