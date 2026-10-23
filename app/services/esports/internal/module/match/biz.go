package match

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	kerrors "github.com/go-kratos/kratos/v3/errors"

	"cyber-ecosystem/shared-go/kratos/security"
	"cyber-ecosystem/shared-go/utils"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"

	"cyber-ecosystem/app/services/esports/internal/ent/schema/types"
	"cyber-ecosystem/app/services/esports/internal/module/game"
	"cyber-ecosystem/app/services/esports/internal/module/profile"
	"cyber-ecosystem/app/services/esports/internal/module/test"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

const (
	statusActive   = "active"
	statusSwitched = "switched"
	statusExpired  = "expired"

	difficultyExperience = "experience"

	matchResumeWindow  = 7 * 24 * time.Hour
	firstSessionBudget = 3
	laterSessionBudget = 1
)

// DO ------------------------------------------------------------------------------------------------------------------

type MatchConfig struct {
	ID          string
	Genre       string
	Ability     string
	RenderOrder int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type MatchSession struct {
	ID           string
	UserID       string
	Genre        string
	Difficulty   string
	Seed         int64
	Steps        []types.MatchStep
	RetestBudget int
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type SessionView struct {
	Session    *MatchSession
	Done       []string
	RetestLeft map[string]int32
	First      bool
	Created    bool
}

// RP port ------------------------------------------------------------------------------------------------------------

type MatchRP interface {
	Create(ctx context.Context, s *MatchSession) (*MatchSession, error)
	UpdateStatus(ctx context.Context, id, status string) error
	ListConfig(ctx context.Context, genre string) ([]MatchConfig, error)
	FindActive(ctx context.Context, userID string) (*MatchSession, error)
	FindByID(ctx context.Context, id string) (*MatchSession, error)
	CountByUser(ctx context.Context, userID string) (int, error)
	CountByUserGenre(ctx context.Context, userID, genre string) (int, error)
}

// UC ------------------------------------------------------------------------------------------------------------------

type MatchUC struct {
	shared.UC
	matchRP   MatchRP
	testRP    test.TestRP
	profileRP profile.ProfileRP
	gameRP    game.GameRP
}

func NewMatchUC(logger *slog.Logger, tm shared.Transaction, matchRP MatchRP, testRP test.TestRP, profileRP profile.ProfileRP, gameRP game.GameRP) *MatchUC {
	return &MatchUC{
		UC:        shared.NewUC(logger.With("module", "module/match"), tm),
		matchRP:   matchRP,
		testRP:    testRP,
		profileRP: profileRP,
		gameRP:    gameRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *MatchUC) Get(ctx context.Context, genre *string) (*SessionView, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("match session requires a session"))
	}
	active, err := uc.matchRP.FindActive(ctx, subject.UserID)
	if err != nil {
		return nil, err
	}
	if active != nil {
		stale, err := expireIfStale(ctx, uc.matchRP, active)
		if err != nil {
			return nil, err
		}
		if !stale {
			return uc.view(ctx, active, false)
		}
	}
	session, err := uc.create(ctx, subject.UserID, genre, 0)
	if err != nil {
		if !isClaimLost(err) {
			return nil, err
		}
		twin, terr := uc.matchRP.FindActive(ctx, subject.UserID)
		if terr != nil || twin == nil {
			return nil, err
		}
		// A twin request won the insert race; for this member the session
		// is freshly born either way.
		return uc.view(ctx, twin, true)
	}
	return uc.view(ctx, session, true)
}

func (uc *MatchUC) Switch(ctx context.Context, genre string) (*SessionView, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("match switch requires a session"))
	}
	// The target validates before anything closes — a switch at an unknown
	// genre must not burn the active session and its remaining budget.
	configs, err := uc.matchRP.ListConfig(ctx, genre)
	if err != nil {
		return nil, err
	}
	if len(configs) == 0 {
		return nil, esportspb.ErrorEsportsMatchGenreUnknown("").WithCause(fmt.Errorf("no match_config rows for genre %q", genre))
	}
	active, err := uc.matchRP.FindActive(ctx, subject.UserID)
	if err != nil {
		return nil, err
	}
	if active != nil {
		// A switch is an explicit abandon — even an untouched session
		// closes; the same test may be drawn again in the replacement.
		if err := uc.matchRP.UpdateStatus(ctx, active.ID, statusSwitched); err != nil {
			return nil, err
		}
	}
	salt, err := uc.matchRP.CountByUserGenre(ctx, subject.UserID, genre)
	if err != nil {
		return nil, err
	}
	session, err := uc.create(ctx, subject.UserID, &genre, salt)
	if err != nil {
		return nil, err
	}
	return uc.view(ctx, session, true)
}

// Private -------------------------------------------------------------------------------------------------------------

func expireIfStale(ctx context.Context, rp MatchRP, s *MatchSession) (bool, error) {
	// An over-window session closes lazily so the next draw is fresh —
	// used retest counts never come back with a new one.
	if time.Since(s.CreatedAt) < matchResumeWindow {
		return false, nil
	}
	if err := rp.UpdateStatus(ctx, s.ID, statusExpired); err != nil {
		return false, err
	}
	return true, nil
}

func isClaimLost(err error) bool {
	var ke *kerrors.Error
	// The session insert's only unique claim is one-active-per-user, so a
	// constraint failure there means a twin create-or-resume won the race.
	return errors.As(err, &ke) && ke.Reason == errorspb.InfraError_INFRA_ERROR_DB_CONSTRAINT.String()
}

func (uc *MatchUC) create(ctx context.Context, userID string, genre *string, salt int) (*MatchSession, error) {
	var resolved string
	if genre != nil {
		resolved = *genre
	} else {
		p, err := uc.profileRP.FindByUserID(ctx, userID)
		if err != nil {
			return nil, err
		}
		games, err := uc.gameRP.List(ctx, nil)
		if err != nil {
			return nil, err
		}
		entries := []types.GameEntry(nil)
		if p != nil {
			entries = p.Games
		}
		voted, ok := majorityGenre(entries, games)
		if !ok {
			return nil, esportspb.ErrorEsportsMatchGenreUndecided("")
		}
		resolved = voted
	}
	configs, err := uc.matchRP.ListConfig(ctx, resolved)
	if err != nil {
		return nil, err
	}
	if len(configs) == 0 {
		return nil, esportspb.ErrorEsportsMatchGenreUnknown("").WithCause(fmt.Errorf("no match_config rows for genre %q", resolved))
	}
	tests, err := uc.testRP.List(ctx)
	if err != nil {
		return nil, err
	}
	prior, err := uc.matchRP.CountByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	budget := laterSessionBudget
	if prior == 0 {
		budget = firstSessionBudget
	}
	// The day boundary rides the server clock (deploy sets CN local) —
	// same day redraws identically, the next day never does.
	seed := drawSeed(userID, time.Now().Format("20060102"), resolved, salt)
	steps := arrange(configAbilities(configs), tests, resolved, seed)
	if len(steps) == 0 {
		// Config rows exist but nothing draws — a drifted catalog, not a
		// member-facing outcome.
		return nil, esportspb.ErrorEsportsUnspecified("").WithCause(fmt.Errorf("match: no drawable tests for genre %q", resolved))
	}
	return uc.matchRP.Create(ctx, &MatchSession{
		UserID:       userID,
		Genre:        resolved,
		Difficulty:   difficultyExperience,
		Seed:         seed,
		Steps:        steps,
		RetestBudget: budget,
		Status:       statusActive,
	})
}

func (uc *MatchUC) view(ctx context.Context, s *MatchSession, created bool) (*SessionView, error) {
	counts, err := uc.testRP.CountBySessionCodes(ctx, s.ID)
	if err != nil {
		return nil, err
	}
	done := make([]string, 0, len(s.Steps))
	left := make(map[string]int32, len(s.Steps))
	for _, st := range s.Steps {
		used := counts[st.TestCode]
		if used > 0 {
			done = append(done, st.TestCode)
		}
		// Every stored record spent a chance — validity never refunds one.
		remain := s.RetestBudget - used
		if remain < 0 {
			remain = 0
		}
		left[st.TestCode] = utils.ConvNum[int32](remain)
	}
	return &SessionView{
		Session:    s,
		Done:       done,
		RetestLeft: left,
		First:      s.RetestBudget == firstSessionBudget,
		Created:    created,
	}, nil
}

func configAbilities(configs []MatchConfig) []string {
	order := make([]string, 0, len(configs))
	for _, c := range configs {
		order = append(order, c.Ability)
	}
	return order
}
