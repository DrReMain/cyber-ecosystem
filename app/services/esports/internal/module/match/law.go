package match

import (
	"context"
	"log/slog"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"

	"cyber-ecosystem/app/services/esports/internal/module/test"
)

type sessionLaw struct {
	log     *slog.Logger
	matchRP MatchRP
}

func NewSessionLaw(logger *slog.Logger, matchRP MatchRP) test.MatchSessionLaw {
	return &sessionLaw{log: logger.With("module", "module/match"), matchRP: matchRP}
}

func (l *sessionLaw) Resolve(ctx context.Context, userID, sessionID string) (*test.SessionBinding, error) {
	s, err := l.matchRP.FindByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	// Ownership folds into the miss: a foreign or absent session reads the
	// same not-found, so the miss leaks nothing. An active session binds
	// at any age — the lineup is locked in the row and resumes forever.
	if s == nil || s.UserID != userID || s.Status != statusActive {
		return nil, esportspb.ErrorEsportsMatchSessionNotFound("")
	}
	return &test.SessionBinding{
		Difficulty:   s.Difficulty,
		RetestBudget: s.RetestBudget,
		Steps:        s.Steps,
	}, nil
}
