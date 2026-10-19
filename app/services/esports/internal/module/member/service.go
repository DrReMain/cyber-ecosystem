package member

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	"cyber-ecosystem/shared-go/helper"
	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"

	"cyber-ecosystem/app/services/esports/internal/ent/schema/types"
)

// Struct --------------------------------------------------------------------------------------------------------------

type MemberService struct {
	esportspb.UnimplementedMemberServiceServer

	log      *slog.Logger
	memberUC *MemberUC
}

func NewMemberService(logger *slog.Logger, memberUC *MemberUC) *MemberService {
	return &MemberService{
		log:      logger.With("module", "module/member_service"),
		memberUC: memberUC,
	}
}

func (s *MemberService) RegisterGRPC(srv *grpc.Server) {
	esportspb.RegisterMemberServiceServer(srv, s)
}

func (s *MemberService) RegisterHTTP(srv *http.Server) {
	esportspb.RegisterMemberServiceHTTPServer(srv, s)
}

func (s *MemberService) RegisterConnect(srv *connecttransport.Server) {
	esportspb.RegisterMemberServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *MemberService) ListMembers(ctx context.Context, in *esportspb.ListMembersRequest) (*esportspb.ListMembersResponse, error) {
	out, err := s.memberUC.List(ctx, &ListIn{
		PageRequest: helper.EnsurePageRequest(in.Page),
		Keyword:     in.Keyword,
		Enabled:     in.Enabled,
		Population:  in.Population,
		AgeBand:     in.AgeBand,
	})
	if err != nil {
		return nil, err
	}
	return &esportspb.ListMembersResponse{
		Page:    out.PageResponse,
		Members: utils.SliceMap(out.Members, mapMemberPB),
	}, nil
}

func (s *MemberService) GetMember(ctx context.Context, in *esportspb.GetMemberRequest) (*esportspb.GetMemberResponse, error) {
	m, err := s.memberUC.Get(ctx, in.UserId)
	if err != nil {
		return nil, err
	}
	return &esportspb.GetMemberResponse{Member: mapMemberPB(m)}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mapMemberPB(m *Member) *esportspb.Member {
	pb := &esportspb.Member{
		UserId:       utils.StringW(m.ID),
		Account:      utils.Wrap(m.Account, utils.StringW),
		Avatar:       utils.Wrap(m.Avatar, utils.StringW),
		Enabled:      utils.BoolW(m.Enabled),
		RegisteredAt: utils.ToTimestamp(&m.CreatedAt),
	}
	if m.Profile != nil {
		pb.Nickname = utils.Wrap(m.Profile.Nickname, utils.StringW)
		pb.BirthYear = utils.Wrap(m.Profile.BirthYear, utils.Int32W)
		pb.BirthMonth = utils.Wrap(m.Profile.BirthMonth, utils.Int32W)
		pb.AgeBand = utils.Wrap(m.Profile.AgeBand, utils.StringW)
		pb.Games = mapGamesPB(m.Profile.Games)
	}
	return pb
}

func mapGamesPB(entries []types.GameEntry) []*esportspb.ProfileGame {
	games := make([]*esportspb.ProfileGame, 0, len(entries))
	for _, g := range entries {
		entry := &esportspb.ProfileGame{}
		if g.GameID != "" {
			entry.GameId = utils.StringW(g.GameID)
		}
		if g.CustomName != "" {
			entry.CustomName = utils.StringW(g.CustomName)
		}
		games = append(games, entry)
	}
	return games
}
