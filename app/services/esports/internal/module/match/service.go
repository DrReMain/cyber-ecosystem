package match

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

type MatchService struct {
	esportspb.UnimplementedMatchServiceServer

	log     *slog.Logger
	matchUC *MatchUC
}

func NewMatchService(logger *slog.Logger, matchUC *MatchUC) *MatchService {
	return &MatchService{
		log:     logger.With("module", "module/match_service"),
		matchUC: matchUC,
	}
}

func (s *MatchService) RegisterGRPC(srv *grpc.Server) {
	esportspb.RegisterMatchServiceServer(srv, s)
}

func (s *MatchService) RegisterHTTP(srv *http.Server) {
	esportspb.RegisterMatchServiceHTTPServer(srv, s)
}

func (s *MatchService) RegisterConnect(srv *connecttransport.Server) {
	esportspb.RegisterMatchServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *MatchService) GetMatchSession(ctx context.Context, in *esportspb.GetMatchSessionRequest) (*esportspb.GetMatchSessionResponse, error) {
	// The optional genre's presence is the vote/no-vote signal — the plain
	// getter would collapse nil into "" and silently vote on an empty word.
	v, err := s.matchUC.Get(ctx, in.Genre)
	if err != nil {
		return nil, err
	}
	return &esportspb.GetMatchSessionResponse{Session: mapSessionPB(v), Created: v.Created}, nil
}

func (s *MatchService) SwitchMatchGenre(ctx context.Context, in *esportspb.SwitchMatchGenreRequest) (*esportspb.SwitchMatchGenreResponse, error) {
	v, err := s.matchUC.Switch(ctx, in.GetGenre())
	if err != nil {
		return nil, err
	}
	return &esportspb.SwitchMatchGenreResponse{Session: mapSessionPB(v)}, nil
}

func (s *MatchService) PeekMatchSession(ctx context.Context, _ *esportspb.PeekMatchSessionRequest) (*esportspb.PeekMatchSessionResponse, error) {
	v, err := s.matchUC.Peek(ctx)
	if err != nil {
		return nil, err
	}
	if v == nil {
		// Absent session = the response's whole truth (message-field
		// presence is the null/zero contract); no error channel for it.
		return &esportspb.PeekMatchSessionResponse{}, nil
	}
	return &esportspb.PeekMatchSessionResponse{Session: mapSessionPB(v)}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mapSessionPB(v *SessionView) *esportspb.MatchSession {
	steps := make([]*esportspb.MatchStep, 0, len(v.Session.Steps))
	for _, st := range v.Session.Steps {
		steps = append(steps, &esportspb.MatchStep{
			Ability:  utils.StringW(st.Ability),
			TestCode: utils.StringW(st.TestCode),
		})
	}
	return &esportspb.MatchSession{
		Id:           utils.StringW(v.Session.ID),
		Genre:        utils.StringW(v.Session.Genre),
		Difficulty:   utils.StringW(v.Session.Difficulty),
		Steps:        steps,
		Done:         v.Done,
		RetestLeft:   v.RetestLeft,
		CreatedAt:      utils.ToTimestamp(&v.Session.CreatedAt),
		FirstSession:   utils.BoolW(v.First),
		AnchorRecordId: utils.Wrap(v.AnchorRecordID, utils.StringW),
	}
}
