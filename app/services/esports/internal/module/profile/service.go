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
	p := s.profileUC.Get(ctx)
	return &esportspb.GetProfileResponse{Profile: mapProfilePB(p)}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mapProfilePB(p *Profile) *esportspb.Profile {
	return &esportspb.Profile{
		Id: utils.StringW(p.ID),
	}
}
