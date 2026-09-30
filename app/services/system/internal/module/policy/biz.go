package policy

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"time"

	"github.com/go-kratos/kratos/v3/errors"

	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"

	commonpb "cyber-ecosystem/gen/go/cyber/shared/common/v1"
	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/shared"
)

// DO ------------------------------------------------------------------------------------------------------------------

type Policy struct {
	ID         string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	TenantID   string
	Kind       string
	Name       string
	Params     map[string]any
	Enabled    bool
	BoundCount int64
}

type PolicyListIn struct {
	*commonpb.PageRequest
	OrderBy []string
	Kind    *string
	Name    *string
	Enabled *bool
}

type PolicyListOut struct {
	*commonpb.PageResponse
	List []*Policy
}

// Port ----------------------------------------------------------------------------------------------------------------

type PolicyRP interface {
	Create(ctx context.Context, p *Policy) (*Policy, error)
	Update(ctx context.Context, fieldsMask []string, p *Policy) (*Policy, error)
	Delete(ctx context.Context, id string) (string, error)
	List(ctx context.Context, in *PolicyListIn) (*PolicyListOut, error)
	FindByID(ctx context.Context, id string) (*Policy, error)
	FindByIDs(ctx context.Context, ids []string) ([]*Policy, error)
	NotifyChanged(ctx context.Context)
}

// UC ------------------------------------------------------------------------------------------------------------------

type PolicyUC struct {
	shared.UC
	policyRP PolicyRP
}

func NewPolicyUC(logger *slog.Logger, tm shared.Transaction, policyRP PolicyRP) *PolicyUC {
	return &PolicyUC{
		UC:       shared.NewUC(logger.With("module", "module/policy"), tm),
		policyRP: policyRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *PolicyUC) Create(ctx context.Context, p *Policy) (*Policy, error) {
	if err := validateParams(p.Kind, p.Params); err != nil {
		return nil, err
	}
	out, err := uc.policyRP.Create(ctx, p)
	if err == nil {
		uc.policyRP.NotifyChanged(ctx)
	}
	return out, err
}

func (uc *PolicyUC) Update(ctx context.Context, fieldsMask []string, p *Policy) (*Policy, error) {
	stored, err := uc.policyRP.FindByID(ctx, p.ID)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, systempb.ErrorSystemPolicyNotFound("")
		}
		return nil, err
	}
	setParams := slices.Contains(fieldsMask, "params")
	if setParams {
		// The kind derives from the oneof case; a different case means a
		// different policy type — reconfigure in place is refused, create a
		// new row instead.
		if p.Kind != stored.Kind {
			return nil, systempb.ErrorSystemPolicyInvalidParams("").WithCause(
				fmt.Errorf("policy kind is immutable: stored %q, request %q", stored.Kind, p.Kind))
		}
		if err := validateParams(p.Kind, p.Params); err != nil {
			return nil, err
		}
	}
	out, err := uc.policyRP.Update(ctx, fieldsMask, p)
	if err != nil {
		return nil, err
	}
	// Params and enabled flip constraint verdicts on the next compile; a
	// name-only edit never reaches the snapshot, skip the rebuild broadcast.
	if setParams || slices.Contains(fieldsMask, "enabled") {
		uc.policyRP.NotifyChanged(ctx)
	}
	return out, nil
}

func (uc *PolicyUC) Delete(ctx context.Context, id string) (string, error) {
	var out string
	err := uc.Tm.InTx(ctx, func(ctx context.Context) error {
		// The join rows must go on the same transaction client as the row
		// itself; both statements live in the RP for that reason.
		deleted, derr := uc.policyRP.Delete(ctx, id)
		out = deleted
		return derr
	})
	if err != nil {
		return "", err
	}
	// Deleting unlinks the policy from its grants — a relax-direction change
	// replicas must see even though no grant row was touched.
	uc.policyRP.NotifyChanged(ctx)
	return out, nil
}

func (uc *PolicyUC) List(ctx context.Context, in *PolicyListIn) (*PolicyListOut, error) {
	return uc.policyRP.List(ctx, in)
}

func (uc *PolicyUC) Get(ctx context.Context, id string) (*Policy, error) {
	out, err := uc.policyRP.FindByID(ctx, id)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, systempb.ErrorSystemPolicyNotFound("")
		}
		return nil, err
	}
	return out, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func validateParams(kind string, params map[string]any) error {
	switch kind {
	case kauthz.PolicyKindTimeWindow:
		return validateTimeWindow(params)
	case kauthz.PolicyKindCalendar:
		return validateCalendar(params)
	default:
		// Unreachable via the API (the oneof cannot express an unregistered
		// kind); the guard is for writes that bypass this UC.
		return systempb.ErrorSystemPolicyUnknownKind("").WithCause(
			fmt.Errorf("kind %q is not a known policy type", kind))
	}
}

var timeWindowHHMM = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)

func validateTimeWindow(params map[string]any) error {
	windows, _ := params["windows"].([]any)
	if len(windows) == 0 {
		return systempb.ErrorSystemPolicyInvalidParams("").WithCause(
			fmt.Errorf("windows: at least one window is required"))
	}
	for i, w := range windows {
		span, ok := w.(map[string]any)
		if !ok {
			return spanError(i, "window must be an object")
		}
		start, _ := span["start"].(string)
		end, _ := span["end"].(string)
		if start == "" || end == "" {
			return spanError(i, "start and end are required")
		}
		if !timeWindowHHMM.MatchString(start) {
			return spanError(i, fmt.Sprintf("invalid start %q: want HH:MM", start))
		}
		if !timeWindowHHMM.MatchString(end) && end != "24:00" {
			return spanError(i, fmt.Sprintf("invalid end %q: want HH:MM or 24:00", end))
		}
		// Canonical zero-padded HH:MM compares chronologically as strings and
		// "24:00" sorts after every valid start, so this is the whole
		// ordering check: spans are same-day, cross-midnight is two windows.
		if start >= end {
			return spanError(i, fmt.Sprintf("window %q-%q must satisfy start < end", start, end))
		}
	}
	return nil
}

func validateCalendar(params map[string]any) error {
	base, _ := params["base"].(string)
	switch base {
	case kauthz.CalendarBaseAll, kauthz.CalendarBaseWeekdays, kauthz.CalendarBaseNone:
	default:
		return systempb.ErrorSystemPolicyInvalidParams("").WithCause(
			fmt.Errorf("base: unknown value %q", base))
	}
	overrides, _ := params["overrides"].([]any)
	seen := make(map[string]struct{}, len(overrides))
	for i, o := range overrides {
		entry, ok := o.(map[string]any)
		if !ok {
			return overrideError(i, "override must be an object")
		}
		date, _ := entry["date"].(string)
		// Real calendar validity, not just the proto shape pattern.
		if _, err := time.Parse("2006-01-02", date); err != nil {
			return overrideError(i, fmt.Sprintf("invalid date %q: want YYYY-MM-DD", date))
		}
		if _, dup := seen[date]; dup {
			return overrideError(i, fmt.Sprintf("duplicate date %q", date))
		}
		seen[date] = struct{}{}
	}
	return nil
}

func spanError(i int, why string) error {
	return systempb.ErrorSystemPolicyInvalidParams("").WithCause(
		fmt.Errorf("windows[%d]: %s", i, why))
}

func overrideError(i int, why string) error {
	return systempb.ErrorSystemPolicyInvalidParams("").WithCause(
		fmt.Errorf("overrides[%d]: %s", i, why))
}
