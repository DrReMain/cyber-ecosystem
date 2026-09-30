package auth

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	systempb "cyber-ecosystem/gen/go/cyber/system/v1"
)

// Struct --------------------------------------------------------------------------------------------------------------

type AuthService struct {
	systempb.UnimplementedAuthServiceServer

	log    *slog.Logger
	authUC *AuthUC
}

func NewAuthService(logger *slog.Logger, authUC *AuthUC) *AuthService {
	return &AuthService{
		log:    logger.With("module", "module/auth_service"),
		authUC: authUC,
	}
}

func (s *AuthService) RegisterGRPC(srv *grpc.Server) {
	systempb.RegisterAuthServiceServer(srv, s)
}

func (s *AuthService) RegisterHTTP(srv *http.Server) {
	systempb.RegisterAuthServiceHTTPServer(srv, s)
}

func (s *AuthService) RegisterConnect(srv *connecttransport.Server) {
	systempb.RegisterAuthServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *AuthService) Login(ctx context.Context, in *systempb.LoginRequest) (*systempb.LoginResponse, error) {
	token, issuedAt, expiresAt, err := s.authUC.Login(ctx, *in.Email, *in.Password)
	if err != nil {
		return nil, err
	}
	return &systempb.LoginResponse{
		SessionToken: token,
		IssuedAt:     utils.ToTimestamp(&issuedAt),
		ExpiresAt:    utils.ToTimestamp(&expiresAt),
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, in *systempb.LogoutRequest) (*systempb.LogoutResponse, error) {
	if err := s.authUC.Logout(ctx); err != nil {
		return nil, err
	}
	return &systempb.LogoutResponse{}, nil
}

func (s *AuthService) GetCurrentUser(ctx context.Context, in *systempb.GetCurrentUserRequest) (*systempb.GetCurrentUserResponse, error) {
	u, err := s.authUC.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	return &systempb.GetCurrentUserResponse{
		User: &systempb.User{
			Id:     utils.StringW(u.ID),
			Email:  utils.StringW(u.Email),
			Avatar: utils.Wrap(u.Avatar, utils.StringW),
		},
		Permissions: u.Permissions,
	}, nil
}
