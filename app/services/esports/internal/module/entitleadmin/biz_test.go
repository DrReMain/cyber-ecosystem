package entitleadmin

import (
	"testing"
	"time"

	"cyber-ecosystem/app/services/esports/internal/module/entitle"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

func ts(t *testing.T, s string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatalf("bad fixture time %q: %v", s, err)
	}
	return parsed
}

func at(t *testing.T, s string) *time.Time {
	t.Helper()
	parsed := ts(t, s)
	return &parsed
}

// periodStart must mirror the grant seam (Grant's MembershipWindow read):
// the furthest window still live at the row's birth sets the start; a prior
// row dead before the birth (terminated, or its window already closed)
// leaves the period starting at the row's own creation.
func TestPeriodStart(t *testing.T) {
	cases := []struct {
		name string
		run  []*entitle.EntitlementRow
		i    int
		want time.Time
	}{
		{
			name: "first period starts at its own creation",
			run: []*entitle.EntitlementRow{
				{CreatedAt: ts(t, "2026-10-05T10:00:00+08:00"), ExpiresAt: at(t, "2026-11-05T10:00:00+08:00"), Status: shared.StatusActive},
			},
			i:    0,
			want: ts(t, "2026-10-05T10:00:00+08:00"),
		},
		{
			name: "tiled renewal starts at the live prior's expiry",
			run: []*entitle.EntitlementRow{
				{CreatedAt: ts(t, "2026-10-05T10:00:00+08:00"), ExpiresAt: at(t, "2026-12-05T10:00:00+08:00"), Status: shared.StatusActive},
				{CreatedAt: ts(t, "2026-11-20T10:00:00+08:00"), ExpiresAt: at(t, "2027-01-04T10:00:00+08:00"), Status: shared.StatusActive},
			},
			i:    1,
			want: ts(t, "2026-12-05T10:00:00+08:00"),
		},
		{
			// The rehearsal shape (cutover-watch ring 7): a terminated prior
			// showed its dead expiry as the new purchase's window head.
			name: "prior terminated before the birth — the period starts at its own creation",
			run: []*entitle.EntitlementRow{
				{CreatedAt: ts(t, "2026-10-05T10:00:00+08:00"), ExpiresAt: at(t, "2026-12-05T16:10:00+08:00"), Status: shared.StatusTerminated, TerminatedAt: at(t, "2026-11-05T14:15:00+08:00")},
				{CreatedAt: ts(t, "2026-11-05T14:20:00+08:00"), ExpiresAt: at(t, "2026-12-05T14:20:00+08:00"), Status: shared.StatusActive},
			},
			i:    1,
			want: ts(t, "2026-11-05T14:20:00+08:00"),
		},
		{
			// The window read drops expired rows too — a lapsed renewal buys
			// forward from purchase, never backfilled from the old expiry.
			name: "prior window already closed at the birth — the period starts at its own creation",
			run: []*entitle.EntitlementRow{
				{CreatedAt: ts(t, "2026-10-01T10:00:00+08:00"), ExpiresAt: at(t, "2026-11-01T10:00:00+08:00"), Status: shared.StatusActive},
				{CreatedAt: ts(t, "2026-11-05T10:00:00+08:00"), ExpiresAt: at(t, "2026-12-05T10:00:00+08:00"), Status: shared.StatusActive},
			},
			i:    1,
			want: ts(t, "2026-11-05T10:00:00+08:00"),
		},
		{
			// Termination is dated: a row terminated after this birth was
			// live when the seam read it and still sets the start.
			name: "prior terminated after the birth still sets the start",
			run: []*entitle.EntitlementRow{
				{CreatedAt: ts(t, "2026-10-05T10:00:00+08:00"), ExpiresAt: at(t, "2026-12-05T10:00:00+08:00"), Status: shared.StatusTerminated, TerminatedAt: at(t, "2026-12-10T10:00:00+08:00")},
				{CreatedAt: ts(t, "2026-11-20T10:00:00+08:00"), ExpiresAt: at(t, "2027-01-04T10:00:00+08:00"), Status: shared.StatusTerminated, TerminatedAt: at(t, "2026-12-10T10:00:00+08:00")},
			},
			i:    1,
			want: ts(t, "2026-12-05T10:00:00+08:00"),
		},
		{
			// Settle can land a later-created row with an earlier expiry than
			// a tiled manual window: the seam took the furthest live expiry,
			// not the adjacent row's.
			name: "furthest of the windows live at the birth wins, not the adjacent row",
			run: []*entitle.EntitlementRow{
				{CreatedAt: ts(t, "2026-10-05T10:00:00+08:00"), ExpiresAt: at(t, "2026-12-20T10:00:00+08:00"), Status: shared.StatusActive},
				{CreatedAt: ts(t, "2026-11-18T10:00:00+08:00"), ExpiresAt: at(t, "2026-12-10T10:00:00+08:00"), Status: shared.StatusActive},
				{CreatedAt: ts(t, "2026-11-19T10:00:00+08:00"), ExpiresAt: at(t, "2027-01-18T10:00:00+08:00"), Status: shared.StatusActive},
			},
			i:    2,
			want: ts(t, "2026-12-20T10:00:00+08:00"),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := periodStart(tc.run, tc.i)
			if !got.Equal(tc.want) {
				t.Fatalf("periodStart(run, %d) = %s, want %s", tc.i, got.Format(time.RFC3339), tc.want.Format(time.RFC3339))
			}
		})
	}
}
