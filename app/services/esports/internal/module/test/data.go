package test

import (
	"context"
	"log/slog"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"

	"cyber-ecosystem/app/services/esports/internal/ent"
	"cyber-ecosystem/app/services/esports/internal/ent/test"
	"cyber-ecosystem/app/services/esports/internal/ent/testconfig"
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
	return &TestConfig{
		TestID:               owned.ID,
		DurationSec:          row.DurationSeconds,
		BackgroundInvalidSec: row.BackgroundInvalidSeconds,
		EngineParams:         row.EngineParams,
	}, nil
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
