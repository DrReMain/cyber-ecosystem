package service

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"

	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/biz"
)

// Struct --------------------------------------------------------------------------------------------------------------

type AuthService struct {
	systempb.UnimplementedAuthServiceServer

	log    *slog.Logger
	authUC *biz.AuthUC
}

func NewAuthService(logger *slog.Logger, authUC *biz.AuthUC) *AuthService {
	return &AuthService{
		log:    logger.With("module", "service/auth"),
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
	token, err := s.authUC.Login(ctx, *in.Email, *in.Password)
	if err != nil {
		return nil, err
	}
	return &systempb.LoginResponse{Token: token}, nil
}
