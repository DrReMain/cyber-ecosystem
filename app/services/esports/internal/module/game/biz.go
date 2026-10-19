package game

import (
	"context"
	"log/slog"
	"time"

	"cyber-ecosystem/app/services/esports/internal/shared"
)

// DO ------------------------------------------------------------------------------------------------------------------

type Game struct {
	ID          string
	Code        string
	Name        string
	Category    *string
	IconURL     *string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// RP port ------------------------------------------------------------------------------------------------------------

type GameRP interface {
	List(ctx context.Context, category *string) ([]*Game, error)
	ExistingCodes(ctx context.Context, codes []string) (map[string]struct{}, error)
}

// UC ------------------------------------------------------------------------------------------------------------------

type GameUC struct {
	shared.UC
	gameRP GameRP
}

func NewGameUC(logger *slog.Logger, tm shared.Transaction, gameRP GameRP) *GameUC {
	return &GameUC{UC: shared.NewUC(logger.With("module", "module/game"), tm), gameRP: gameRP}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *GameUC) List(ctx context.Context, category *string) ([]*Game, error) {
	return uc.gameRP.List(ctx, category)
}
