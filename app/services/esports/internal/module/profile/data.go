package profile

import (
	"context"
	"log/slog"

	"cyber-ecosystem/shared-go/helper"
	"cyber-ecosystem/shared-go/utils"

	"cyber-ecosystem/app/services/esports/internal/ent"
	"cyber-ecosystem/app/services/esports/internal/ent/predicate"
	"cyber-ecosystem/app/services/esports/internal/ent/profile"
	"cyber-ecosystem/app/services/esports/internal/platform"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

// Repo ----------------------------------------------------------------------------------------------------------------

type profileRP struct {
	shared.RP
}

func NewProfileRP(logger *slog.Logger, p *platform.Platform) ProfileRP {
	return &profileRP{RP: shared.NewRP(logger.With("module", "module/profile_rp"), p)}
}

// Method --------------------------------------------------------------------------------------------------------------

func (rp *profileRP) FindByUserID(ctx context.Context, userID string) (*Profile, error) {
	row, err := rp.Platform.GetClient(ctx).Profile.Query().
		Where(profile.UserIDEQ(userID)).
		Only(ctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapProfile(row), nil
}

func (rp *profileRP) ListByUserIDs(ctx context.Context, userIDs []string) ([]*Profile, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}
	rows, err := rp.Platform.GetClient(ctx).Profile.Query().
		Where(profile.UserIDIn(userIDs...)).
		All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return utils.SliceMap(rows, mapProfile), nil
}

func (rp *profileRP) ListUserIDs(ctx context.Context, keyword, ageBand *string) ([]string, error) {
	query := rp.Platform.GetClient(ctx).Profile.Query()
	helper.Where(query, keyword != nil, func() predicate.Profile { return profile.NicknameContainsFold(*keyword) })
	helper.Where(query, ageBand != nil, func() predicate.Profile { return profile.AgeBandEQ(*ageBand) })
	ids, err := query.Unique(true).Select(profile.FieldUserID).Strings(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return ids, nil
}

func (rp *profileRP) Save(ctx context.Context, p *Profile) (*Profile, error) {
	client := rp.Platform.GetClient(ctx)
	var (
		row *ent.Profile
		err error
	)
	if p.ID == "" {
		row, err = client.Profile.Create().
			SetUserID(p.UserID).
			SetNillableNickname(p.Nickname).
			SetNillableBirthYear(utils.ConvPtr[int](p.BirthYear)).
			SetNillableBirthMonth(utils.ConvPtr[int](p.BirthMonth)).
			SetNillableAgeBand(p.AgeBand).
			SetNillableAvatarURL(p.AvatarURL).
			SetNillableGuardianConsentedAt(p.GuardianConsentedAt).
			SetGames(p.Games).
			Save(ctx)
	} else {
		row, err = client.Profile.UpdateOneID(p.ID).
			SetNillableNickname(p.Nickname).
			SetNillableBirthYear(utils.ConvPtr[int](p.BirthYear)).
			SetNillableBirthMonth(utils.ConvPtr[int](p.BirthMonth)).
			SetNillableAgeBand(p.AgeBand).
			SetNillableAvatarURL(p.AvatarURL).
			SetNillableGuardianConsentedAt(p.GuardianConsentedAt).
			SetGames(p.Games).
			Save(ctx)
	}
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapProfile(row), nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mapProfile(row *ent.Profile) *Profile {
	return &Profile{
		ID:                  row.ID,
		UserID:              row.UserID,
		Nickname:            row.Nickname,
		BirthYear:           utils.ConvPtr[int32](row.BirthYear),
		BirthMonth:          utils.ConvPtr[int32](row.BirthMonth),
		AgeBand:             row.AgeBand,
		AvatarURL:           row.AvatarURL,
		GuardianConsentedAt: row.GuardianConsentedAt,
		Games:               row.Games,
		CreatedAt:           row.CreatedAt,
		UpdatedAt:           row.UpdatedAt,
	}
}
