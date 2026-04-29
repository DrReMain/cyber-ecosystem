package memory

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cyber-ecosystem/shared-go/cache"
)

type session struct {
	m      *Memory
	prefix string
}

func newSession(m *Memory) *session {
	return &session{
		m:      m,
		prefix: "session:",
	}
}

func (s *session) sessionKey(sessionID, key string) string {
	return fmt.Sprintf("%s%s:%s", s.prefix, sessionID, key)
}

func (s *session) Get(ctx context.Context, sessionID, key string) ([]byte, error) {
	opStart := time.Now()
	if err := cache.ValidateSessionID(sessionID); err != nil {
		s.m.logOperation(ctx, "get", "sessionID", sessionID, "key", key, "latency", time.Since(opStart), "error", err)
		return nil, err
	}
	if err := cache.ValidateKey(key); err != nil {
		s.m.logOperation(ctx, "get", "sessionID", sessionID, "key", key, "latency", time.Since(opStart), "error", err)
		return nil, err
	}

	var result []byte
	var err error

	s.m.traceOperationWithArgs(ctx, "GET", &cache.OperationArgs{SessionID: sessionID, Key: key}, func(ctx context.Context) error {
		s.m.mu.RLock()
		defer s.m.mu.RUnlock()

		sk := s.sessionKey(sessionID, key)
		e, ok := s.m.data[sk]
		if !ok || e.isExpired() {
			err = cache.ErrSessionNotFound
			return nil
		}
		result = make([]byte, len(e.value))
		copy(result, e.value)
		return nil
	})

	s.m.logOperation(ctx, "get", "sessionID", sessionID, "key", key, "latency", time.Since(opStart), "error", err)
	return result, err
}

func (s *session) Set(ctx context.Context, sessionID, key string, val []byte, ttl time.Duration) error {
	opStart := time.Now()
	if err := cache.ValidateSessionID(sessionID); err != nil {
		s.m.logOperation(ctx, "set", "sessionID", sessionID, "key", key, "ttl", ttl, "latency", time.Since(opStart), "error", err)
		return err
	}
	if err := cache.ValidateKey(key); err != nil {
		s.m.logOperation(ctx, "set", "sessionID", sessionID, "key", key, "ttl", ttl, "latency", time.Since(opStart), "error", err)
		return err
	}

	var err error

	s.m.traceOperationWithArgs(ctx, "SET", &cache.OperationArgs{SessionID: sessionID, Key: key, Value: val}, func(ctx context.Context) error {
		s.m.mu.Lock()
		defer s.m.mu.Unlock()

		e := &entry{
			value: make([]byte, len(val)),
		}
		copy(e.value, val)
		if ttl > 0 {
			e.expiresAt = time.Now().Add(ttl)
		}
		s.m.data[s.sessionKey(sessionID, key)] = e
		return nil
	})

	s.m.logOperation(ctx, "set", "sessionID", sessionID, "key", key, "ttl", ttl, "latency", time.Since(opStart))
	return err
}

func (s *session) Delete(ctx context.Context, sessionID, key string) error {
	opStart := time.Now()
	if err := cache.ValidateSessionID(sessionID); err != nil {
		s.m.logOperation(ctx, "del", "sessionID", sessionID, "key", key, "latency", time.Since(opStart), "error", err)
		return err
	}
	if err := cache.ValidateKey(key); err != nil {
		s.m.logOperation(ctx, "del", "sessionID", sessionID, "key", key, "latency", time.Since(opStart), "error", err)
		return err
	}

	var err error

	s.m.traceOperationWithArgs(ctx, "DEL", &cache.OperationArgs{SessionID: sessionID, Key: key}, func(ctx context.Context) error {
		s.m.mu.Lock()
		defer s.m.mu.Unlock()

		delete(s.m.data, s.sessionKey(sessionID, key))
		return nil
	})

	s.m.logOperation(ctx, "del", "sessionID", sessionID, "key", key, "latency", time.Since(opStart))
	return err
}

func (s *session) Exists(ctx context.Context, sessionID string) (bool, error) {
	opStart := time.Now()
	if err := cache.ValidateSessionID(sessionID); err != nil {
		s.m.logOperation(ctx, "exists", "sessionID", sessionID, "latency", time.Since(opStart), "error", err)
		return false, err
	}

	var exist bool

	s.m.traceOperationWithArgs(ctx, "EXISTS", &cache.OperationArgs{SessionID: sessionID}, func(ctx context.Context) error {
		s.m.mu.RLock()
		defer s.m.mu.RUnlock()

		prefix := s.prefix + sessionID + ":"

		for key := range s.m.data {
			if strings.HasPrefix(key, prefix) {
				e := s.m.data[key]
				if !e.isExpired() {
					exist = true
					return nil
				}
			}
		}
		return nil
	})

	s.m.logOperation(ctx, "exists", "sessionID", sessionID, "exist", exist, "latency", time.Since(opStart))
	return exist, nil
}

func (s *session) Refresh(ctx context.Context, sessionID string, ttl time.Duration) error {
	opStart := time.Now()
	if err := cache.ValidateSessionID(sessionID); err != nil {
		s.m.logOperation(ctx, "refresh", "sessionID", sessionID, "ttl", ttl, "latency", time.Since(opStart), "error", err)
		return err
	}

	var err error

	s.m.traceOperationWithArgs(ctx, "REFRESH", &cache.OperationArgs{SessionID: sessionID}, func(ctx context.Context) error {
		s.m.mu.Lock()
		defer s.m.mu.Unlock()

		prefix := s.prefix + sessionID + ":"
		now := time.Now()
		found := false

		for key, e := range s.m.data {
			if strings.HasPrefix(key, prefix) && !e.isExpired() {
				found = true
				if ttl > 0 {
					e.expiresAt = now.Add(ttl)
				} else {
					e.expiresAt = time.Time{}
				}
			}
		}
		if !found {
			err = cache.ErrSessionNotFound
		}
		return nil
	})

	s.m.logOperation(ctx, "refresh", "sessionID", sessionID, "ttl", ttl, "latency", time.Since(opStart), "error", err)
	return err
}

func (s *session) Destroy(ctx context.Context, sessionID string) error {
	opStart := time.Now()
	if err := cache.ValidateSessionID(sessionID); err != nil {
		s.m.logOperation(ctx, "destroy", "sessionID", sessionID, "latency", time.Since(opStart), "error", err)
		return err
	}

	var err error

	s.m.traceOperationWithArgs(ctx, "DESTROY", &cache.OperationArgs{SessionID: sessionID}, func(ctx context.Context) error {
		s.m.mu.Lock()
		defer s.m.mu.Unlock()

		prefix := s.prefix + sessionID + ":"
		for key := range s.m.data {
			if strings.HasPrefix(key, prefix) {
				delete(s.m.data, key)
			}
		}
		return nil
	})

	s.m.logOperation(ctx, "destroy", "sessionID", sessionID, "latency", time.Since(opStart))
	return err
}

func (s *session) Keys(ctx context.Context, sessionID string) ([]string, error) {
	opStart := time.Now()
	if err := cache.ValidateSessionID(sessionID); err != nil {
		s.m.logOperation(ctx, "keys", "sessionID", sessionID, "latency", time.Since(opStart), "error", err)
		return nil, err
	}

	var keys []string

	s.m.traceOperationWithArgs(ctx, "KEYS", &cache.OperationArgs{SessionID: sessionID}, func(ctx context.Context) error {
		s.m.mu.RLock()
		defer s.m.mu.RUnlock()

		prefix := s.prefix + sessionID + ":"
		prefixLen := len(prefix)
		for key := range s.m.data {
			if strings.HasPrefix(key, prefix) {
				e := s.m.data[key]
				if !e.isExpired() {
					// Return only the user-level key (without session:sessionID: prefix)
					keys = append(keys, key[prefixLen:])
				}
			}
		}
		return nil
	})

	s.m.logOperation(ctx, "keys", "sessionID", sessionID, "count", len(keys), "latency", time.Since(opStart))
	return keys, nil
}
