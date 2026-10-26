package shared

import (
	"context"
	"log/slog"

	"cyber-ecosystem/app/services/esports/internal/platform"
)

const ExtraAccount = "account"

const (
	ScopeTrial      = "trial" // the ledger's scope vocabulary keeps one home: pay and entitle both write these rows, and a drifted literal would silently split the ledger
	ScopeMembership = "membership"
	ScopeReportType = "report_type"
)

const (
	ReportCognitive = "cognitive" // the report-family and policy vocabulary: entitle judges by it, entitleadmin sets it — two modules, one home, or operator config drifts from judgement
	ReportCareer    = "career"
	ReportSchool    = "school"

	PolicyOpen       = "open"
	PolicyMembership = "membership"
	PolicyTenant     = "tenant"
	PolicyClosed     = "closed"

	SexMale         = "male"
	SexFemale       = "female"
	SexUndetermined = "undetermined"

	TierS = "S"
	TierA = "A"
	TierB = "B"
	TierC = "C"
)

type Transaction interface {
	InTx(ctx context.Context, fn func(context.Context) error) error
}

type HookRegistry interface {
	OnStart(fn func(context.Context) error)
	OnStop(fn func(context.Context) error)
}

type TrialCompleted struct {
	UserID   string
	RecordID string
}

type DomainEventBus interface {
	PublishTrialCompleted(ctx context.Context, e TrialCompleted) // in-process, synchronous, best-effort: the writer publishes a fact, the owner subscribes, a handler failure stays inside the bus — the publisher's transaction is truth
	SubscribeTrialCompleted(handler func(context.Context, TrialCompleted))
}

// Base ----------------------------------------------------------------------------------------------------------------

type UC struct {
	Log *slog.Logger
	Tm  Transaction
}

func NewUC(log *slog.Logger, tm Transaction) UC {
	return UC{Log: log, Tm: tm}
}

type RP struct {
	Log      *slog.Logger
	Platform *platform.Platform
}

func NewRP(log *slog.Logger, platform *platform.Platform) RP {
	return RP{Log: log, Platform: platform}
}
