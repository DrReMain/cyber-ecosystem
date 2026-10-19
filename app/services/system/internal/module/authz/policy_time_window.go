package authz

import (
	"context"
	"fmt"
	"slices"
	"time"

	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"
)

func init() { registerPolicy(timeWindowPolicy{}) }

type timeWindowParams struct {
	Windows []timeWindowSpan `json:"windows"`
	Days    []uint32         `json:"days,omitempty"` // 0-6, Sunday first; empty = every day
}

type timeWindowSpan struct {
	Start string `json:"start"` // HH:MM
	End   string `json:"end"`
}

type timeWindowPolicy struct{}

func (timeWindowPolicy) Kind() string { return kauthz.PolicyKindTimeWindow }

func (timeWindowPolicy) Attrs() []string { return []string{kauthz.AttrTimeNow} }

func (timeWindowPolicy) Evaluate(_ context.Context, attrs, params map[string]any) (bool, error) {
	now, ok := attrs[kauthz.AttrTimeNow].(time.Time)
	if !ok {
		return false, fmt.Errorf("attr %q is not a time", kauthz.AttrTimeNow)
	}
	var p timeWindowParams
	if err := decodeParams(params, &p); err != nil {
		return false, err
	}
	if len(p.Days) > 0 && !slices.ContainsFunc(p.Days, func(d uint32) bool {
		return int(d) == int(now.Weekday())
	}) {
		return false, nil
	}
	minute := now.Hour()*60 + now.Minute()
	for _, w := range p.Windows {
		start, err := minuteOfDay(w.Start)
		if err != nil {
			return false, err
		}
		end, err := minuteOfDayEnd(w.End)
		if err != nil {
			return false, err
		}
		// Same-day half-open [start, end); end "24:00" runs through the end
		// of the day. Windows OR within params (two linked policies would
		// AND — segment-OR is why windows is a list).
		if start >= end {
			return false, fmt.Errorf("window %q-%q must satisfy start < end", w.Start, w.End)
		}
		if minute >= start && minute < end {
			return true, nil
		}
	}
	return false, nil
}

func minuteOfDay(s string) (int, error) {
	t, err := time.Parse("15:04", s)
	if err != nil {
		return 0, fmt.Errorf("invalid wall time %q: want HH:MM", s)
	}
	return t.Hour()*60 + t.Minute(), nil
}

func minuteOfDayEnd(s string) (int, error) {
	// "24:00" is the end-of-day sentinel (1440) — the wall clock itself has
	// no twenty-fourth hour.
	if s == "24:00" {
		return 24 * 60, nil
	}
	return minuteOfDay(s)
}
