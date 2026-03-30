package memory

import (
	"context"
	"time"

	"cyber-ecosystem/shared-go/cache"
)

type kv struct {
	m *Memory
}

func (k *kv) Get(ctx context.Context, key string) ([]byte, error) {
	opStart := time.Now()
	if err := cache.ValidateKey(key); err != nil {
		k.m.logOperation(ctx, "get", "key", key, "duration", time.Since(opStart), "error", err)
		return nil, err
	}

	var result []byte
	var err error

	k.m.traceOperationWithArgs(ctx, "GET", &cache.OperationArgs{Key: key}, func(ctx context.Context) error {
		k.m.mu.RLock()
		defer k.m.mu.RUnlock()

		e, ok := k.m.data[key]
		if !ok || e.isExpired() {
			err = cache.ErrCacheMiss
			return nil
		}
		result = make([]byte, len(e.value))
		copy(result, e.value)
		return nil
	})

	k.m.logOperation(ctx, "get", "key", key, "duration", time.Since(opStart), "error", err)
	return result, err
}

func (k *kv) Set(ctx context.Context, key string, val []byte, ttl time.Duration) error {
	opStart := time.Now()
	if err := cache.ValidateKey(key); err != nil {
		k.m.logOperation(ctx, "set", "key", key, "ttl", ttl, "duration", time.Since(opStart), "error", err)
		return err
	}

	var err error

	k.m.traceOperationWithArgs(ctx, "SET", &cache.OperationArgs{Key: key, Value: val}, func(ctx context.Context) error {
		k.m.mu.Lock()
		defer k.m.mu.Unlock()

		e := &entry{
			value: make([]byte, len(val)),
		}
		copy(e.value, val)
		if ttl > 0 {
			e.expiresAt = time.Now().Add(ttl)
		}
		k.m.data[key] = e
		return nil
	})

	k.m.logOperation(ctx, "set", "key", key, "ttl", ttl, "duration", time.Since(opStart), "error", err)
	return err
}

func (k *kv) Delete(ctx context.Context, key string) error {
	opStart := time.Now()
	if err := cache.ValidateKey(key); err != nil {
		k.m.logOperation(ctx, "del", "key", key, "duration", time.Since(opStart), "error", err)
		return err
	}

	var err error

	k.m.traceOperationWithArgs(ctx, "DEL", &cache.OperationArgs{Key: key}, func(ctx context.Context) error {
		k.m.mu.Lock()
		defer k.m.mu.Unlock()

		delete(k.m.data, key)
		return nil
	})

	k.m.logOperation(ctx, "del", "key", key, "duration", time.Since(opStart), "error", err)
	return err
}

func (k *kv) Exist(ctx context.Context, key string) (bool, error) {
	opStart := time.Now()
	if err := cache.ValidateKey(key); err != nil {
		k.m.logOperation(ctx, "exists", "key", key, "duration", time.Since(opStart), "error", err)
		return false, err
	}

	var exist bool
	var err error

	k.m.traceOperationWithArgs(ctx, "EXISTS", &cache.OperationArgs{Key: key}, func(ctx context.Context) error {
		k.m.mu.RLock()
		defer k.m.mu.RUnlock()

		e, ok := k.m.data[key]
		if !ok || e.isExpired() {
			exist = false
			return nil
		}
		exist = true
		return nil
	})

	k.m.logOperation(ctx, "exists", "key", key, "exist", exist, "duration", time.Since(opStart), "error", err)
	return exist, err
}

func (k *kv) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	opStart := time.Now()
	if err := cache.ValidateKey(key); err != nil {
		k.m.logOperation(ctx, "ttl", "key", key, "duration", time.Since(opStart), "error", err)
		return 0, err
	}

	var ttl time.Duration
	var err error

	k.m.traceOperationWithArgs(ctx, "TTL", &cache.OperationArgs{Key: key}, func(ctx context.Context) error {
		k.m.mu.RLock()
		defer k.m.mu.RUnlock()

		e, ok := k.m.data[key]
		if !ok || e.isExpired() {
			err = cache.ErrCacheMiss
			return nil
		}
		if e.expiresAt.IsZero() {
			ttl = 0
		} else {
			ttl = time.Until(e.expiresAt)
		}
		return nil
	})

	k.m.logOperation(ctx, "ttl", "key", key, "ttl", ttl, "duration", time.Since(opStart), "error", err)
	return ttl, err
}

func (k *kv) MGet(ctx context.Context, keys ...string) ([][]byte, error) {
	opStart := time.Now()
	if len(keys) == 0 {
		return nil, nil
	}
	if err := cache.ValidateKeys(keys...); err != nil {
		k.m.logOperation(ctx, "mget", "keys", keys, "duration", time.Since(opStart), "error", err)
		return nil, err
	}

	result := make([][]byte, len(keys))

	k.m.traceOperationWithArgs(ctx, "MGET", &cache.OperationArgs{Keys: keys}, func(ctx context.Context) error {
		k.m.mu.RLock()
		defer k.m.mu.RUnlock()

		for i, key := range keys {
			e, ok := k.m.data[key]
			if !ok || e.isExpired() {
				result[i] = nil
				continue
			}
			result[i] = make([]byte, len(e.value))
			copy(result[i], e.value)
		}
		return nil
	})

	k.m.logOperation(ctx, "mget", "keys", keys, "duration", time.Since(opStart))
	return result, nil
}

func (k *kv) MSet(ctx context.Context, pairs map[string][]byte, ttl time.Duration) error {
	opStart := time.Now()
	if len(pairs) == 0 {
		return nil
	}
	if err := cache.ValidatePairs(pairs); err != nil {
		k.m.logOperation(ctx, "mset", "count", len(pairs), "ttl", ttl, "duration", time.Since(opStart), "error", err)
		return err
	}

	k.m.traceOperationWithArgs(ctx, "MSET", &cache.OperationArgs{Values: pairs}, func(ctx context.Context) error {
		k.m.mu.Lock()
		defer k.m.mu.Unlock()

		for key, val := range pairs {
			e := &entry{
				value: make([]byte, len(val)),
			}
			copy(e.value, val)
			if ttl > 0 {
				e.expiresAt = time.Now().Add(ttl)
			}
			k.m.data[key] = e
		}
		return nil
	})

	k.m.logOperation(ctx, "mset", "count", len(pairs), "ttl", ttl, "duration", time.Since(opStart))
	return nil
}
