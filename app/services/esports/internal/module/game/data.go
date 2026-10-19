package game

import (
	"context"
	"log/slog"

	"cyber-ecosystem/app/services/esports/internal/ent"
	"cyber-ecosystem/app/services/esports/internal/ent/game"
	"cyber-ecosystem/app/services/esports/internal/platform"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

// Repo ----------------------------------------------------------------------------------------------------------------

type gameRP struct {
	shared.RP
}

func NewGameRP(logger *slog.Logger, p *platform.Platform) GameRP {
	return &gameRP{RP: shared.NewRP(logger.With("module", "module/game_rp"), p)}
}

// Method --------------------------------------------------------------------------------------------------------------

func (rp *gameRP) List(ctx context.Context, category *string) ([]*Game, error) {
	query := rp.Platform.GetClient(ctx).Game.Query().
		Order(ent.Asc(game.FieldCreatedAt))
	if category != nil {
		query = query.Where(game.CategoryEQ(*category))
	}
	rows, err := query.All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	games := make([]*Game, 0, len(rows))
	for _, row := range rows {
		games = append(games, mapGame(row))
	}
	return games, nil
}

func (rp *gameRP) ExistingCodes(ctx context.Context, codes []string) (map[string]struct{}, error) {
	if len(codes) == 0 {
		return map[string]struct{}{}, nil
	}
	rows, err := rp.Platform.GetClient(ctx).Game.Query().
		Where(game.CodeIn(codes...)).
		All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	set := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		set[row.Code] = struct{}{}
	}
	return set, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mapGame(row *ent.Game) *Game {
	return &Game{
		ID:          row.ID,
		Code:        row.Code,
		Name:        row.Name,
		Category:    row.Category,
		IconURL:     row.IconURL,
		Description: row.Description,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}
