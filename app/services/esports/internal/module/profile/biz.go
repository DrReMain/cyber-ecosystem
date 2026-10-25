package profile

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"cyber-ecosystem/shared-go/kratos/security"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"

	"cyber-ecosystem/app/services/esports/internal/ent/schema/types"
	"cyber-ecosystem/app/services/esports/internal/module/game"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

const customGameCap = 3 // the product ruling's value, hardcoded by the config-scarcity ruling — changing it redeploys anyway; the contract's static 20 stays the backstop

// DO ------------------------------------------------------------------------------------------------------------------

type Profile struct {
	ID                  string
	UserID              string
	Nickname            *string
	BirthYear           *int32
	BirthMonth          *int32
	Sex                 *string
	AgeBand             *string
	AvatarURL           *string
	GuardianConsentedAt *time.Time
	Games               []types.GameEntry
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type GamePreference struct {
	GameIDs     []string
	CustomNames []string
}

type UpdateIn struct {
	Nickname         *string
	AvatarURL        *string
	BirthYear        *int32
	BirthMonth       *int32
	Sex              *string
	GuardianAccepted bool
	Games            *GamePreference // replace-set: presence rewrites the whole list
}

// Port ----------------------------------------------------------------------------------------------------------------

type ProfileRP interface {
	FindByUserID(ctx context.Context, userID string) (*Profile, error)
	ListByUserIDs(ctx context.Context, userIDs []string) ([]*Profile, error)
	ListUserIDs(ctx context.Context, keyword, ageBand *string) ([]string, error)
	Save(ctx context.Context, p *Profile) (*Profile, error)
}

// UC ------------------------------------------------------------------------------------------------------------------

type ProfileUC struct {
	shared.UC
	profileRP ProfileRP
	gameRP    game.GameRP
}

func NewProfileUC(logger *slog.Logger, tm shared.Transaction, profileRP ProfileRP, gameRP game.GameRP) *ProfileUC {
	return &ProfileUC{
		UC:        shared.NewUC(logger.With("module", "module/profile"), tm),
		profileRP: profileRP,
		gameRP:    gameRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *ProfileUC) Get(ctx context.Context) (*Profile, error) {
	subject, _ := security.SubjectFromCtx(ctx)
	if subject == nil {
		return &Profile{}, nil
	}
	p, err := uc.profileRP.FindByUserID(ctx, subject.UserID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		// Lazy row: born on first save, absent until then.
		return &Profile{UserID: subject.UserID}, nil
	}
	return p, nil
}

func (uc *ProfileUC) Update(ctx context.Context, in *UpdateIn) (*Profile, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("profile update requires a session"))
	}
	p, err := uc.profileRP.FindByUserID(ctx, subject.UserID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		p = &Profile{UserID: subject.UserID}
	}
	if in.Nickname != nil {
		p.Nickname = in.Nickname
	}
	if in.AvatarURL != nil {
		p.AvatarURL = in.AvatarURL
	}
	if in.BirthYear != nil {
		p.BirthYear = in.BirthYear
	}
	if in.BirthMonth != nil {
		p.BirthMonth = in.BirthMonth
	}
	if in.Sex != nil {
		p.Sex = in.Sex
	}
	// The band follows the birth pair wherever it lands — derived, never
	// client-supplied.
	p.AgeBand = ageBandOf(p.BirthYear, p.BirthMonth, time.Now())
	if age := derivedAge(p.BirthYear, p.BirthMonth, time.Now()); age != nil && *age < 13 {
		switch {
		case in.GuardianAccepted && p.GuardianConsentedAt == nil:
			stamped := time.Now()
			p.GuardianConsentedAt = &stamped
		case p.GuardianConsentedAt == nil:
			return nil, esportspb.ErrorEsportsProfileGuardianConsentRequired("")
		}
	}
	if in.Games != nil {
		entries, err := uc.buildGames(ctx, in.Games)
		if err != nil {
			return nil, err
		}
		p.Games = entries
	}
	return uc.profileRP.Save(ctx, p)
}

// Private -------------------------------------------------------------------------------------------------------------

func (uc *ProfileUC) buildGames(ctx context.Context, pref *GamePreference) ([]types.GameEntry, error) {
	if len(pref.CustomNames) > customGameCap {
		return nil, esportspb.ErrorEsportsProfileCustomGameCap("")
	}
	if len(pref.GameIDs) > 0 {
		existing, err := uc.gameRP.ExistingCodes(ctx, pref.GameIDs)
		if err != nil {
			return nil, err
		}
		for _, id := range pref.GameIDs {
			if _, ok := existing[id]; !ok {
				return nil, esportspb.ErrorEsportsProfileGameUnknown("").WithCause(fmt.Errorf("unknown game %q", id))
			}
		}
	}
	entries := make([]types.GameEntry, 0, len(pref.GameIDs)+len(pref.CustomNames))
	for _, id := range pref.GameIDs {
		entries = append(entries, types.GameEntry{GameID: id})
	}
	for _, name := range pref.CustomNames {
		entries = append(entries, types.GameEntry{CustomName: name})
	}
	return entries, nil
}

func derivedAge(year, month *int32, now time.Time) *int {
	if year == nil || month == nil {
		return nil
	}
	age := now.Year() - int(*year)
	if int(now.Month()) < int(*month) {
		age--
	}
	if age < 0 {
		age = 0
	}
	return &age
}

func ageBandOf(year, month *int32, now time.Time) *string {
	age := derivedAge(year, month, now)
	if age == nil {
		return nil
	}
	var band string
	switch a := *age; {
	case a < 12:
		band = "under12"
	case a <= 14:
		band = "12-14"
	case a <= 17:
		band = "15-17"
	case a <= 22:
		band = "18-22"
	case a <= 25:
		band = "23-25"
	case a <= 29:
		band = "26-29"
	case a <= 40:
		band = "30-40"
	default:
		band = "over40"
	}
	return &band
}
