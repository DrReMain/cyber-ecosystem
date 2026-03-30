package memory

import (
	"context"
	"encoding/base64"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"cyber-ecosystem/shared-go/cache"
)

type sortedSet struct {
	m *Memory
}

func newSortedSet(m *Memory) *sortedSet {
	return &sortedSet{m: m}
}

type zsetEntry struct {
	members map[string]float64
}

func (z *zsetEntry) add(member string, score float64) {
	z.members[member] = score
}

func (z *zsetEntry) getScore(member string) (float64, bool) {
	score, exists := z.members[member]
	return score, exists
}

func (z *zsetEntry) remove(member string) bool {
	if _, ok := z.members[member]; ok {
		delete(z.members, member)
		return true
	}
	return false
}

func (z *zsetEntry) rank(member string, ascending bool) (int64, bool) {
	type memberScore struct {
		member string
		score  float64
	}

	var sorted []memberScore
	for m, s := range z.members {
		sorted = append(sorted, memberScore{m, s})
	}

	sort.Slice(sorted, func(i, j int) bool {
		if ascending {
			return sorted[i].score < sorted[j].score
		}
		return sorted[i].score > sorted[j].score
	})

	for i, ms := range sorted {
		if ms.member == member {
			return int64(i), true
		}
	}
	return 0, false
}

func (z *zsetEntry) rangeByRank(start, stop int64, ascending bool) []cache.Member {
	type memberScore struct {
		member string
		score  float64
	}

	var sorted []memberScore
	for m, s := range z.members {
		sorted = append(sorted, memberScore{m, s})
	}

	sort.Slice(sorted, func(i, j int) bool {
		if ascending {
			return sorted[i].score < sorted[j].score
		}
		return sorted[i].score > sorted[j].score
	})

	if start < 0 {
		start = int64(len(sorted)) + start
	}
	if stop < 0 {
		stop = int64(len(sorted)) + stop
	}
	if start > int64(len(sorted)) {
		return nil
	}
	if stop > int64(len(sorted))-1 {
		stop = int64(len(sorted)) - 1
	}
	if start > stop {
		return nil
	}

	result := make([]cache.Member, 0, stop-start+1)
	for i := start; i <= stop; i++ {
		result = append(result, cache.Member{
			Member: sorted[i].member,
			Score:  sorted[i].score,
		})
	}
	return result
}

func (s *sortedSet) Add(ctx context.Context, key string, members ...cache.Member) error {
	opStart := time.Now()
	if len(members) == 0 {
		return nil
	}
	if err := cache.ValidateKey(key); err != nil {
		s.m.logOperation(ctx, "zadd", "key", key, "members", len(members), "duration", time.Since(opStart), "error", err)
		return err
	}
	if err := cache.ValidateMembers(members); err != nil {
		s.m.logOperation(ctx, "zadd", "key", key, "members", len(members), "duration", time.Since(opStart), "error", err)
		return err
	}

	var err error

	s.m.traceOperationWithArgs(ctx, "ZADD", &cache.OperationArgs{Key: key, Members: members}, func(ctx context.Context) error {
		s.m.mu.Lock()
		defer s.m.mu.Unlock()

		e, ok := s.m.data[key]
		if !ok || e.isExpired() {
			e = &entry{
				value: make([]byte, 0),
			}
		}

		ze := deserializeZSet(e.value)
		for _, m := range members {
			ze.add(m.Member, m.Score)
		}

		data := serializeZSet(ze)
		e.value = data
		s.m.data[key] = e
		return nil
	})

	s.m.logOperation(ctx, "zadd", "key", key, "members", len(members), "duration", time.Since(opStart), "error", err)
	return err
}

func (s *sortedSet) Incr(ctx context.Context, key string, member string, delta float64) (float64, error) {
	opStart := time.Now()
	if err := cache.ValidateKey(key); err != nil {
		s.m.logOperation(ctx, "zincrby", "key", key, "member", member, "delta", delta, "duration", time.Since(opStart), "error", err)
		return 0, err
	}
	if err := cache.ValidateMember(member); err != nil {
		s.m.logOperation(ctx, "zincrby", "key", key, "member", member, "delta", delta, "duration", time.Since(opStart), "error", err)
		return 0, err
	}

	var newScore float64
	var err error

	s.m.traceOperationWithArgs(ctx, "ZINCRBY", &cache.OperationArgs{Key: key, Member: member, Delta: delta}, func(ctx context.Context) error {
		s.m.mu.Lock()
		defer s.m.mu.Unlock()

		e, ok := s.m.data[key]
		if !ok || e.isExpired() {
			e = &entry{value: make([]byte, 0)}
		}

		ze := deserializeZSet(e.value)
		oldScore, _ := ze.getScore(member)
		newScore = oldScore + delta
		ze.add(member, newScore)

		e.value = serializeZSet(ze)
		s.m.data[key] = e
		return nil
	})

	s.m.logOperation(ctx, "zincrby", "key", key, "member", member, "delta", delta, "newScore", newScore, "duration", time.Since(opStart), "error", err)
	return newScore, err
}

func (s *sortedSet) Rank(ctx context.Context, key string, member string) (int64, error) {
	opStart := time.Now()
	if err := cache.ValidateKey(key); err != nil {
		s.m.logOperation(ctx, "zrank", "key", key, "member", member, "duration", time.Since(opStart), "error", err)
		return 0, err
	}
	if err := cache.ValidateMember(member); err != nil {
		s.m.logOperation(ctx, "zrank", "key", key, "member", member, "duration", time.Since(opStart), "error", err)
		return 0, err
	}

	var rank int64
	var err error

	s.m.traceOperationWithArgs(ctx, "ZRANK", &cache.OperationArgs{Key: key, Member: member}, func(ctx context.Context) error {
		s.m.mu.RLock()
		defer s.m.mu.RUnlock()

		e, ok := s.m.data[key]
		if !ok || e.isExpired() {
			err = cache.ErrKeyNotFound
			return nil
		}

		ze := deserializeZSet(e.value)
		rankVal, ok := ze.rank(member, true)
		if !ok {
			err = cache.ErrKeyNotFound
			return nil
		}
		rank = rankVal
		return nil
	})

	s.m.logOperation(ctx, "zrank", "key", key, "member", member, "rank", rank, "duration", time.Since(opStart), "error", err)
	return rank, err
}

func (s *sortedSet) RevRank(ctx context.Context, key string, member string) (int64, error) {
	opStart := time.Now()
	if err := cache.ValidateKey(key); err != nil {
		s.m.logOperation(ctx, "zrevrank", "key", key, "member", member, "duration", time.Since(opStart), "error", err)
		return 0, err
	}
	if err := cache.ValidateMember(member); err != nil {
		s.m.logOperation(ctx, "zrevrank", "key", key, "member", member, "duration", time.Since(opStart), "error", err)
		return 0, err
	}

	var rank int64
	var err error

	s.m.traceOperationWithArgs(ctx, "ZREVRANK", &cache.OperationArgs{Key: key, Member: member}, func(ctx context.Context) error {
		s.m.mu.RLock()
		defer s.m.mu.RUnlock()

		e, ok := s.m.data[key]
		if !ok || e.isExpired() {
			err = cache.ErrKeyNotFound
			return nil
		}

		ze := deserializeZSet(e.value)
		r, ok := ze.rank(member, false)
		if !ok {
			err = cache.ErrKeyNotFound
			return nil
		}
		rank = r
		return nil
	})

	s.m.logOperation(ctx, "zrevrank", "key", key, "member", member, "rank", rank, "duration", time.Since(opStart), "error", err)
	return rank, err
}

func (s *sortedSet) Score(ctx context.Context, key string, member string) (float64, error) {
	opStart := time.Now()
	if err := cache.ValidateKey(key); err != nil {
		s.m.logOperation(ctx, "zscore", "key", key, "member", member, "duration", time.Since(opStart), "error", err)
		return 0, err
	}
	if err := cache.ValidateMember(member); err != nil {
		s.m.logOperation(ctx, "zscore", "key", key, "member", member, "duration", time.Since(opStart), "error", err)
		return 0, err
	}

	var score float64
	var err error

	s.m.traceOperationWithArgs(ctx, "ZSCORE", &cache.OperationArgs{Key: key, Member: member}, func(ctx context.Context) error {
		s.m.mu.RLock()
		defer s.m.mu.RUnlock()

		e, ok := s.m.data[key]
		if !ok || e.isExpired() {
			err = cache.ErrKeyNotFound
			return nil
		}

		ze := deserializeZSet(e.value)
		memberScore, ok := ze.getScore(member)
		if !ok {
			err = cache.ErrKeyNotFound
			return nil
		}
		score = memberScore
		return nil
	})

	s.m.logOperation(ctx, "zscore", "key", key, "member", member, "score", score, "duration", time.Since(opStart), "error", err)
	return score, err
}

func (s *sortedSet) Range(ctx context.Context, key string, start, stop int64) ([]cache.Member, error) {
	opStart := time.Now()
	if err := cache.ValidateKey(key); err != nil {
		s.m.logOperation(ctx, "zrange", "key", key, "start", start, "stop", stop, "duration", time.Since(opStart), "error", err)
		return nil, err
	}

	var result []cache.Member
	var err error

	s.m.traceOperationWithArgs(ctx, "ZRANGE", &cache.OperationArgs{Key: key}, func(ctx context.Context) error {
		s.m.mu.RLock()
		defer s.m.mu.RUnlock()

		e, ok := s.m.data[key]
		if !ok || e.isExpired() {
			result = make([]cache.Member, 0)
			return nil
		}

		ze := deserializeZSet(e.value)
		result = ze.rangeByRank(start, stop, true)
		return nil
	})

	s.m.logOperation(ctx, "zrange", "key", key, "start", start, "stop", stop, "count", len(result), "duration", time.Since(opStart), "error", err)
	return result, err
}

func (s *sortedSet) RevRange(ctx context.Context, key string, start, stop int64) ([]cache.Member, error) {
	opStart := time.Now()
	if err := cache.ValidateKey(key); err != nil {
		s.m.logOperation(ctx, "zrevrange", "key", key, "start", start, "stop", stop, "duration", time.Since(opStart), "error", err)
		return nil, err
	}

	var result []cache.Member
	var err error

	s.m.traceOperationWithArgs(ctx, "ZREVRANGE", &cache.OperationArgs{Key: key}, func(ctx context.Context) error {
		s.m.mu.RLock()
		defer s.m.mu.RUnlock()

		e, ok := s.m.data[key]
		if !ok || e.isExpired() {
			result = make([]cache.Member, 0)
			return nil
		}

		ze := deserializeZSet(e.value)
		result = ze.rangeByRank(start, stop, false)
		return nil
	})

	s.m.logOperation(ctx, "zrevrange", "key", key, "start", start, "stop", stop, "count", len(result), "duration", time.Since(opStart), "error", err)
	return result, err
}

func (s *sortedSet) Remove(ctx context.Context, key string, members ...string) error {
	opStart := time.Now()
	if len(members) == 0 {
		return nil
	}
	if err := cache.ValidateKey(key); err != nil {
		s.m.logOperation(ctx, "zrem", "key", key, "members", len(members), "duration", time.Since(opStart), "error", err)
		return err
	}
	if err := cache.ValidateMemberStrings(members...); err != nil {
		s.m.logOperation(ctx, "zrem", "key", key, "members", len(members), "duration", time.Since(opStart), "error", err)
		return err
	}

	var err error

	s.m.traceOperationWithArgs(ctx, "ZREM", &cache.OperationArgs{Key: key}, func(ctx context.Context) error {
		s.m.mu.Lock()
		defer s.m.mu.Unlock()

		e, ok := s.m.data[key]
		if !ok || e.isExpired() {
			return nil
		}

		ze := deserializeZSet(e.value)
		for _, m := range members {
			ze.remove(m)
		}
		e.value = serializeZSet(ze)
		return nil
	})

	s.m.logOperation(ctx, "zrem", "key", key, "members", len(members), "duration", time.Since(opStart), "error", err)
	return err
}

func serializeZSet(ze *zsetEntry) []byte {
	// Format: base64(member)|score|base64(member)|score|...
	// Using | as separator between entries, : as separator within entry
	data := make([]byte, 0, len(ze.members)*32)
	encoder := base64.RawURLEncoding
	for member, score := range ze.members {
		encodedMember := make([]byte, encoder.EncodedLen(len(member)))
		encoder.Encode(encodedMember, []byte(member))
		data = append(data, encodedMember...)
		data = append(data, ':')
		data = append(data, []byte(fmt.Sprintf("%f", score))...)
		data = append(data, '|')
	}
	return data
}

func deserializeZSet(data []byte) *zsetEntry {
	ze := &zsetEntry{members: make(map[string]float64)}
	if len(data) == 0 {
		return ze
	}

	decoder := base64.RawURLEncoding
	parts := strings.Split(string(data), "|")
	for _, part := range parts {
		if part == "" {
			continue
		}
		idx := strings.Index(part, ":")
		if idx == -1 {
			continue
		}
		encodedMember := part[:idx]
		scoreStr := part[idx+1:]
		score, err := strconv.ParseFloat(scoreStr, 64)
		if err != nil {
			continue
		}
		decodedMemberBytes := make([]byte, decoder.DecodedLen(len(encodedMember)))
		n, err := decoder.Decode(decodedMemberBytes, []byte(encodedMember))
		if err != nil {
			continue
		}
		ze.members[string(decodedMemberBytes[:n])] = score
	}
	return ze
}
