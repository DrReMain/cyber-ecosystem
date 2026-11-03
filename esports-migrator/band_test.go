package main

import (
	"testing"
	"time"
)

// The band derivation is a verbatim copy of the live write path's, so its
// boundaries are pinned here: drift between the two copies would store
// wrong cohort bands silently. now is anchored to 2026-12-01 — the month,
// not the day, flips a tier, so a December birth has its birthday "this
// month" while a January 2015 birth is still 11.
func TestAgeBandOfBoundaries(t *testing.T) {
	now := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		year, month int32
		want        string
	}{
		{2015, 1, "under12"}, // 11 — one month short of 12
		{2014, 12, "12-14"},  // turns 12 this month
		{2012, 12, "12-14"},  // 14 — tier's last year
		{2011, 12, "15-17"},  // turns 15 this month
		{2009, 12, "15-17"},  // 17
		{2008, 12, "18-22"},  // turns 18 this month
		{2004, 12, "18-22"},  // 22 — tier's last year
		{2003, 12, "23-25"},  // turns 23 this month
		{2001, 12, "23-25"},  // 25
		{2000, 12, "26-29"},  // turns 26 this month
		{1997, 12, "26-29"},  // 29
		{1996, 12, "30-40"},  // turns 30 this month
		{1986, 12, "30-40"},  // 40 — tier's last year
		{1985, 12, "over40"}, // turns 41 this month
		{2027, 1, "under12"}, // future birth clamps to 0, never negative
	}
	for _, tc := range cases {
		got := ageBandOf(&tc.year, &tc.month, now)
		if got == nil || *got != tc.want {
			t.Errorf("ageBandOf(%d, %d) = %v, want %q", tc.year, tc.month, got, tc.want)
		}
	}
	if b := ageBandOf(nil, nil, now); b != nil {
		t.Errorf("ageBandOf(nil, nil) = %v, want nil (incomplete profile keeps the slot empty)", *b)
	}
}
