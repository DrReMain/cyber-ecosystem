package redis

import (
	"context"
	"time"

	nanoid "github.com/matoous/go-nanoid/v2"
	"github.com/redis/go-redis/v9"

	"cyber-ecosystem/shared-go/cache"
)

type RateLimiter struct {
	client *redis.Client
}

var rateLimitScript = redis.NewScript(`
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local quota = tonumber(ARGV[3])
local member = ARGV[4]
local windowStart = now - window

redis.call('ZADD', key, now, member)
redis.call('ZREMRANGEBYSCORE', key, 0, windowStart)
local count = redis.call('ZCARD', key)
redis.call('PEXPIRE', key, window)

if count > quota then
	return 0
end
return 1
`)

func NewRateLimiter(client *redis.Client) cache.RateLimiter {
	return &RateLimiter{client: client}
}

func (l *RateLimiter) Allow(ctx context.Context, key string, quota int64, window time.Duration) (bool, error) {
	if err := cache.ValidateRateLimit(key, quota, window); err != nil {
		return false, err
	}

	now := time.Now().UnixMilli()

	id, err := nanoid.New()
	if err != nil {
		return false, err
	}
	member := formatMember(now, id)

	res, err := rateLimitScript.Run(ctx, l.client, []string{key}, now, window.Milliseconds(), quota, member).Int64()
	if err != nil {
		return false, err
	}
	if res == 0 {
		return false, cache.ErrQuotaExceeded
	}
	return true, nil
}

func formatMember(now int64, id string) string {
	return time.UnixMilli(now).UTC().Format(time.RFC3339Nano) + ":" + id
}
