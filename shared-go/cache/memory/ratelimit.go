package memory

import (
	"context"
	"time"

	"cyber-ecosystem/shared-go/cache"
)

type rateLimiter struct {
	m *Memory
}

func newRateLimiter(m *Memory) *rateLimiter {
	return &rateLimiter{m: m}
}

func (r *rateLimiter) Allow(ctx context.Context, key string, quota int64, window time.Duration) (bool, error) {
	opStart := time.Now()
	if err := cache.ValidateRateLimit(key, quota, window); err != nil {
		r.m.logOperation(ctx, "allow", "key", key, "quota", quota, "window", window, "latency", time.Since(opStart), "error", err)
		return false, err
	}

	var allowed bool

	err := r.m.traceOperationWithArgs(ctx, "ALLOW", &cache.OperationArgs{Key: key, Quota: quota, Window: window.Milliseconds()}, func(ctx context.Context) error {
		r.m.mu.Lock()
		defer r.m.mu.Unlock()

		now := time.Now().UnixMilli()
		windowStart := now - window.Milliseconds()

		e, ok := r.m.data[key]
		if !ok || e == nil || e.isExpired() {
			e = &entry{
				value: make([]byte, 0),
			}
		}

		zs := deserializeRateLimitData(e.value)
		var newMembers []int64
		var keepCount int64

		for _, ts := range zs.members {
			if ts > windowStart {
				keepCount++
				newMembers = append(newMembers, ts)
			}
		}

		newMembers = append(newMembers, now)
		keepCount++

		e.value = serializeRateLimitData(newMembers)
		if window > 0 {
			e.expiresAt = time.Now().Add(window)
		}
		r.m.data[key] = e

		if keepCount > quota {
			allowed = false
			return cache.ErrQuotaExceeded
		}
		allowed = true
		return nil
	})

	r.m.logOperation(ctx, "allow", "key", key, "quota", quota, "window", window, "allowed", allowed, "latency", time.Since(opStart), "error", err)
	return allowed, err
}

type rateLimitData struct {
	members []int64
}

func serializeRateLimitData(timestamps []int64) []byte {
	data := make([]byte, 0, len(timestamps)*8)
	for _, ts := range timestamps {
		data = append(data, byte(ts>>56)&0xFF)
		data = append(data, byte(ts>>48)&0xFF)
		data = append(data, byte(ts>>40)&0xFF)
		data = append(data, byte(ts>>32)&0xFF)
		data = append(data, byte(ts>>24)&0xFF)
		data = append(data, byte(ts>>16)&0xFF)
		data = append(data, byte(ts>>8)&0xFF)
		data = append(data, byte(ts)&0xFF)
	}
	return data
}

func deserializeRateLimitData(data []byte) *rateLimitData {
	zs := &rateLimitData{}
	if len(data) == 0 {
		return zs
	}

	for i := 0; i+8 <= len(data); i += 8 {
		var ts int64
		ts |= int64(data[i]) << 56
		ts |= int64(data[i+1]) << 48
		ts |= int64(data[i+2]) << 40
		ts |= int64(data[i+3]) << 32
		ts |= int64(data[i+4]) << 24
		ts |= int64(data[i+5]) << 16
		ts |= int64(data[i+6]) << 8
		ts |= int64(data[i+7])
		zs.members = append(zs.members, ts)
	}
	return zs
}
