package redis

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"

	"cyber-ecosystem/shared-go/cache"
)

type SortedSet struct {
	client *redis.Client
}

func NewSortedSet(client *redis.Client) cache.SortedSet {
	return &SortedSet{client: client}
}

func (s *SortedSet) Add(ctx context.Context, key string, members ...cache.Member) error {
	if len(members) == 0 {
		return nil
	}
	if err := cache.ValidateKey(key); err != nil {
		return err
	}
	if err := cache.ValidateMembers(members); err != nil {
		return err
	}
	zs := make([]redis.Z, len(members))
	for i, m := range members {
		zs[i] = redis.Z{Score: m.Score, Member: m.Member}
	}
	return s.client.ZAdd(ctx, key, zs...).Err()
}

func (s *SortedSet) Incr(ctx context.Context, key string, member string, delta float64) (float64, error) {
	if err := cache.ValidateKey(key); err != nil {
		return 0, err
	}
	if err := cache.ValidateMember(member); err != nil {
		return 0, err
	}
	return s.client.ZIncrBy(ctx, key, delta, member).Result()
}

func (s *SortedSet) Rank(ctx context.Context, key string, member string) (int64, error) {
	if err := cache.ValidateKey(key); err != nil {
		return 0, err
	}
	if err := cache.ValidateMember(member); err != nil {
		return 0, err
	}
	rank, err := s.client.ZRank(ctx, key, member).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, cache.ErrKeyNotFound
		}
		return 0, err
	}
	return rank, nil
}

func (s *SortedSet) RevRank(ctx context.Context, key string, member string) (int64, error) {
	if err := cache.ValidateKey(key); err != nil {
		return 0, err
	}
	if err := cache.ValidateMember(member); err != nil {
		return 0, err
	}
	rank, err := s.client.ZRevRank(ctx, key, member).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, cache.ErrKeyNotFound
		}
		return 0, err
	}
	return rank, nil
}

func (s *SortedSet) Score(ctx context.Context, key string, member string) (float64, error) {
	if err := cache.ValidateKey(key); err != nil {
		return 0, err
	}
	if err := cache.ValidateMember(member); err != nil {
		return 0, err
	}
	score, err := s.client.ZScore(ctx, key, member).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, cache.ErrKeyNotFound
		}
		return 0, err
	}
	return score, nil
}

func (s *SortedSet) Range(ctx context.Context, key string, start, stop int64) ([]cache.Member, error) {
	if err := cache.ValidateKey(key); err != nil {
		return nil, err
	}
	result, err := s.client.ZRangeWithScores(ctx, key, start, stop).Result()
	if err != nil {
		return nil, err
	}
	members := make([]cache.Member, len(result))
	for i, z := range result {
		members[i] = cache.Member{
			Member: z.Member.(string),
			Score:  z.Score,
		}
	}
	return members, nil
}

func (s *SortedSet) RevRange(ctx context.Context, key string, start, stop int64) ([]cache.Member, error) {
	if err := cache.ValidateKey(key); err != nil {
		return nil, err
	}
	result, err := s.client.ZRevRangeWithScores(ctx, key, start, stop).Result()
	if err != nil {
		return nil, err
	}
	members := make([]cache.Member, len(result))
	for i, z := range result {
		members[i] = cache.Member{
			Member: z.Member.(string),
			Score:  z.Score,
		}
	}
	return members, nil
}

func (s *SortedSet) Remove(ctx context.Context, key string, members ...string) error {
	if len(members) == 0 {
		return nil
	}
	if err := cache.ValidateKey(key); err != nil {
		return err
	}
	if err := cache.ValidateMemberStrings(members...); err != nil {
		return err
	}
	args := make([]interface{}, len(members))
	for i, m := range members {
		args[i] = m
	}
	return s.client.ZRem(ctx, key, args...).Err()
}
