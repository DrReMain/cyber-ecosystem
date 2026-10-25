package test

import (
	"context"
	"log/slog"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"

	"cyber-ecosystem/app/services/esports/internal/ent"
	"cyber-ecosystem/app/services/esports/internal/ent/questionbank"
	"cyber-ecosystem/app/services/esports/internal/ent/test"
	"cyber-ecosystem/app/services/esports/internal/ent/testconfig"
	"cyber-ecosystem/app/services/esports/internal/ent/testrecord"
	"cyber-ecosystem/app/services/esports/internal/platform"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

// Repo ----------------------------------------------------------------------------------------------------------------

type testRP struct {
	shared.RP
}

func NewTestRP(logger *slog.Logger, p *platform.Platform) TestRP {
	return &testRP{RP: shared.NewRP(logger.With("module", "module/test_rp"), p)}
}

// Method --------------------------------------------------------------------------------------------------------------

func (rp *testRP) List(ctx context.Context) ([]*Test, error) {
	rows, err := rp.Platform.GetClient(ctx).Test.Query().
		Order(ent.Asc(test.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	tests := make([]*Test, 0, len(rows))
	for _, row := range rows {
		tests = append(tests, mapTest(row))
	}
	return tests, nil
}

func (rp *testRP) GetByID(ctx context.Context, id string) (*Test, error) {
	row, err := rp.Platform.GetClient(ctx).Test.Get(ctx, id)
	if err != nil {
		return nil, mapTestLoadError(rp, err)
	}
	return mapTest(row), nil
}

func (rp *testRP) GetConfigByID(ctx context.Context, id string) (*TestConfig, error) {
	// The config row joins by the catalog code, the contract's join key;
	// the request carries the test id, so the owning row resolves first.
	owned, err := rp.Platform.GetClient(ctx).Test.Get(ctx, id)
	if err != nil {
		return nil, mapTestLoadError(rp, err)
	}
	row, err := rp.Platform.GetClient(ctx).TestConfig.Query().
		Where(testconfig.TestCodeEQ(owned.Code)).
		Only(ctx)
	if err != nil {
		return nil, mapTestLoadError(rp, err)
	}
	tc := mapTestConfig(row)
	tc.TestID = owned.ID
	return tc, nil
}

func (rp *testRP) GetByCode(ctx context.Context, code string) (*Test, error) {
	row, err := rp.Platform.GetClient(ctx).Test.Query().
		Where(test.CodeEQ(code)).
		Only(ctx)
	if err != nil {
		return nil, mapTestLoadError(rp, err)
	}
	return mapTest(row), nil
}

func (rp *testRP) GetConfigByCode(ctx context.Context, code string) (*TestConfig, error) {
	row, err := rp.Platform.GetClient(ctx).TestConfig.Query().
		Where(testconfig.TestCodeEQ(code)).
		Only(ctx)
	if err != nil {
		return nil, mapTestLoadError(rp, err)
	}
	return mapTestConfig(row), nil
}

func (rp *testRP) ListAnswers(ctx context.Context, testCode string) (map[string]map[string]any, error) {
	// The soft-delete interceptor keeps tombstoned scenarios out; the bank's
	// unique (test_code, scenario_code) makes the projection a clean map.
	rows, err := rp.Platform.GetClient(ctx).QuestionBank.Query().
		Where(questionbank.TestCodeEQ(testCode)).
		All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	key := make(map[string]map[string]any, len(rows))
	for _, row := range rows {
		key[row.ScenarioCode] = row.Answer
	}
	return key, nil
}

func (rp *testRP) CountBySessionCodes(ctx context.Context, sessionID string) (map[string]int, error) {
	var rows []struct {
		TestCode string `json:"test_code"`
		Count    int    `json:"count"`
	}
	err := rp.Platform.GetClient(ctx).TestRecord.Query().
		Where(testrecord.SessionIDEQ(sessionID)).
		GroupBy(testrecord.FieldTestCode).
		Aggregate(ent.Count()).
		Scan(ctx, &rows)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	counts := make(map[string]int, len(rows))
	for _, row := range rows {
		counts[row.TestCode] = row.Count
	}
	return counts, nil
}

func (rp *testRP) GetAttempt(ctx context.Context, id string) (*AttemptRecord, error) {
	row, err := rp.Platform.GetClient(ctx).TestRecord.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, rp.Platform.HandleEntError(err)
	}
	grade := (*string)(nil)
	if row.Grade != nil {
		g := string(*row.Grade)
		grade = &g
	}
	return &AttemptRecord{
		ID:            row.ID,
		TestCode:      row.TestCode,
		UserID:        row.UserID,
		AttemptNo:     row.AttemptNo,
		SessionID:     row.SessionID,
		StartedAt:     row.StartedAt,
		EndedAt:       row.EndedAt,
		Backgrounds:   row.Backgrounds,
		RawDetail:     row.RawDetail,
		RawScore:      row.RawScore,
		Authoritative: row.AuthoritativeScore,
		Grade:         grade,
		Valid:         row.Valid,
		InvalidReason: row.InvalidReason,
	}, nil
}

func (rp *testRP) CreateAttempt(ctx context.Context, rec *AttemptRecord) (*AttemptRecord, error) {
	row, err := rp.Platform.GetClient(ctx).TestRecord.Create().
		SetTestCode(rec.TestCode).
		SetUserID(rec.UserID).
		SetAttemptNo(rec.AttemptNo).
		SetNillableSessionID(rec.SessionID).
		SetStartedAt(rec.StartedAt).
		SetEndedAt(rec.EndedAt).
		SetBackgrounds(rec.Backgrounds).
		SetRawDetail(rec.RawDetail).
		SetNillableRawScore(rec.RawScore).
		SetNillableAuthoritativeScore(rec.Authoritative).
		SetNillableGrade((*testrecord.Grade)(rec.Grade)).
		SetValid(rec.Valid).
		SetNillableInvalidReason(rec.InvalidReason).
		Save(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	rec.ID = row.ID
	return rec, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mapTestLoadError(rp *testRP, err error) error {
	// A missing row is the domain's not-found, not the generic infra word —
	// both read paths resolve through the catalog first.
	if ent.IsNotFound(err) {
		return esportspb.ErrorEsportsTestNotFound("")
	}
	return rp.Platform.HandleEntError(err)
}

func mapTestConfig(row *ent.TestConfig) *TestConfig {
	return &TestConfig{
		DurationSec:          row.DurationSeconds,
		BackgroundInvalidSec: row.BackgroundInvalidSeconds,
		EngineParams:         row.EngineParams,
	}
}

func mapTest(row *ent.Test) *Test {
	return &Test{
		ID:             row.ID,
		Code:           row.Code,
		Name:           row.Name,
		Category:       row.Category,
		Ability:        row.Ability,
		Description:    row.Description,
		UnitScoreLabel: row.UnitScoreLabel,
		DurationSec:    row.DurationSeconds,
		Introduction:   row.Introduction,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
}
