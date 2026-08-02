package cache

import "context"

// Member is a sorted-set entry: a member string ordered by Score.
type Member struct {
	Score  float64
	Member string
}

// SortedSet orders members by score (leaderboards, delayed queues, priority).
// Score/Rank on a missing member return ErrKeyNotFound.
type SortedSet interface {
	Add(ctx context.Context, key string, members ...Member) error
	IncrBy(ctx context.Context, key, member string, delta float64) (float64, error)
	Score(ctx context.Context, key, member string) (float64, error)
	Rank(ctx context.Context, key, member string) (int64, error)    // ascending, 0-indexed
	RevRank(ctx context.Context, key, member string) (int64, error) // descending, 0-indexed
	Range(ctx context.Context, key string, start, stop int64) ([]Member, error)
	RevRange(ctx context.Context, key string, start, stop int64) ([]Member, error)
	RangeByScore(ctx context.Context, key string, min, max float64, offset, count int64) ([]Member, error)
	Remove(ctx context.Context, key string, members ...string) error
	Card(ctx context.Context, key string) (int64, error)
}
