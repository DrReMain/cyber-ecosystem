package cache

import "context"

// Set is an unordered collection of unique members (tags, membership, dedup).
type Set interface {
	SAdd(ctx context.Context, key string, members ...string) (int64, error)
	SRem(ctx context.Context, key string, members ...string) (int64, error)
	SMembers(ctx context.Context, key string) ([]string, error)
	SIsMember(ctx context.Context, key, member string) (bool, error)
	SCard(ctx context.Context, key string) (int64, error)
	SInter(ctx context.Context, keys ...string) ([]string, error)
	SUnion(ctx context.Context, keys ...string) ([]string, error)
}
