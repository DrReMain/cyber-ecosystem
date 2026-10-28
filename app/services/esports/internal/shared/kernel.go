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

	PolicyOpen       = "open" // report_policy's admission modes — entitle judges by them, entitleadmin sets them, same two-modules-one-home law as the family above
	PolicyMembership = "membership"
	PolicyTenant     = "tenant"
	PolicyClosed     = "closed"

	SexMale         = "male" // the trio profile collects and the report cohort keys on — undetermined is the honest not-said, never guessed from anything
	SexFemale       = "female"
	SexUndetermined = "undetermined"

	TierS = "S" // the letters the judgement, the grade column, and the unlock-time snapshot all carry
	TierA = "A"
	TierB = "B"
	TierC = "C"

	TierUnjudged = "unjudged" // the snapshot's honest no-verdict: every cohort too thin for a percentile at pin time — a letter would be a made-up number

	GenreFPS  = "fps"  // the two genres the product ships — the match engine's draw, the digest's split
	GenreMOBA = "moba"
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

type ReportUnlocked struct {
	UserID   string
	RecordID string
}

type DomainEventBus interface {
	PublishTrialCompleted(ctx context.Context, e TrialCompleted) // in-process, synchronous, best-effort: the writer publishes a fact, the owner subscribes, a handler failure stays inside the bus — the publisher's transaction is truth
	SubscribeTrialCompleted(handler func(context.Context, TrialCompleted))
	PublishReportUnlocked(ctx context.Context, e ReportUnlocked) // a trial grant names its record, so the report owner can freeze the judgement at unlock time; membership and family grants name no record — those freeze lazily on first read
	SubscribeReportUnlocked(handler func(context.Context, ReportUnlocked))
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
