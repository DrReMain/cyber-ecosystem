package test

import (
	"context"
	"log/slog"
	"time"

	"cyber-ecosystem/app/services/esports/internal/shared"
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

// RP port ------------------------------------------------------------------------------------------------------------

type TestRP interface {
	List(ctx context.Context) ([]*Test, error)
	GetByID(ctx context.Context, id string) (*Test, error)
	GetConfigByID(ctx context.Context, id string) (*TestConfig, error)
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
