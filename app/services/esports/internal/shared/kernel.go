package shared

import (
	"context"
	"log/slog"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"

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

	GenreFPS  = "fps" // the two genres the product ships — the match engine's draw, the digest's split
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

var DefaultReportPolicy = map[string]string{ // the shipped defaults; report_policy rows are the only override — career open, cognitive the member plane, school codes-only
	ReportCareer:    PolicyOpen,
	ReportCognitive: PolicyMembership,
	ReportSchool:    PolicyClosed,
}

func ReportAccessVerdict(family string, codeGrantActive, policyRowPresent bool, policy string, memberActive bool) (bool, esportspb.ReportAccessReason) {
	// The order's one home: a redeemed code outranks the policy, and the
	// shipped default answers only when no row overrode it. The entitle
	// face and the report cluster faces both quote this, so the facts may
	// drift per caller but the order cannot; the order takes all three
	// facts up front — the lazy membership read the entitle face once
	// saved is a query, not semantics.
	if codeGrantActive {
		return true, esportspb.ReportAccessReason_REPORT_ACCESS_REASON_CODE_REDEEMED
	}
	if !policyRowPresent {
		policy = DefaultReportPolicy[family]
	}
	explicitClosed := policyRowPresent && policy == PolicyClosed
	switch policy {
	case PolicyOpen:
		return true, esportspb.ReportAccessReason_REPORT_ACCESS_REASON_POLICY_OPEN
	case PolicyMembership:
		if memberActive {
			return true, esportspb.ReportAccessReason_REPORT_ACCESS_REASON_MEMBERSHIP_ACTIVE
		}
	case PolicyTenant, PolicyClosed:
	}
	if explicitClosed {
		return false, esportspb.ReportAccessReason_REPORT_ACCESS_REASON_POLICY_CLOSED
	}
	return false, esportspb.ReportAccessReason_REPORT_ACCESS_REASON_NO_ENTITLEMENT
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
