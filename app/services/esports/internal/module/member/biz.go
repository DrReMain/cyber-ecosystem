package member

import (
	"context"
	"log/slog"
	"time"

	"cyber-ecosystem/shared-go/helper"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
	commonv1 "cyber-ecosystem/gen/go/cyber/shared/common/v1"

	"cyber-ecosystem/app/services/esports/internal/module/profile"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

const defaultPopulation = "member"

// DO ------------------------------------------------------------------------------------------------------------------

type RosterUser struct {
	ID        string
	Account   *string
	Avatar    *string
	Enabled   bool
	CreatedAt time.Time
}

type RosterQuery struct {
	Population string
	Enabled    *bool
	UserIDs    []string
	OrderBy    []string
	Page       *commonv1.PageRequest
}

type RosterPage struct {
	*commonv1.PageResponse
	Users []*RosterUser
}

type Member struct {
	RosterUser
	Profile *profile.Profile
}

type ListIn struct {
	*commonv1.PageRequest
	Keyword    *string
	Enabled    *bool
	Population *string
	AgeBand    *string
}

type ListOut struct {
	*commonv1.PageResponse
	Members []*Member
}

// Port ----------------------------------------------------------------------------------------------------------------

type SystemRP interface {
	ListRoster(ctx context.Context, q *RosterQuery) (*RosterPage, error)
}

// UC ------------------------------------------------------------------------------------------------------------------

type MemberUC struct {
	shared.UC
	systemRP  SystemRP
	profileRP profile.ProfileRP
}

func NewMemberUC(logger *slog.Logger, tm shared.Transaction, systemRP SystemRP, profileRP profile.ProfileRP) *MemberUC {
	return &MemberUC{
		UC:        shared.NewUC(logger.With("module", "module/member"), tm),
		systemRP:  systemRP,
		profileRP: profileRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *MemberUC) List(ctx context.Context, in *ListIn) (*ListOut, error) {
	q := &RosterQuery{
		Population: defaultPop(in.Population),
		Enabled:    in.Enabled,
		OrderBy:    []string{"createdAt:desc"},
		Page:       in.PageRequest,
	}
	if in.Keyword != nil || in.AgeBand != nil {
		// Profile-local filters resolve to an id set first, so the roster
		// query keeps filtering, ordering, and pagination at the query
		// level; a lazy profile matches neither filter.
		ids, err := uc.profileRP.ListUserIDs(ctx, in.Keyword, in.AgeBand)
		if err != nil {
			return nil, err
		}
		if len(ids) == 0 {
			return &ListOut{PageResponse: helper.BuildPageResponse(0, 0, 0)}, nil
		}
		q.UserIDs = ids
	}
	roster, err := uc.systemRP.ListRoster(ctx, q)
	if err != nil {
		return nil, err
	}
	profiles, err := uc.profileRP.ListByUserIDs(ctx, rosterIDs(roster.Users))
	if err != nil {
		return nil, err
	}
	byUser := make(map[string]*profile.Profile, len(profiles))
	for _, p := range profiles {
		byUser[p.UserID] = p
	}
	members := make([]*Member, 0, len(roster.Users))
	for _, r := range roster.Users {
		members = append(members, &Member{RosterUser: *r, Profile: byUser[r.ID]})
	}
	return &ListOut{PageResponse: roster.PageResponse, Members: members}, nil
}

func (uc *MemberUC) Get(ctx context.Context, userID string) (*Member, error) {
	roster, err := uc.systemRP.ListRoster(ctx, &RosterQuery{
		Population: defaultPopulation,
		UserIDs:    []string{userID},
		Page:       &commonv1.PageRequest{},
	})
	if err != nil {
		return nil, err
	}
	if len(roster.Users) == 0 {
		return nil, esportspb.ErrorEsportsMemberNotFound("")
	}
	p, err := uc.profileRP.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &Member{RosterUser: *roster.Users[0], Profile: p}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func defaultPop(population *string) string {
	if population == nil || *population == "" {
		return defaultPopulation
	}
	return *population
}

func rosterIDs(users []*RosterUser) []string {
	ids := make([]string, 0, len(users))
	for _, r := range users {
		ids = append(ids, r.ID)
	}
	return ids
}
