package redis

import (
	"context"
	"errors"
	"math"
	"strconv"

	"github.com/redis/go-redis/v9"

	"cyber-ecosystem/shared-go/capability/cache"
)

type sortedSet struct{ client *redis.Client }

// NewSortedSet returns the redis-backed SortedSet implementation.
func NewSortedSet(client *redis.Client) cache.SortedSet { return &sortedSet{client: client} }

func (z *sortedSet) Add(ctx context.Context, key string, members ...cache.Member) error {
	if err := cache.ValidateKey(key); err != nil {
		return err
	}
	if len(members) == 0 {
		return nil
	}
	return mapErr(z.client.ZAdd(ctx, key, toZ(members)...).Err())
}

func (z *sortedSet) IncrBy(ctx context.Context, key, member string, delta float64) (float64, error) {
	if err := cache.ValidateKey(key); err != nil {
		return 0, err
	}
	v, err := z.client.ZIncrBy(ctx, key, delta, member).Result()
	return v, mapErr(err)
}

func (z *sortedSet) Score(ctx context.Context, key, member string) (float64, error) {
	if err := cache.ValidateKey(key); err != nil {
		return 0, err
	}
	v, err := z.client.ZScore(ctx, key, member).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, cache.ErrKeyNotFound
		}
		return 0, mapErr(err)
	}
	return v, nil
}

func (z *sortedSet) Rank(ctx context.Context, key, member string) (int64, error) {
	if err := cache.ValidateKey(key); err != nil {
		return 0, err
	}
	v, err := z.client.ZRank(ctx, key, member).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, cache.ErrKeyNotFound
		}
		return 0, mapErr(err)
	}
	return v, nil
}

func (z *sortedSet) RevRank(ctx context.Context, key, member string) (int64, error) {
	if err := cache.ValidateKey(key); err != nil {
		return 0, err
	}
	v, err := z.client.ZRevRank(ctx, key, member).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, cache.ErrKeyNotFound
		}
		return 0, mapErr(err)
	}
	return v, nil
}

func (z *sortedSet) Range(ctx context.Context, key string, start, stop int64) ([]cache.Member, error) {
	if err := cache.ValidateKey(key); err != nil {
		return nil, err
	}
	vals, err := z.client.ZRangeWithScores(ctx, key, start, stop).Result()
	if err != nil {
		return nil, mapErr(err)
	}
	return fromZ(vals), nil
}

func (z *sortedSet) RevRange(ctx context.Context, key string, start, stop int64) ([]cache.Member, error) {
	if err := cache.ValidateKey(key); err != nil {
		return nil, err
	}
	vals, err := z.client.ZRevRangeWithScores(ctx, key, start, stop).Result()
	if err != nil {
		return nil, mapErr(err)
	}
	return fromZ(vals), nil
}

func (z *sortedSet) RangeByScore(ctx context.Context, key string, min, max float64, offset, count int64) ([]cache.Member, error) {
	if err := cache.ValidateKey(key); err != nil {
		return nil, err
	}
	// NaN/Inf produce strings redis rejects at runtime — fail fast at the boundary.
	if math.IsNaN(min) || math.IsNaN(max) || math.IsInf(min, 0) || math.IsInf(max, 0) {
		return nil, cache.ErrInvalidArgument
	}
	vals, err := z.client.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
		Min:    strconv.FormatFloat(min, 'f', -1, 64),
		Max:    strconv.FormatFloat(max, 'f', -1, 64),
		Offset: offset,
		Count:  count,
	}).Result()
	if err != nil {
		return nil, mapErr(err)
	}
	return fromZ(vals), nil
}

func (z *sortedSet) Remove(ctx context.Context, key string, members ...string) error {
	if err := cache.ValidateKey(key); err != nil {
		return err
	}
	if len(members) == 0 {
		return nil
	}
	return mapErr(z.client.ZRem(ctx, key, toAny(members)...).Err())
}

func (z *sortedSet) Card(ctx context.Context, key string) (int64, error) {
	if err := cache.ValidateKey(key); err != nil {
		return 0, err
	}
	n, err := z.client.ZCard(ctx, key).Result()
	return n, mapErr(err)
}

func toZ(members []cache.Member) []redis.Z {
	out := make([]redis.Z, len(members))
	for i, m := range members {
		out[i] = redis.Z{Score: m.Score, Member: m.Member}
	}
	return out
}

func fromZ(vals []redis.Z) []cache.Member {
	out := make([]cache.Member, len(vals))
	for i, v := range vals {
		// redis.Z.Member is interface{}; go-redis decodes members as strings for
		// the read commands used here, but use comma-ok to stay panic-safe.
		if s, ok := v.Member.(string); ok {
			out[i] = cache.Member{Score: v.Score, Member: s}
		}
	}
	return out
}
