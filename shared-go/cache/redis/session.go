package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/shared-go/cache"
)

type Session struct {
	client        *redis.Client
	logger        log.Logger
	logLevel      log.Level
	slowQuery     bool
	slowThreshold time.Duration
	prefix        string
}

func NewSession(client *redis.Client, logger log.Logger, logLevel string, slowQuery bool, slowThreshold time.Duration) cache.Session {
	return &Session{
		client:        client,
		logger:        logger,
		logLevel:      parseLogLevel(logLevel),
		slowQuery:     slowQuery,
		slowThreshold: slowThreshold,
		prefix:        "session:",
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

	start := time.Now()
	exists, err := s.client.Exists(ctx, s.sessionKey(sessionID)).Result()
	if err != nil {
		s.logOperation(ctx, "EXISTS", sessionID, "", err, time.Since(start))
		return false, err
	}
	s.logOperation(ctx, "EXISTS", sessionID, "", nil, time.Since(start))
	return exists > 0, nil
}

func (s *Session) Refresh(ctx context.Context, sessionID string, ttl time.Duration) error {
	if err := cache.ValidateSessionID(sessionID); err != nil {
		return err
	}

	start := time.Now()
	sessionKey := s.sessionKey(sessionID)

	exists, err := s.client.Exists(ctx, sessionKey).Result()
	if err != nil {
		s.logOperation(ctx, "REFRESH", sessionID, "", err, time.Since(start))
		return err
	}
	if exists == 0 {
		err = cache.ErrSessionNotFound
		s.logOperation(ctx, "REFRESH", sessionID, "", err, time.Since(start))
		return err
	}

	if ttl == 0 {
		_, err = s.client.Persist(ctx, sessionKey).Result()
	} else {
		err = s.client.Expire(ctx, sessionKey, ttl).Err()
	}
	if err != nil {
		s.logOperation(ctx, "REFRESH", sessionID, "", err, time.Since(start))
		return err
	}

	s.logOperation(ctx, "REFRESH", sessionID, "", nil, time.Since(start))
	return nil
}

func (s *Session) Destroy(ctx context.Context, sessionID string) error {
	if err := cache.ValidateSessionID(sessionID); err != nil {
		return err
	}

	start := time.Now()
	err := s.client.Del(ctx, s.sessionKey(sessionID)).Err()
	if err != nil {
		s.logOperation(ctx, "DESTROY", sessionID, "", err, time.Since(start))
		return err
	}
	s.logOperation(ctx, "DESTROY", sessionID, "", nil, time.Since(start))
	return nil
}

func (s *Session) Keys(ctx context.Context, sessionID string) ([]string, error) {
	if err := cache.ValidateSessionID(sessionID); err != nil {
		return nil, err
	}

	start := time.Now()
	keys, err := s.client.HKeys(ctx, s.sessionKey(sessionID)).Result()
	if err != nil {
		s.logOperation(ctx, "KEYS", sessionID, "", err, time.Since(start))
		return nil, err
	}
	s.logOperation(ctx, "KEYS", sessionID, "", nil, time.Since(start))
	return keys, nil
}

func (s *Session) logOperation(ctx context.Context, op, sessionID, key string, err error, duration time.Duration) {
	if s.logger == nil {
		return
	}

	logger := log.WithContext(ctx, s.logger)

	fields := []any{
		"msg", "Cache operation",
		"component", "cache",
		"backend", "redis",
		"operation", op,
		"sessionID", sessionID,
		"duration", duration.String(),
	}

	if key != "" {
		fields = append(fields, "key", key)
	}

	if err != nil {
		fields = append(fields, "error", err.Error())
		if errors.Is(err, cache.ErrSessionNotFound) || errors.Is(err, redis.Nil) {
			fields = append(fields, "expected_error", true)
			_ = logger.Log(s.logLevel, fields...)
			return
		}
		_ = logger.Log(log.LevelError, fields...)
	} else if s.slowQuery && s.slowThreshold > 0 && duration > s.slowThreshold {
		fields = append(fields, "slow_query", true, "threshold", s.slowThreshold.String())
		_ = logger.Log(log.LevelWarn, fields...)
	} else {
		_ = logger.Log(s.logLevel, fields...)
	}
}
