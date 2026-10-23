package test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	kerrors "github.com/go-kratos/kratos/v3/errors"
	"google.golang.org/protobuf/encoding/protojson"

	"cyber-ecosystem/shared-go/kratos/security"
	"cyber-ecosystem/shared-go/utils"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"

	"cyber-ecosystem/app/services/esports/internal/ent/schema/types"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

const (
	difficultyExperienceFactor = 1.0
	difficultyStandardFactor   = 1.5

	// The mode word the session aggregate freezes at birth; the port
	// carries it as a plain string, so this consumer spells it once more —
	// importing the owning module would close a cycle.
	difficultyStandard = "standard"
)

// DO ------------------------------------------------------------------------------------------------------------------

type Test struct {
	ID             string
	Code           string
	Name           string
	Category       *string
	Ability        *string
	Description    *string
	UnitScoreLabel *string
	DurationSec    int
	Introduction   *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type TestConfig struct {
	TestID               string
	DurationSec          int
	BackgroundInvalidSec int
	EngineParams         map[string]string
}

type AttemptRecord struct {
	ID            string
	TestCode      string
	UserID        string
	AttemptNo     int
	SessionID     *string
	StartedAt     time.Time
	EndedAt       time.Time
	Backgrounds   []types.BackgroundSpan
	RawDetail     map[string]any
	RawScore      *float64
	Authoritative *float64
	Grade         *string
	Valid         bool
	InvalidReason *string
}

type SubmitOut struct {
	RecordID      string
	RawScore      *float64
	Authoritative *float64
	Grade         *string
	Ability       *string
	Valid         bool
	InvalidReason *string
	RetestLeft    int32
}

type SessionBinding struct {
	Difficulty   string
	RetestBudget int
	Steps        []types.MatchStep
}

// RP port ------------------------------------------------------------------------------------------------------------

type TestRP interface {
	List(ctx context.Context) ([]*Test, error)
	GetByID(ctx context.Context, id string) (*Test, error)
	GetByCode(ctx context.Context, code string) (*Test, error)
	GetConfigByID(ctx context.Context, id string) (*TestConfig, error)
	GetConfigByCode(ctx context.Context, code string) (*TestConfig, error)
	ListAnswers(ctx context.Context, testCode string) (map[string]map[string]any, error)
	CountBySessionCodes(ctx context.Context, sessionID string) (map[string]int, error)
	CreateAttempt(ctx context.Context, rec *AttemptRecord) (*AttemptRecord, error)
}

type MatchSessionLaw interface {
	// Resolve validates the session for the submitting member — an active,
	// in-window, member-owned session yields its binding; anything else is
	// a miss. Implemented by the match module over its own aggregate.
	Resolve(ctx context.Context, userID, sessionID string) (*SessionBinding, error)
}

// UC ------------------------------------------------------------------------------------------------------------------

type TestUC struct {
	shared.UC
	testRP   TestRP
	matchLaw MatchSessionLaw
}

func NewTestUC(logger *slog.Logger, tm shared.Transaction, testRP TestRP, matchLaw MatchSessionLaw) *TestUC {
	return &TestUC{UC: shared.NewUC(logger.With("module", "module/test"), tm), testRP: testRP, matchLaw: matchLaw}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *TestUC) List(ctx context.Context) ([]*Test, error) {
	return uc.testRP.List(ctx)
}

func (uc *TestUC) Get(ctx context.Context, id string) (*Test, error) {
	return uc.testRP.GetByID(ctx, id)
}

func (uc *TestUC) GetConfig(ctx context.Context, id string) (*TestConfig, error) {
	return uc.testRP.GetConfigByID(ctx, id)
}

func (uc *TestUC) Submit(ctx context.Context, in *esportspb.SubmitTestAttemptRequest) (*SubmitOut, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("test submit requires a session"))
	}
	t, err := uc.testRP.GetByCode(ctx, in.GetTestCode())
	if err != nil {
		return nil, err
	}
	cfg, err := uc.testRP.GetConfigByCode(ctx, in.GetTestCode())
	if err != nil {
		return nil, err
	}
	// The bank's answer key rides the params — question-driven correctness
	// is judged server-side against it (question_bank.answer law); the read
	// is empty for fact-stream games with no bank.
	answers, err := uc.testRP.ListAnswers(ctx, t.Code)
	if err != nil {
		return nil, err
	}
	sc, ok := lookupScorer(t.Code)
	if !ok {
		return nil, esportspb.ErrorEsportsUnspecified("").WithCause(fmt.Errorf("no scorer registered for %s", t.Code))
	}
	backgrounds := make([]types.BackgroundSpan, 0, len(in.GetBackgrounds()))
	for _, b := range in.GetBackgrounds() {
		backgrounds = append(backgrounds, types.BackgroundSpan{
			ElapsedMS:  int(b.GetElapsedMs()),
			DurationMS: int(b.GetDurationMs()),
		})
	}
	rec := &AttemptRecord{
		TestCode: t.Code,
		UserID:   subject.UserID,
		// The practice default: no session, so the ordinal is vacuous and
		// the session branch below overwrites it with the real count.
		AttemptNo:   1,
		StartedAt:   *utils.FromTimestamp(in.GetStartedAt()),
		EndedAt:     *utils.FromTimestamp(in.GetEndedAt()),
		Backgrounds: backgrounds,
		Valid:       true,
	}
	// Practice rides outside the session budget — the receipt's budget face
	// reports zero rather than a budget that does not exist.
	factor := difficultyExperienceFactor
	retestLeft := int32(0)
	var binding *SessionBinding
	sid := in.GetMatchSessionId()
	if sid != "" {
		binding, err = uc.matchLaw.Resolve(ctx, subject.UserID, sid)
		if err != nil {
			return nil, err
		}
		inSteps := false
		for _, st := range binding.Steps {
			if st.TestCode == t.Code {
				inSteps = true
				break
			}
		}
		if !inSteps {
			// The lineup is server-drawn; a submit outside it is a client
			// selecting its own game, which the flow never allows.
			return nil, esportspb.ErrorEsportsMatchTestNotInSession("").WithCause(fmt.Errorf("test %s outside session %s steps", t.Code, sid))
		}
		counts, err := uc.testRP.CountBySessionCodes(ctx, sid)
		if err != nil {
			return nil, err
		}
		ordinal := counts[t.Code] + 1
		if ordinal > binding.RetestBudget {
			// Every stored record spent a chance, valid or not — used
			// budget never comes back inside a session.
			return nil, esportspb.ErrorEsportsMatchRetestExhausted("").WithCause(fmt.Errorf("test %s at attempt %d of budget %d in session %s", t.Code, ordinal, binding.RetestBudget, sid))
		}
		rec.SessionID = in.MatchSessionId
		rec.AttemptNo = ordinal
		factor = difficultyFactorOf(binding.Difficulty)
	}
	if rec.RawDetail, err = rawDetailMap(in); err != nil {
		// The helper's prose stays in the cause channel; identity is a
		// catalog factory — a bare fmt.Errorf must never escape as one.
		return nil, errorspb.ErrorGeneralErrorInternal("").WithCause(err)
	}
	// Per-span reading: the threshold names one backgrounding event ("longer
	// than this"), not the run's total pause — three short peeks stay play.
	for _, b := range backgrounds {
		if b.DurationMS > cfg.BackgroundInvalidSec*1000 {
			rec.Valid = false
			rec.InvalidReason = utils.Ptr(reasonBackgroundOverBudget)
			break
		}
	}
	if rec.Valid {
		res, serr := sc.recompute(in, scoringParams{
			answerKey:        answers,
			engineParams:     cfg.EngineParams,
			durationSec:      cfg.DurationSec,
			difficultyFactor: factor,
		})
		if serr != nil {
			var invalid *invalidAttemptError
			if !errors.As(serr, &invalid) {
				return nil, serr
			}
			rec.Valid = false
			rec.InvalidReason = utils.Ptr(invalid.reason)
		} else {
			rec.RawScore = utils.Ptr(res.raw)
			rec.Authoritative = res.authoritative
			rec.Grade = res.grade
		}
	}
	rec, err = uc.storeAttempt(ctx, rec, binding)
	if err != nil {
		return nil, err
	}
	if rec.SessionID != nil {
		// The stored ordinal is the claim's truth — a concurrent submit may
		// have moved it past the count this request read.
		retestLeft = utils.ConvNum[int32](binding.RetestBudget - rec.AttemptNo)
	}
	return &SubmitOut{
		RecordID:      rec.ID,
		RawScore:      rec.RawScore,
		Authoritative: rec.Authoritative,
		Grade:         rec.Grade,
		Ability:       t.Ability,
		Valid:         rec.Valid,
		InvalidReason: rec.InvalidReason,
		RetestLeft:    retestLeft,
	}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func (uc *TestUC) storeAttempt(ctx context.Context, rec *AttemptRecord, binding *SessionBinding) (*AttemptRecord, error) {
	// The budget check reads a count a concurrent submit can invalidate
	// before this insert lands, so the partial unique on (session_id,
	// test_code, attempt_no) turns that race into a constraint failure —
	// one recount pass claims the next slot, and a second collision
	// escapes as the infra error it is.
	stored, err := uc.testRP.CreateAttempt(ctx, rec)
	if err == nil || rec.SessionID == nil || !isClaimLost(err) {
		return stored, err
	}
	counts, cerr := uc.testRP.CountBySessionCodes(ctx, *rec.SessionID)
	if cerr != nil {
		return nil, err
	}
	ordinal := counts[rec.TestCode] + 1
	if ordinal > binding.RetestBudget {
		return nil, esportspb.ErrorEsportsMatchRetestExhausted("").WithCause(fmt.Errorf("test %s at attempt %d of budget %d in session %s after a concurrent submit", rec.TestCode, ordinal, binding.RetestBudget, *rec.SessionID))
	}
	rec.AttemptNo = ordinal
	return uc.testRP.CreateAttempt(ctx, rec)
}

func isClaimLost(err error) bool {
	var ke *kerrors.Error
	// This table's only unique constraint beyond the primary key is the
	// session ordinal claim, so a constraint failure means a concurrent
	// submit stored the same ordinal first.
	return errors.As(err, &ke) && ke.Reason == errorspb.InfraError_INFRA_ERROR_DB_CONSTRAINT.String()
}

func difficultyFactorOf(mode string) float64 {
	if mode == difficultyStandard {
		return difficultyStandardFactor
	}
	// Experience is the baseline; unknown words degrade to it rather than
	// failing the submit — the mode vocabulary is born with the session.
	return difficultyExperienceFactor
}

func rawDetailMap(in *esportspb.SubmitTestAttemptRequest) (map[string]any, error) {
	m := in.ProtoReflect()
	od := m.Descriptor().Oneofs().ByName("raw_detail")
	set := m.WhichOneof(od)
	if set == nil {
		return map[string]any{}, nil
	}
	raw, err := protojson.Marshal(m.Get(set).Message().Interface())
	if err != nil {
		return nil, fmt.Errorf("submit: marshal %s detail: %w", set.TextName(), err)
	}
	detail, err := utils.Unmarshal[map[string]any](raw)
	if err != nil {
		return nil, fmt.Errorf("submit: decode %s detail: %w", set.TextName(), err)
	}
	return map[string]any{set.TextName(): detail}, nil
}
