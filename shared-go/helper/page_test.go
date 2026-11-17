package helper

import (
	"context"
	"math"
	"testing"

	kratoserrors "github.com/go-kratos/kratos/v3/errors"
	"pgregory.net/rapid"

	commonv1 "cyber-ecosystem/gen/go/cyber/shared/common/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
)

// In-package bounds guard (self-containment law: helper copies out without
// utils.ConvNum) — rapid draws are range-bounded, the clamp makes the
// narrowing safety static, not assumed.
func toI32(v int) int32 {
	if v > math.MaxInt32 {
		return math.MaxInt32
	}
	if v < math.MinInt32 {
		return math.MinInt32
	}
	return int32(v)
}

// fakeQuery satisfies QueryPaginator over an in-memory total: Offset/Limit
// record the applied window, Count reports the canned total from a clone so
// the original builder's recorded window survives the count.
type fakeQuery struct {
	total  int
	offset int
	limit  int
}

func (f *fakeQuery) Offset(o int) *fakeQuery            { f.offset = o; return f }
func (f *fakeQuery) Limit(l int) *fakeQuery             { f.limit = l; return f }
func (f *fakeQuery) Clone() *fakeQuery                  { c := *f; return &c }
func (f *fakeQuery) Count(context.Context) (int, error) { return f.total, nil }

var paginationErr = errorspb.ErrorGeneralErrorPaginationInvalidArgument("")

// Property: consecutive pages partition the row space — no gap, no overlap.
func TestPageWindowPartitions(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		total := rapid.IntRange(0, 5000).Draw(t, "total")
		size := rapid.IntRange(1, DefaultPageSizeMax).Draw(t, "pageSize")
		pages := (total + size - 1) / size
		covered := 0
		for p := 1; p <= pages; p++ {
			offset, limit := PageWindow(&commonv1.PageRequest{PageNo: new(toI32(p)), PageSize: new(toI32(size))})
			if offset != covered {
				t.Fatalf("page %d starts at %d, expected %d — gap or overlap", p, offset, covered)
			}
			covered += limit
		}
		if pages > 0 && covered < total {
			t.Fatalf("pages cover %d of %d rows", covered, total)
		}
	})
}

// Property: the two resolvers never drift — PageWindow (in-memory faces) and
// ApplyPagination (query faces) answer the same (offset, limit) for every
// request under the shared cap. The "can't drift" claim in page.go is law;
// this is its machine check.
func TestPageWindowMatchesApplyPagination(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		pageNo := rapid.IntRange(1, 1000).Draw(t, "pageNo")
		pageSize := rapid.IntRange(1, 10000).Draw(t, "pageSize")
		total := rapid.IntRange(0, 100000).Draw(t, "total")
		req := &commonv1.PageRequest{PageNo: new(toI32(pageNo)), PageSize: new(toI32(pageSize))}
		q := &fakeQuery{total: total}
		gotTotal, qOffset, qLimit, err := ApplyPagination(context.Background(), q, req,
			NewPageConfig(DefaultPageSize, DefaultPageSizeMax), paginationErr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		wOffset, wLimit := PageWindow(req)
		if qOffset != wOffset || qLimit != wLimit {
			t.Fatalf("resolvers drifted: query=(%d,%d) window=(%d,%d)", qOffset, qLimit, wOffset, wLimit)
		}
		if gotTotal != total || q.offset != qOffset || q.limit != qLimit {
			t.Fatalf("query window not applied: total=%d applied=(%d,%d)", gotTotal, q.offset, q.limit)
		}
	})
}

// Property: More, PageNo, and PageSize in the response derive exactly from
// the window — for every (total, window) pair the response is a function of
// the inputs, never of the caller's expectations.
func TestBuildPageResponseInvariants(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		total := rapid.IntRange(0, 100000).Draw(t, "total")
		pageNo := rapid.IntRange(1, 1000).Draw(t, "pageNo")
		size := rapid.IntMin(1).Draw(t, "pageSize")
		offset, limit := PageWindow(&commonv1.PageRequest{PageNo: new(toI32(pageNo)), PageSize: new(toI32(size))})
		res := BuildPageResponse(total, offset, limit)
		if int(res.PageNo) != offset/limit+1 {
			t.Fatalf("PageNo %d != window page %d", res.PageNo, offset/limit+1)
		}
		if int(res.PageSize) != limit {
			t.Fatalf("PageSize %d != limit %d", res.PageSize, limit)
		}
		if int(res.Total) != total {
			t.Fatalf("Total %d != %d", res.Total, total)
		}
		if res.More != (total > offset+limit) {
			t.Fatalf("More %v wrong at total=%d offset=%d limit=%d", res.More, total, offset, limit)
		}
	})
}

// Property: defaults and the cap are clamps, not suggestions — every
// non-positive input maps to the default, every oversized size maps to the
// cap, and pageNo is never below 1.
func TestPageWindowClamps(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		badNo := rapid.IntRange(-1000, 0).Draw(t, "badPageNo")
		badSize := rapid.IntRange(-1000, 0).Draw(t, "badPageSize")
		offset, limit := PageWindow(&commonv1.PageRequest{PageNo: new(toI32(badNo)), PageSize: new(toI32(badSize))})
		if offset != 0 || limit != DefaultPageSize {
			t.Fatalf("non-positive inputs must clamp to defaults, got (%d,%d)", offset, limit)
		}
		big := rapid.IntRange(DefaultPageSizeMax+1, 1<<20).Draw(t, "bigPageSize")
		_, limit = PageWindow(&commonv1.PageRequest{PageNo: new(toI32(1)), PageSize: new(toI32(big))})
		if limit != DefaultPageSizeMax {
			t.Fatalf("oversized size must clamp to %d, got %d", DefaultPageSizeMax, limit)
		}
	})
}

// All=true is lawful only on unbounded configs: a bounded config refuses it
// with the shared pagination reason (the paginate-all trap), an unbounded
// one returns the full count with no window.
func TestApplyPaginationAllChannel(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		total := rapid.IntRange(0, 100000).Draw(t, "total")
		q := &fakeQuery{total: total}
		_, _, _, err := ApplyPagination(context.Background(), q,
			&commonv1.PageRequest{All: new(true)}, NewPageConfig(DefaultPageSize, 100), paginationErr)
		if err == nil {
			t.Fatal("All on a bounded config must refuse")
		}
		if !errorspb.IsGeneralErrorPaginationInvalidArgument(err) {
			t.Fatalf("refusal must carry the shared pagination reason, got %v", err)
		}
		gotTotal, offset, limit, err := ApplyPagination(context.Background(), q,
			&commonv1.PageRequest{All: new(true)}, NewPageConfig(DefaultPageSize, 0), paginationErr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotTotal != total || offset != 0 || limit != 0 {
			t.Fatalf("All on unbounded must return (total,0,0), got (%d,%d,%d)", gotTotal, offset, limit)
		}
	})
}

// Counterexample rows: the degenerate inputs that must fail closed, not
// panic or default silently.
func TestApplyPaginationRefuses(t *testing.T) {
	q := &fakeQuery{total: 10}
	if _, _, _, err := ApplyPagination(context.Background(), q, nil, NewPageConfig(10, 100), paginationErr); err == nil {
		t.Fatal("nil request must refuse")
	}
	if _, _, _, err := ApplyPagination(context.Background(), q, &commonv1.PageRequest{}, NewPageConfig(10, 100), nil); err == nil {
		t.Fatal("nil catalog error must refuse")
	}
	if _, _, _, err := ApplyPagination(context.Background(), q, &commonv1.PageRequest{}, NewPageConfig(10, 100), paginationErr); err != nil {
		t.Fatalf("zero-value request must take defaults, got %v", err)
	}
	if !kratoserrors.IsBadRequest(errorspb.ErrorGeneralErrorPaginationInvalidArgument("")) {
		t.Fatal("the pagination error must render as 400-class on the wire")
	}
}
