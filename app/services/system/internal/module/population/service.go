package population

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	"cyber-ecosystem/shared-go/helper"
	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	systempb "cyber-ecosystem/gen/go/cyber/system/v1"
)

// Struct --------------------------------------------------------------------------------------------------------------

type PopulationService struct {
	systempb.UnimplementedPopulationServiceServer

	log          *slog.Logger
	populationUC *PopulationUC
}

func NewPopulationService(logger *slog.Logger, populationUC *PopulationUC) *PopulationService {
	return &PopulationService{
		log:          logger.With("module", "module/population_service"),
		populationUC: populationUC,
	}
}

func (s *PopulationService) RegisterGRPC(srv *grpc.Server) {
	systempb.RegisterPopulationServiceServer(srv, s)
}

func (s *PopulationService) RegisterHTTP(srv *http.Server) {
	systempb.RegisterPopulationServiceHTTPServer(srv, s)
}

func (s *PopulationService) RegisterConnect(srv *connecttransport.Server) {
	systempb.RegisterPopulationServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *PopulationService) ListPopulations(ctx context.Context, in *systempb.ListPopulationsRequest) (*systempb.ListPopulationsResponse, error) {
	out, err := s.populationUC.List(ctx, &PopulationListIn{
		PageRequest: helper.EnsurePageRequest(in.Page),
		OrderBy:     in.OrderBy,
		Code:        in.Code,
		Name:        in.Name,
	})
	if err != nil {
		return nil, err
	}
	return &systempb.ListPopulationsResponse{
		Page: out.PageResponse,
		List: utils.SliceMap(out.List, toProtoPopulation),
	}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func toProtoPopulation(p *Population) *systempb.Population {
	return &systempb.Population{
		Id:        utils.StringW(p.ID),
		Code:      utils.StringW(p.Code),
		Name:      utils.StringW(p.Name),
		Plane:     utils.StringW(p.Plane),
		CreatedAt: utils.ToTimestamp(&p.CreatedAt),
		UpdatedAt: utils.ToTimestamp(&p.UpdatedAt),
	}
}
