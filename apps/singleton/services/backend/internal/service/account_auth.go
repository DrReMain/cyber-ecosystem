package service

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	"cyber-ecosystem/shared-go/kratos/transport/connect"

	singletonV1 "cyber-ecosystem/apps/singleton/gen/go/v1"
	singletonV1connect "cyber-ecosystem/apps/singleton/gen/go/v1/singletonV1connect"
	"cyber-ecosystem/apps/singleton/services/backend/internal/biz"
	"cyber-ecosystem/apps/singleton/services/backend/internal/pkg/auth"
)

type AccountAuthService struct {
	singletonV1.UnimplementedAccountAuthServiceServer
	log       *log.Helper
	accountUC *biz.AccountUC
}

func NewAccountAuthService(logger log.Logger, accountUC *biz.AccountUC) *AccountAuthService {
	return &AccountAuthService{
		log:       log.NewHelper(log.With(logger, "module", "service/account_auth")),
		accountUC: accountUC,
	}
}
func (s *AccountAuthService) RegisterGRPC(srv *grpc.Server) {
	singletonV1.RegisterAccountAuthServiceServer(srv, s)
}
func (s *AccountAuthService) RegisterHTTP(srv *http.Server) {
	singletonV1.RegisterAccountAuthServiceHTTPServer(srv, s)
}
func (s *AccountAuthService) RegisterConnect(srv *connect.Server) {
	srv.Register(singletonV1connect.NewAccountAuthServiceHandler(s, srv.HandlerOptions()...))
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *AccountAuthService) LoginPassword(ctx context.Context, in *singletonV1.LoginPasswordRequest) (*singletonV1.LoginPasswordResponse, error) {
	pair, err := s.accountUC.LoginPassword(ctx, *in.Email, *in.Password)
	if err != nil {
		return nil, err
	}
	return &singletonV1.LoginPasswordResponse{
		AccessToken:  &singletonV1.AccessToken{Token: pair.AccessToken.Value, Expire: timestamppb.New(pair.AccessToken.ExpireAt)},
		RefreshToken: &singletonV1.RefreshToken{Token: pair.RefreshToken.Value, Expire: timestamppb.New(pair.RefreshToken.ExpireAt)},
	}, nil
}

func (s *AccountAuthService) RefreshToken(ctx context.Context, in *singletonV1.RefreshTokenRequest) (*singletonV1.RefreshTokenResponse, error) {
	pair, err := s.accountUC.RefreshToken(ctx, *in.RefreshToken)
	if err != nil {
		return nil, err
	}
	return &singletonV1.RefreshTokenResponse{
		AccessToken:  &singletonV1.AccessToken{Token: pair.AccessToken.Value, Expire: timestamppb.New(pair.AccessToken.ExpireAt)},
		RefreshToken: &singletonV1.RefreshToken{Token: pair.RefreshToken.Value, Expire: timestamppb.New(pair.RefreshToken.ExpireAt)},
	}, nil
}

func (s *AccountAuthService) Logout(ctx context.Context, _ *singletonV1.LogoutRequest) (*singletonV1.LogoutResponse, error) {
	claims, err := auth.IdentityFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.accountUC.Logout(ctx, claims); err != nil {
		return nil, err
	}
	return &singletonV1.LogoutResponse{}, nil
}
