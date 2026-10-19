package profile

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
)

// Struct --------------------------------------------------------------------------------------------------------------

type ProfileService struct {
	esportspb.UnimplementedProfileServiceServer

	log       *slog.Logger
	profileUC *ProfileUC
}

func NewProfileService(logger *slog.Logger, profileUC *ProfileUC) *ProfileService {
	return &ProfileService{
		log:       logger.With("module", "module/profile_service"),
		profileUC: profileUC,
	}
}

func (s *ProfileService) RegisterGRPC(srv *grpc.Server) {
	esportspb.RegisterProfileServiceServer(srv, s)
}

func (s *ProfileService) RegisterHTTP(srv *http.Server) {
	esportspb.RegisterProfileServiceHTTPServer(srv, s)
}

func (s *ProfileService) RegisterConnect(srv *connecttransport.Server) {
	esportspb.RegisterProfileServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *ProfileService) GetProfile(ctx context.Context, in *esportspb.GetProfileRequest) (*esportspb.GetProfileResponse, error) {
	p, err := s.profileUC.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &esportspb.GetProfileResponse{Profile: mapProfilePB(p)}, nil
}

func (s *ProfileService) UpdateProfile(ctx context.Context, in *esportspb.UpdateProfileRequest) (*esportspb.UpdateProfileResponse, error) {
	upd := &UpdateIn{
		Nickname:         in.Nickname,
		AvatarURL:        in.AvatarUrl,
		BirthYear:        in.BirthYear,
		BirthMonth:       in.BirthMonth,
		GuardianAccepted: in.GetGuardianAccepted(),
	}
	if in.GamePreference != nil {
		upd.Games = &GamePreference{
			GameIDs:     in.GamePreference.GameIds,
			CustomNames: in.GamePreference.CustomNames,
		}
	}
	p, err := s.profileUC.Update(ctx, upd)
	if err != nil {
		return nil, err
	}
	return &esportspb.UpdateProfileResponse{Profile: mapProfilePB(p)}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mapProfilePB(p *Profile) *esportspb.Profile {
	games := make([]*esportspb.ProfileGame, 0, len(p.Games))
	for _, g := range p.Games {
		entry := &esportspb.ProfileGame{}
		if g.GameID != "" {
			entry.GameId = utils.StringW(g.GameID)
		}
		if g.CustomName != "" {
			entry.CustomName = utils.StringW(g.CustomName)
		}
		games = append(games, entry)
	}
	return &esportspb.Profile{
		Id:                  utils.StringW(p.UserID),
		Nickname:            utils.Wrap(p.Nickname, utils.StringW),
		BirthYear:           utils.Wrap(p.BirthYear, utils.Int32W),
		BirthMonth:          utils.Wrap(p.BirthMonth, utils.Int32W),
		AgeBand:             utils.Wrap(p.AgeBand, utils.StringW),
		AvatarUrl:           utils.Wrap(p.AvatarURL, utils.StringW),
		GuardianConsentedAt: utils.ToTimestamp(p.GuardianConsentedAt),
		Games:               games,
		CreatedAt:           utils.ToTimestamp(&p.CreatedAt),
		UpdatedAt:           utils.ToTimestamp(&p.UpdatedAt),
	}
}
