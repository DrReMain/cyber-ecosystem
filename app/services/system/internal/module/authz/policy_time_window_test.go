package authz

import (
	"context"
	"fmt"
	"testing"
	"time"

	"pgregory.net/rapid"

	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"
)

// Properties over the time-window policy: membership is exactly the
// half-open interval test, windows OR within params, the days gate rides the
// Go weekday convention (0 = Sunday), and malformed windows fail closed with
// an error, never a silent deny that reads as a policy verdict.

func hhmm(minute int) string { return fmt.Sprintf("%02d:%02d", minute/60, minute%60) }

func windowParams(windows ...map[string]any) map[string]any {
	list := make([]any, len(windows))
	for i, w := range windows {
		list[i] = w
	}
	return map[string]any{"windows": list}
}

func evaluateTimeWindow(tb rapid.TB, now time.Time, params map[string]any) (bool, error) {
	tb.Helper()
	return timeWindowPolicy{}.Evaluate(context.Background(),
		map[string]any{kauthz.AttrTimeNow: now}, params)
}

// drawTime yields any wall-clock moment; days stay within 1-28 so month
// length never rejects a combination the properties never care about.
func drawTime(t *rapid.T) time.Time {
	return time.Date(
		rapid.IntRange(2000, 2035).Draw(t, "year"),
		time.Month(rapid.IntRange(1, 12).Draw(t, "month")),
		rapid.IntRange(1, 28).Draw(t, "day"),
		rapid.IntRange(0, 23).Draw(t, "hour"),
		rapid.IntRange(0, 59).Draw(t, "minute"),
		0, 0, time.Local)
}

// Property: the verdict is exactly `start <= m < end` for the window —
// boundary inclusion at start, exclusion at end, half-open by construction.
func TestTimeWindowMembershipIsHalfOpen(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		start := rapid.IntRange(0, 1439).Draw(t, "start")
		end := rapid.IntRange(start+1, 1440).Draw(t, "end")
		m := rapid.IntRange(0, 1439).Draw(t, "minute")
		// hhmm(1440) renders the "24:00" end-of-day sentinel by construction.
		now := time.Date(2026, 3, 10, m/60, m%60, 0, 0, time.Local)
		got, err := evaluateTimeWindow(t, now,
			windowParams(map[string]any{"start": hhmm(start), "end": hhmm(end)}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := m >= start && m < end
		if got != want {
			t.Fatalf("window [%d,%d) at minute %d: got %v want %v", start, end, m, got, want)
		}
	})
}

// Property: windows OR within params — a minute in any window allows, a
// minute in none denies; disjoint windows never interfere.
func TestTimeWindowDisjunction(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		s1 := rapid.IntRange(0, 700).Draw(t, "s1")
		e1 := rapid.IntRange(s1+1, 719).Draw(t, "e1")
		s2 := rapid.IntRange(720, 1438).Draw(t, "s2")
		e2 := rapid.IntRange(s2+1, 1440).Draw(t, "e2")
		m := rapid.IntRange(0, 1439).Draw(t, "minute")
		now := time.Date(2026, 3, 10, m/60, m%60, 0, 0, time.Local)
		got, err := evaluateTimeWindow(t, now, windowParams(
			map[string]any{"start": hhmm(s1), "end": hhmm(e1)},
			map[string]any{"start": hhmm(s2), "end": hhmm(e2)},
		))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := (m >= s1 && m < e1) || (m >= s2 && m < e2)
		if got != want {
			t.Fatalf("OR over [%d,%d)∪[%d,%d) at %d: got %v want %v", s1, e1, s2, e2, m, got, want)
		}
	})
}

// Property: the days gate speaks the Go weekday convention (0 = Sunday) —
// today's own weekday admits an all-day window, any other weekday denies it,
// even mid-window. The convention is the historical off-by-one trap.
func TestTimeWindowDaysGateSpeaksGoWeekday(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		now := drawTime(t)
		m := now.Hour()*60 + now.Minute()
		allDay := windowParams(map[string]any{"start": "00:00", "end": "24:00"})
		today := []any{float64(int(now.Weekday()))}
		other := []any{float64((int(now.Weekday()) + 1) % 7)}
		gotToday, err := evaluateTimeWindow(t, now, withDays(allDay, today))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !gotToday {
			t.Fatalf("all-day window with days=[today(%d)] denied at %02d:%02d", now.Weekday(), m/60, m%60)
		}
		gotOther, err := evaluateTimeWindow(t, now, withDays(allDay, other))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotOther {
			t.Fatal("all-day window with days=[tomorrow] allowed — weekday mapping drifted")
		}
	})
}

func withDays(base map[string]any, days []any) map[string]any {
	out := map[string]any{"windows": base["windows"], "days": days}
	return out
}

// Counterexample rows: degenerate inputs fail closed with an error — an
// inverted or empty window is a config bug, and a config bug must be loud,
// never a silent deny that masquerades as a policy verdict.
func TestTimeWindowFailClosed(t *testing.T) {
	now := time.Date(2026, 3, 10, 12, 0, 0, 0, time.Local)
	if _, err := evaluateTimeWindow(t, now,
		windowParams(map[string]any{"start": "17:00", "end": "09:00"})); err == nil {
		t.Fatal("inverted window must error")
	}
	if _, err := evaluateTimeWindow(t, now,
		windowParams(map[string]any{"start": "09:00", "end": "09:00"})); err == nil {
		t.Fatal("empty window must error")
	}
	if _, err := evaluateTimeWindow(t, now,
		windowParams(map[string]any{"start": "25:00", "end": "17:00"})); err == nil {
		t.Fatal("out-of-range wall time must error")
	}
}
