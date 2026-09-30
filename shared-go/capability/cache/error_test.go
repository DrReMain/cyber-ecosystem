package cache

import (
	"errors"
	"testing"

	kratoserrors "github.com/go-kratos/kratos/v3/errors"
)

// fakeErrs builds a CacheDefaultError with distinct *errors.Error per slot so
// the test can assert which branch HandleCacheError took. No app error-proto
// dependency needed: the mapping logic only moves *errors.Error around.
func fakeErrs() *CacheDefaultError {
	mk := func(reason string) *kratoserrors.Error {
		return kratoserrors.New(500, reason, "")
	}
	return &CacheDefaultError{
		CacheMiss:       mk("CACHE_MISS"),
		KeyNotFound:     mk("KEY_NOT_FOUND"),
		SessionNotFound: mk("SESSION_NOT_FOUND"),
		QuotaExceeded:   mk("QUOTA_EXCEEDED"),
		InvalidArgument: mk("INVALID_ARGUMENT"),
		LockNotAcquired: mk("LOCK_NOT_ACQUIRED"),
		Unavailable:     mk("UNAVAILABLE"),
	}
}

func TestHandleCacheError(t *testing.T) {
	errs := fakeErrs()
	cases := []struct {
		in   error
		want string
	}{
		{ErrCacheMiss, "CACHE_MISS"},
		{ErrKeyNotFound, "KEY_NOT_FOUND"},
		{ErrSessionNotFound, "SESSION_NOT_FOUND"},
		{ErrQuotaExceeded, "QUOTA_EXCEEDED"},
		{ErrInvalidArgument, "INVALID_ARGUMENT"},
		{ErrLockNotAcquired, "LOCK_NOT_ACQUIRED"},
		{errors.New("something broke"), "UNAVAILABLE"}, // unknown → Unavailable
	}
	for _, c := range cases {
		got := HandleCacheError(c.in, errs)
		if kratoserrors.Reason(got) != c.want {
			t.Errorf("HandleCacheError(%v): reason=%s, want %s", c.in, kratoserrors.Reason(got), c.want)
		}
	}
}

// wrapped sentinel must still match (errors.Is unwraps), so the mapping works
// for errors returned with %w.
func TestHandleCacheErrorWrapped(t *testing.T) {
	errs := fakeErrs()
	wrapped := errors.Join(ErrCacheMiss, errors.New("ctx"))
	got := HandleCacheError(wrapped, errs)
	if kratoserrors.Reason(got) != "CACHE_MISS" {
		t.Errorf("wrapped ErrCacheMiss: reason=%s, want CACHE_MISS", kratoserrors.Reason(got))
	}
}

// TestHandleCacheErrorUnavailableNil pins the nil-safety of the default branch:
// when Unavailable is unset, an unknown error is returned unchanged (raw
// pass-through) instead of nil-derefing.
func TestHandleCacheErrorUnavailableNil(t *testing.T) {
	errs := fakeErrs()
	errs.Unavailable = nil
	raw := errors.New("boom")
	if got := HandleCacheError(raw, errs); !errors.Is(got, raw) {
		t.Fatalf("Unavailable=nil: want raw pass-through, got %v", got)
	}
}

// TestValidateCacheDefaultError asserts a nil required sentinel slot is
// rejected (so HandleCacheError surfaces misconfig instead of panicking),
// while a nil Unavailable is accepted (it is optional).
func TestValidateCacheDefaultError(t *testing.T) {
	if err := ValidateCacheDefaultError(fakeErrs()); err != nil {
		t.Fatalf("fully populated: unexpected %v", err)
	}
	bad := fakeErrs()
	bad.CacheMiss = nil
	if err := ValidateCacheDefaultError(bad); err == nil {
		t.Fatal("nil CacheMiss slot: want validation error, got nil")
	}
	opt := fakeErrs()
	opt.Unavailable = nil // optional
	if err := ValidateCacheDefaultError(opt); err != nil {
		t.Fatalf("nil Unavailable: unexpected %v", err)
	}
}
