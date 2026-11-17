package authz

import (
	"context"
	"testing"
	"time"

	"pgregory.net/rapid"

	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"
)

// Properties over the calendar policy: an override on today always wins
// (holiday denies a WEEKDAYS base, makeup day allows), and without a hit the
// base speaks — WEEKDAYS is exactly Monday..Friday in the Go weekday
// convention, ALL and NONE are constants, an unknown base errors.

func evaluateCalendar(tb rapid.TB, now time.Time, params map[string]any) (bool, error) {
	tb.Helper()
	return calendarPolicy{}.Evaluate(context.Background(),
		map[string]any{kauthz.AttrTimeNow: now}, params)
}

// Property: an override dated today wins over every base — the verdict is
// the override's own allowed flag, independent of base and weekday.
func TestCalendarOverrideWins(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		now := drawTime(t)
		base := rapid.SampledFrom([]string{kauthz.CalendarBaseAll, kauthz.CalendarBaseWeekdays, kauthz.CalendarBaseNone}).Draw(t, "base")
		allowed := rapid.Bool().Draw(t, "allowed")
		got, err := evaluateCalendar(t, now, map[string]any{
			"base": base,
			"overrides": []any{map[string]any{
				"date":    now.Format("2006-01-02"),
				"allowed": allowed,
			}},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != allowed {
			t.Fatalf("override (%q,%v) over base %q on %s: got %v", now.Format("2006-01-02"), allowed, base, now.Weekday(), got)
		}
	})
}

// Property: without a hit, the base is exactly its semantics — WEEKDAYS is
// Monday..Friday, ALL is always true, NONE is always false. No date surprises.
func TestCalendarBaseSemantics(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		now := drawTime(t)
		cases := []struct {
			base string
			want bool
		}{
			{kauthz.CalendarBaseAll, true},
			{kauthz.CalendarBaseNone, false},
			{kauthz.CalendarBaseWeekdays, now.Weekday() >= time.Monday && now.Weekday() <= time.Friday},
		}
		for _, tc := range cases {
			got, err := evaluateCalendar(t, now, map[string]any{"base": tc.base})
			if err != nil {
				t.Fatalf("base %q: unexpected error: %v", tc.base, err)
			}
			if got != tc.want {
				t.Fatalf("base %q on %s(%d): got %v want %v", tc.base, now.Format("2006-01-02"), now.Weekday(), got, tc.want)
			}
		}
	})
}

// Counterexample row: an unknown base is a config bug and must be loud —
// fail closed with an error, never a silent deny.
func TestCalendarUnknownBaseFailsClosed(t *testing.T) {
	if _, err := evaluateCalendar(t, time.Date(2026, 3, 10, 12, 0, 0, 0, time.Local),
		map[string]any{"base": "SOMETIMES"}); err == nil {
		t.Fatal("unknown base must error")
	}
}
