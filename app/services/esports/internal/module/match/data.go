package match

import (
	"context"
	"log/slog"

	"cyber-ecosystem/app/services/esports/internal/ent"
	"cyber-ecosystem/app/services/esports/internal/ent/matchconfig"
	"cyber-ecosystem/app/services/esports/internal/ent/matchsession"
	"cyber-ecosystem/app/services/esports/internal/platform"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

// Repo ----------------------------------------------------------------------------------------------------------------

type matchRP struct {
	shared.RP
}

func NewMatchRP(logger *slog.Logger, p *platform.Platform) MatchRP {
	return &matchRP{RP: shared.NewRP(logger.With("module", "module/match_rp"), p)}
}

// Method --------------------------------------------------------------------------------------------------------------

func (rp *matchRP) Create(ctx context.Context, s *MatchSession) (*MatchSession, error) {
	row, err := rp.Platform.GetClient(ctx).MatchSession.Create().
		SetUserID(s.UserID).
		SetGenre(s.Genre).
		SetDifficulty(matchsession.Difficulty(s.Difficulty)).
		SetSeed(s.Seed).
		SetSteps(s.Steps).
		SetRetestBudget(s.RetestBudget).
		SetStatus(matchsession.Status(s.Status)).
		Save(ctx)
	if err != nil {
		// The partial unique (user_id where active) turns a concurrent
		// create-or-resume race into a constraint error here — the caller
		// re-reads, the surviving session wins.
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapMatchSession(row), nil
}

func (rp *matchRP) UpdateStatus(ctx context.Context, id, status string) error {
	if err := rp.Platform.GetClient(ctx).MatchSession.UpdateOneID(id).
		SetStatus(matchsession.Status(status)).
		Exec(ctx); err != nil {
		return rp.Platform.HandleEntError(err)
	}
	return nil
}

func (rp *matchRP) ListConfig(ctx context.Context, genre string) ([]MatchConfig, error) {
	rows, err := rp.Platform.GetClient(ctx).MatchConfig.Query().
		Where(matchconfig.GenreEQ(genre)).
		Order(ent.Asc(matchconfig.FieldRenderOrder), ent.Asc(matchconfig.FieldAbility)).
		All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	configs := make([]MatchConfig, 0, len(rows))
	for _, row := range rows {
		configs = append(configs, *mapMatchConfig(row))
	}
	return configs, nil
}

func (rp *matchRP) FindActive(ctx context.Context, userID string) (*MatchSession, error) {
	row, err := rp.Platform.GetClient(ctx).MatchSession.Query().
		Where(matchsession.UserIDEQ(userID), matchsession.StatusEQ(statusActive)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapMatchSession(row), nil
}

func (rp *matchRP) FindByID(ctx context.Context, id string) (*MatchSession, error) {
	row, err := rp.Platform.GetClient(ctx).MatchSession.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapMatchSession(row), nil
}

func (rp *matchRP) CountByUser(ctx context.Context, userID string) (int, error) {
	count, err := rp.Platform.GetClient(ctx).MatchSession.Query().
		Where(matchsession.UserIDEQ(userID)).
		Count(ctx)
	if err != nil {
		return 0, rp.Platform.HandleEntError(err)
	}
	return count, nil
}

func (rp *matchRP) CountByUserGenre(ctx context.Context, userID, genre string) (int, error) {
	count, err := rp.Platform.GetClient(ctx).MatchSession.Query().
		Where(matchsession.UserIDEQ(userID), matchsession.GenreEQ(genre)).
		Count(ctx)
	if err != nil {
		return 0, rp.Platform.HandleEntError(err)
	}
	return count, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mapMatchSession(row *ent.MatchSession) *MatchSession {
	return &MatchSession{
		ID:           row.ID,
		UserID:       row.UserID,
		Genre:        row.Genre,
		Difficulty:   string(row.Difficulty),
		Seed:         row.Seed,
		Steps:        row.Steps,
		RetestBudget: row.RetestBudget,
		Status:       string(row.Status),
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

func mapMatchConfig(row *ent.MatchConfig) *MatchConfig {
	return &MatchConfig{
		ID:          row.ID,
		Genre:       row.Genre,
		Ability:     row.Ability,
		RenderOrder: row.RenderOrder,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}
