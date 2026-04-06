package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	"cyber-ecosystem/shared-go/kratos/transport/connect"

	singletonV1 "cyber-ecosystem/apps/singleton/gen/go/v1"
	singletonV1connect "cyber-ecosystem/apps/singleton/gen/go/v1/singletonV1connect"
	"cyber-ecosystem/apps/singleton/services/backend/internal/biz"
	"cyber-ecosystem/apps/singleton/services/backend/internal/pkg/auth"
)

type AccountDetailService struct {
	singletonV1.UnimplementedAccountDetailServiceServer
	log       *log.Helper
	accountUC *biz.AccountUC
}

func NewAccountDetailService(logger log.Logger, accountUC *biz.AccountUC) *AccountDetailService {
	return &AccountDetailService{
		log:       log.NewHelper(log.With(logger, "module", "service/account_detail")),
		accountUC: accountUC,
	}
}
func (s *AccountDetailService) RegisterGRPC(srv *grpc.Server) {
	singletonV1.RegisterAccountDetailServiceServer(srv, s)
}
func (s *AccountDetailService) RegisterHTTP(srv *http.Server) {
	singletonV1.RegisterAccountDetailServiceHTTPServer(srv, s)
}
func (s *AccountDetailService) RegisterConnect(srv *connect.Server) {
	srv.Register(singletonV1connect.NewAccountDetailServiceHandler(s, srv.HandlerOptions()...))
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *AccountDetailService) GetProfile(ctx context.Context, _ *singletonV1.GetProfileRequest) (*singletonV1.GetProfileResponse, error) {
	claims, err := auth.IdentityFromContext(ctx)
	if err != nil {
		return nil, err
	}
	profile, err := s.accountUC.GetProfile(ctx, claims.Subject)
	if err != nil {
		return nil, err
	}
	return &singletonV1.GetProfileResponse{
		Id:    *profile.ID,
		Email: *profile.Email,
	}, nil
}
