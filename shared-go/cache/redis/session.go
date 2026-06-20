package redis

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"cyber-ecosystem/shared-go/cache"
)

type session struct{ client *redis.Client }

// NewSession returns the redis-backed Session implementation.
func NewSession(client *redis.Client) cache.Session { return &session{client: client} }

const sessionNamespace = "session:"

func sessionKey(id, key string) string      { return sessionNamespace + id + ":" + key }
func sessionMatch(id string) string         { return sessionNamespace + id + ":*" }
func sessionUserKey(id, full string) string { return strings.TrimPrefix(full, sessionNamespace+id+":") }

func (s *session) Get(ctx context.Context, id, key string) ([]byte, error) {
	if err := cache.ValidateSessionID(id); err != nil {
		return nil, err
	}
	if err := cache.ValidateSessionKey(key); err != nil {
		return nil, err
	}
	val, err := s.client.Get(ctx, sessionKey(id, key)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, cache.ErrCacheMiss
		}
		return nil, mapErr(err)
	}
	return val, nil
}

func (s *session) Set(ctx context.Context, id, key string, val []byte, ttl time.Duration) error {
	if err := cache.ValidateSessionID(id); err != nil {
		return err
	}
	if err := cache.ValidateSessionKey(key); err != nil {
		return err
	}
	return mapErr(s.client.Set(ctx, sessionKey(id, key), val, ttl).Err())
}

func (s *session) Del(ctx context.Context, id, key string) error {
	if err := cache.ValidateSessionID(id); err != nil {
		return err
	}
	if err := cache.ValidateSessionKey(key); err != nil {
		return err
	}
	return mapErr(s.client.Del(ctx, sessionKey(id, key)).Err())
}

func (s *session) Exists(ctx context.Context, id string) (bool, error) {
	if err := cache.ValidateSessionID(id); err != nil {
		return false, err
	}
	iter := s.client.Scan(ctx, 0, sessionMatch(id), 100).Iterator()
	ok := iter.Next(ctx)
	return ok, mapErr(iter.Err())
}

func (s *session) Refresh(ctx context.Context, id string, ttl time.Duration) error {
	if err := cache.ValidateSessionID(id); err != nil {
		return err
	}
	keys, err := scanKeys(ctx, s.client, sessionMatch(id))
	if err != nil {
		return mapErr(err)
	}
	if len(keys) == 0 {
		return cache.ErrSessionNotFound
	}
	pipe := s.client.Pipeline()
	for _, k := range keys {
		pipe.Expire(ctx, k, ttl)
	}
	_, err = pipe.Exec(ctx)
	return mapErr(err)
}

func (s *session) Destroy(ctx context.Context, id string) error {
	if err := cache.ValidateSessionID(id); err != nil {
		return err
	}
	keys, err := scanKeys(ctx, s.client, sessionMatch(id))
	if err != nil {
		return mapErr(err)
	}
	if len(keys) == 0 {
		return nil
	}
	return mapErr(s.client.Del(ctx, keys...).Err())
}

func (s *session) Keys(ctx context.Context, id string) ([]string, error) {
	if err := cache.ValidateSessionID(id); err != nil {
		return nil, err
	}
	keys, err := scanKeys(ctx, s.client, sessionMatch(id))
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]string, len(keys))
	for i, k := range keys {
		out[i] = sessionUserKey(id, k)
	}
	return out, nil
}

// scanKeys collects all keys matching pattern via a SCAN iterator. NOTE: SCAN is
// non-atomic — keys mutated during iteration may be missed or stale; Refresh/
// Destroy are therefore best-effort over a multi-key pseudo-entity on a single
// redis instance.
func scanKeys(ctx context.Context, client *redis.Client, pattern string) ([]string, error) {
	var keys []string
	iter := client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	return keys, iter.Err()
}
