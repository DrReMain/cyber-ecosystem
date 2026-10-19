package game

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

type GameService struct {
	esportspb.UnimplementedGameServiceServer

	log    *slog.Logger
	gameUC *GameUC
}

func NewGameService(logger *slog.Logger, gameUC *GameUC) *GameService {
	return &GameService{
		log:    logger.With("module", "module/game_service"),
		gameUC: gameUC,
	}
}

func (s *GameService) RegisterGRPC(srv *grpc.Server) {
	esportspb.RegisterGameServiceServer(srv, s)
}

func (s *GameService) RegisterHTTP(srv *http.Server) {
	esportspb.RegisterGameServiceHTTPServer(srv, s)
}

func (s *GameService) RegisterConnect(srv *connecttransport.Server) {
	esportspb.RegisterGameServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *GameService) ListGames(ctx context.Context, in *esportspb.ListGamesRequest) (*esportspb.ListGamesResponse, error) {
	var category *string
	if in.Category != nil {
		category = in.Category
	}
	games, err := s.gameUC.List(ctx, category)
	if err != nil {
		return nil, err
	}
	out := make([]*esportspb.Game, 0, len(games))
	for _, g := range games {
		out = append(out, mapGamePB(g))
	}
	return &esportspb.ListGamesResponse{Games: out}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mapGamePB(g *Game) *esportspb.Game {
	pb := &esportspb.Game{
		// The contract's id IS the code join key (id "lol"), the same law
		// as tenant/population codes; the row ULID never crosses the wire.
		Id:          utils.StringW(g.Code),
		Name:        utils.StringW(g.Name),
		Category:    utils.Wrap(g.Category, utils.StringW),
		IconUrl:     utils.Wrap(g.IconURL, utils.StringW),
		Description: utils.Wrap(g.Description, utils.StringW),
		CreatedAt:   utils.ToTimestamp(&g.CreatedAt),
		UpdatedAt:   utils.ToTimestamp(&g.UpdatedAt),
	}
	return pb
}
