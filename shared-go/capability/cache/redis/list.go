package redis

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"

	"cyber-ecosystem/shared-go/capability/cache"
)

type list struct{ client *redis.Client }

// NewList returns the redis-backed List implementation.
func NewList(client *redis.Client) cache.List { return &list{client: client} }

func (l *list) LPush(ctx context.Context, key string, vals ...[]byte) (int64, error) {
	if err := cache.ValidateKey(key); err != nil {
		return 0, err
	}
	if len(vals) == 0 {
		return 0, nil
	}
	n, err := l.client.LPush(ctx, key, toAny(vals)...).Result()
	return n, mapErr(err)
}

func (l *list) RPush(ctx context.Context, key string, vals ...[]byte) (int64, error) {
	if err := cache.ValidateKey(key); err != nil {
		return 0, err
	}
	if len(vals) == 0 {
		return 0, nil
	}
	n, err := l.client.RPush(ctx, key, toAny(vals)...).Result()
	return n, mapErr(err)
}

func (l *list) LPop(ctx context.Context, key string) ([]byte, error) {
	if err := cache.ValidateKey(key); err != nil {
		return nil, err
	}
	val, err := l.client.LPop(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, cache.ErrCacheMiss
		}
		return nil, mapErr(err)
	}
	return val, nil
}

func (l *list) RPop(ctx context.Context, key string) ([]byte, error) {
	if err := cache.ValidateKey(key); err != nil {
		return nil, err
	}
	val, err := l.client.RPop(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, cache.ErrCacheMiss
		}
		return nil, mapErr(err)
	}
	return val, nil
}

func (l *list) LRange(ctx context.Context, key string, start, stop int64) ([][]byte, error) {
	if err := cache.ValidateKey(key); err != nil {
		return nil, err
	}
	vals, err := l.client.LRange(ctx, key, start, stop).Result()
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([][]byte, len(vals))
	for i, v := range vals {
		out[i] = []byte(v)
	}
	return out, nil
}

func (l *list) LLen(ctx context.Context, key string) (int64, error) {
	if err := cache.ValidateKey(key); err != nil {
		return 0, err
	}
	n, err := l.client.LLen(ctx, key).Result()
	return n, mapErr(err)
}

func (l *list) LTrim(ctx context.Context, key string, start, stop int64) error {
	if err := cache.ValidateKey(key); err != nil {
		return err
	}
	return mapErr(l.client.LTrim(ctx, key, start, stop).Err())
}

// toAny converts a slice into []any for variadic go-redis commands.
func toAny[T any](vals []T) []any {
	args := make([]any, len(vals))
	for i, v := range vals {
		args[i] = v
	}
	return args
}
