package authz

import (
	"context"
	"fmt"
	"time"

	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"
)

func init() { registerPolicy(calendarPolicy{}) }

type calendarParams struct {
	Base      string             `json:"base"` // kauthz.CalendarBase*
	Overrides []calendarOverride `json:"overrides,omitempty"`
}

type calendarOverride struct {
	Date    string `json:"date"` // YYYY-MM-DD, server-local
	Allowed bool   `json:"allowed"`
}

type calendarPolicy struct{}

func (calendarPolicy) Kind() string { return kauthz.PolicyKindCalendar }

func (calendarPolicy) Attrs() []string { return []string{kauthz.AttrTimeNow} }

func (calendarPolicy) Evaluate(_ context.Context, attrs, params map[string]any) (bool, error) {
	now, ok := attrs[kauthz.AttrTimeNow].(time.Time)
	if !ok {
		return false, fmt.Errorf("attr %q is not a time", kauthz.AttrTimeNow)
	}
	var p calendarParams
	if err := decodeParams(params, &p); err != nil {
		return false, err
	}
	today := now.Format("2006-01-02")
	// Overrides win over the periodic base: holidays deny on a WEEKDAYS
	// calendar, makeup workdays allow on the holiday they bridge.
	for _, o := range p.Overrides {
		if o.Date == today {
			return o.Allowed, nil
		}
	}
	switch p.Base {
	case kauthz.CalendarBaseAll:
		return true, nil
	case kauthz.CalendarBaseWeekdays:
		return now.Weekday() >= time.Monday && now.Weekday() <= time.Friday, nil
	case kauthz.CalendarBaseNone:
		return false, nil
	default:
		return false, fmt.Errorf("unknown base %q", p.Base)
	}
}
