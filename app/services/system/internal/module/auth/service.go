package auth

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/module/user"
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
	token, issuedAt, expiresAt, err := s.authUC.Login(ctx, *in.Account, *in.Password, in.GetPopulation())
	if err != nil {
		return nil, err
	}
	return &systempb.LoginResponse{
		SessionToken: token,
		IssuedAt:     utils.ToTimestamp(&issuedAt),
		ExpiresAt:    utils.ToTimestamp(&expiresAt),
	}, nil
}

func (s *AuthService) LoginByIdentity(ctx context.Context, in *systempb.LoginByIdentityRequest) (*systempb.LoginByIdentityResponse, error) {
	token, issuedAt, expiresAt, err := s.authUC.LoginByIdentity(ctx, in.GetProvider(), in.GetState(), in)
	if err != nil {
		return nil, err
	}
	return &systempb.LoginByIdentityResponse{
		SessionToken: token,
		IssuedAt:     utils.ToTimestamp(&issuedAt),
		ExpiresAt:    utils.ToTimestamp(&expiresAt),
	}, nil
}

func (s *AuthService) BeginIdentityLogin(ctx context.Context, in *systempb.BeginIdentityLoginRequest) (*systempb.BeginIdentityLoginResponse, error) {
	redirect, err := s.authUC.BeginIdentityLogin(ctx, in.GetProvider(), in.GetRedirectUri())
	if err != nil {
		return nil, err
	}
	return &systempb.BeginIdentityLoginResponse{AuthorizeUrl: redirect}, nil
}

func (s *AuthService) SendSmsCode(ctx context.Context, in *systempb.SendSmsCodeRequest) (*systempb.SendSmsCodeResponse, error) {
	state, err := s.authUC.SendSmsCode(ctx, in.GetProvider(), in.GetPhone())
	if err != nil {
		return nil, err
	}
	return &systempb.SendSmsCodeResponse{State: state}, nil
}

func (s *AuthService) BindCredential(ctx context.Context, in *systempb.BindCredentialRequest) (*systempb.BindCredentialResponse, error) {
	if err := s.authUC.BindCredential(ctx, in.GetProvider(), in.GetState(), in); err != nil {
		return nil, err
	}
	return &systempb.BindCredentialResponse{}, nil
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
			Id:      utils.StringW(u.ID),
			Account: utils.StringW(u.Account),
			Avatar:  utils.Wrap(u.Avatar, utils.StringW),
		},
		Permissions: u.Permissions,
		Memberships: utils.SliceMap(u.Seats, func(seat *user.Seat) *systempb.Membership {
			return &systempb.Membership{Tenant: seat.TenantID, Population: seat.Population}
		}),
	}, nil
}
