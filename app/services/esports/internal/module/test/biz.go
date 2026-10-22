package test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/protobuf/encoding/protojson"

	"cyber-ecosystem/shared-go/kratos/security"
	"cyber-ecosystem/shared-go/utils"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"

	"cyber-ecosystem/app/services/esports/internal/ent/schema/types"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

const firstSessionRetestBudget = 3

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

// RP port ------------------------------------------------------------------------------------------------------------

type TestRP interface {
	List(ctx context.Context) ([]*Test, error)
	GetByID(ctx context.Context, id string) (*Test, error)
	GetByCode(ctx context.Context, code string) (*Test, error)
	GetConfigByID(ctx context.Context, id string) (*TestConfig, error)
	GetConfigByCode(ctx context.Context, code string) (*TestConfig, error)
	CreateAttempt(ctx context.Context, rec *AttemptRecord) (*AttemptRecord, error)
}

// UC ------------------------------------------------------------------------------------------------------------------

type TestUC struct {
	shared.UC
	testRP TestRP
}

func NewTestUC(logger *slog.Logger, tm shared.Transaction, testRP TestRP) *TestUC {
	return &TestUC{UC: shared.NewUC(logger.With("module", "module/test"), tm), testRP: testRP}
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
	if in.GetAttemptNo() < 1 {
		// The ordinal rides without required presence on the wire; its law
		// (E4) starts at 1, and the storage schema rejects zero besides.
		return nil, errorspb.ErrorGeneralErrorValidationFailed("").WithCause(fmt.Errorf("attempt_no must be >= 1"))
	}
	t, err := uc.testRP.GetByCode(ctx, in.GetTestCode())
	if err != nil {
		return nil, err
	}
	cfg, err := uc.testRP.GetConfigByCode(ctx, in.GetTestCode())
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
		TestCode:    t.Code,
		UserID:      subject.UserID,
		AttemptNo:   int(in.GetAttemptNo()),
		StartedAt:   *utils.FromTimestamp(in.GetStartedAt()),
		EndedAt:     *utils.FromTimestamp(in.GetEndedAt()),
		Backgrounds: backgrounds,
		Valid:       true,
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
		// The mode axis (experience 1.0 / standard 1.5, metric.md §三) is
		// match-session law (E4); until the session aggregate lands every
		// attempt scores on the experience baseline.
		res, serr := sc.recompute(in, scoringParams{engineParams: cfg.EngineParams})
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
	rec, err = uc.testRP.CreateAttempt(ctx, rec)
	if err != nil {
		return nil, err
	}
	retestLeft := utils.ConvNum[int32](firstSessionRetestBudget - int(in.GetAttemptNo()) + 1)
	if retestLeft < 0 {
		retestLeft = 0
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
