package redis

import (
	"context"

	"github.com/redis/go-redis/v9"

	"cyber-ecosystem/shared-go/cache"
)

type set struct{ client *redis.Client }

// NewSet returns the redis-backed Set implementation.
func NewSet(client *redis.Client) cache.Set { return &set{client: client} }

func (s *set) SAdd(ctx context.Context, key string, members ...string) (int64, error) {
	if err := cache.ValidateKey(key); err != nil {
		return 0, err
	}
	if len(members) == 0 {
		return 0, nil
	}
	n, err := s.client.SAdd(ctx, key, toAny(members)...).Result()
	return n, mapErr(err)
}

func (s *set) SRem(ctx context.Context, key string, members ...string) (int64, error) {
	if err := cache.ValidateKey(key); err != nil {
		return 0, err
	}
	if len(members) == 0 {
		return 0, nil
	}
	n, err := s.client.SRem(ctx, key, toAny(members)...).Result()
	return n, mapErr(err)
}

func (s *set) SMembers(ctx context.Context, key string) ([]string, error) {
	if err := cache.ValidateKey(key); err != nil {
		return nil, err
	}
	m, err := s.client.SMembers(ctx, key).Result()
	return m, mapErr(err)
}

func (s *set) SIsMember(ctx context.Context, key, member string) (bool, error) {
	if err := cache.ValidateKey(key); err != nil {
		return false, err
	}
	ok, err := s.client.SIsMember(ctx, key, member).Result()
	return ok, mapErr(err)
}

func (s *set) SCard(ctx context.Context, key string) (int64, error) {
	if err := cache.ValidateKey(key); err != nil {
		return 0, err
	}
	n, err := s.client.SCard(ctx, key).Result()
	return n, mapErr(err)
}

func (s *set) SInter(ctx context.Context, keys ...string) ([]string, error) {
	if err := cache.ValidateKeys(keys...); err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return nil, nil
	}
	m, err := s.client.SInter(ctx, keys...).Result()
	return m, mapErr(err)
}

func (s *set) SUnion(ctx context.Context, keys ...string) ([]string, error) {
	if err := cache.ValidateKeys(keys...); err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return nil, nil
	}
	m, err := s.client.SUnion(ctx, keys...).Result()
	return m, mapErr(err)
}
