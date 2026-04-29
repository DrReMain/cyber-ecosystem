package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"cyber-ecosystem/shared-go/cache"
)

type Session struct {
	client *redis.Client
	prefix string
}

func NewSession(client *redis.Client) cache.Session {
	return &Session{
		client: client,
		prefix: "session:",
	}
}

func (s *Session) sessionKey(sessionID string) string {
	return fmt.Sprintf("%s%s", s.prefix, sessionID)
}

func (s *Session) Get(ctx context.Context, sessionID, key string) ([]byte, error) {
	if err := cache.ValidateSessionID(sessionID); err != nil {
		return nil, err
	}
	if err := cache.ValidateKey(key); err != nil {
		return nil, err
	}

	val, err := s.client.HGet(ctx, s.sessionKey(sessionID), key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, cache.ErrSessionNotFound
		}
		return nil, err
	}
	return val, nil
}

func (s *Session) Set(ctx context.Context, sessionID, key string, val []byte, ttl time.Duration) error {
	if err := cache.ValidateSessionID(sessionID); err != nil {
		return err
	}
	if err := cache.ValidateKey(key); err != nil {
		return err
	}

	sessionKey := s.sessionKey(sessionID)
	pipe := s.client.Pipeline()
	pipe.HSet(ctx, sessionKey, key, val)
	if ttl == 0 {
		pipe.Persist(ctx, sessionKey)
	} else {
		pipe.Expire(ctx, sessionKey, ttl)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (s *Session) Delete(ctx context.Context, sessionID, key string) error {
	if err := cache.ValidateSessionID(sessionID); err != nil {
		return err
	}
	if err := cache.ValidateKey(key); err != nil {
		return err
	}

	return s.client.HDel(ctx, s.sessionKey(sessionID), key).Err()
}

func (s *Session) Exists(ctx context.Context, sessionID string) (bool, error) {
	if err := cache.ValidateSessionID(sessionID); err != nil {
		return false, err
	}

	exists, err := s.client.Exists(ctx, s.sessionKey(sessionID)).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func (s *Session) Refresh(ctx context.Context, sessionID string, ttl time.Duration) error {
	if err := cache.ValidateSessionID(sessionID); err != nil {
		return err
	}

	sessionKey := s.sessionKey(sessionID)

	exists, err := s.client.Exists(ctx, sessionKey).Result()
	if err != nil {
		return err
	}
	if exists == 0 {
		return cache.ErrSessionNotFound
	}

	if ttl == 0 {
		_, err = s.client.Persist(ctx, sessionKey).Result()
	} else {
		err = s.client.Expire(ctx, sessionKey, ttl).Err()
	}
	return err
}

func (s *Session) Destroy(ctx context.Context, sessionID string) error {
	if err := cache.ValidateSessionID(sessionID); err != nil {
		return err
	}

	return s.client.Del(ctx, s.sessionKey(sessionID)).Err()
}

func (s *Session) Keys(ctx context.Context, sessionID string) ([]string, error) {
	if err := cache.ValidateSessionID(sessionID); err != nil {
		return nil, err
	}

	return s.client.HKeys(ctx, s.sessionKey(sessionID)).Result()
}
