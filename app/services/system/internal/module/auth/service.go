package auth

import (
	"context"
	"log/slog"
	"strings"

	"github.com/go-kratos/kratos/v3/transport"
	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"
	"google.golang.org/protobuf/types/known/wrapperspb"

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
	token, issuedAt, expiresAt, profile, err := s.authUC.LoginByIdentity(ctx, in.GetProvider(), in.GetState(), in)
	if err != nil {
		return nil, err
	}
	return &systempb.LoginByIdentityResponse{
		SessionToken:  token,
		IssuedAt:      utils.ToTimestamp(&issuedAt),
		ExpiresAt:     utils.ToTimestamp(&expiresAt),
		WechatProfile: toWireWechatProfile(profile),
	}, nil
}

func (s *AuthService) BeginIdentityLogin(ctx context.Context, in *systempb.BeginIdentityLoginRequest) (*systempb.BeginIdentityLoginResponse, error) {
	redirect, err := s.authUC.BeginIdentityLogin(ctx, in.GetProvider(), in.GetRedirectUri(), in.GetNext())
	if err != nil {
		return nil, err
	}
	return &systempb.BeginIdentityLoginResponse{AuthorizeUrl: redirect}, nil
}

func (s *AuthService) SendSmsCode(ctx context.Context, in *systempb.SendSmsCodeRequest) (*systempb.SendSmsCodeResponse, error) {
	state, err := s.authUC.SendSmsCode(ctx, in.GetProvider(), in.GetPhone(), callerIP(ctx))
	if err != nil {
		return nil, err
	}
	return &systempb.SendSmsCodeResponse{State: state}, nil
}

func (s *AuthService) GetLoginConfig(ctx context.Context, in *systempb.GetLoginConfigRequest) (*systempb.GetLoginConfigResponse, error) {
	providers, err := s.authUC.LoginConfig(ctx, in.GetPlane())
	if err != nil {
		return nil, err
	}
	return &systempb.GetLoginConfigResponse{Providers: providers}, nil
}

func (s *AuthService) BindCredential(ctx context.Context, in *systempb.BindCredentialRequest) (*systempb.BindCredentialResponse, error) {
	profile, err := s.authUC.BindCredential(ctx, in.GetProvider(), in.GetState(), in)
	if err != nil {
		return nil, err
	}
	return &systempb.BindCredentialResponse{WechatProfile: toWireWechatProfile(profile)}, nil
}

func (s *AuthService) BeginBindCredential(ctx context.Context, in *systempb.BeginBindCredentialRequest) (*systempb.BeginBindCredentialResponse, error) {
	redirect, err := s.authUC.BeginBindCredential(ctx, in.GetProvider(), in.GetRedirectUri())
	if err != nil {
		return nil, err
	}
	return &systempb.BeginBindCredentialResponse{AuthorizeUrl: redirect}, nil
}

func (s *AuthService) SendBindSmsCode(ctx context.Context, in *systempb.SendBindSmsCodeRequest) (*systempb.SendBindSmsCodeResponse, error) {
	// The bind twin shares SendSmsCode's body: the rung is the whole
	// difference — the authenticated mint binds the returned state to
	// this session, which the public face's anonymous mint never does.
	state, err := s.authUC.SendSmsCode(ctx, in.GetProvider(), in.GetPhone(), callerIP(ctx))
	if err != nil {
		return nil, err
	}
	return &systempb.SendBindSmsCodeResponse{State: state}, nil
}

func (s *AuthService) UnbindCredential(ctx context.Context, in *systempb.UnbindCredentialRequest) (*systempb.UnbindCredentialResponse, error) {
	if err := s.authUC.UnbindCredential(ctx, in.GetProvider()); err != nil {
		return nil, err
	}
	return &systempb.UnbindCredentialResponse{}, nil
}

func (s *AuthService) AdminUnbindCredential(ctx context.Context, in *systempb.AdminUnbindCredentialRequest) (*systempb.AdminUnbindCredentialResponse, error) {
	if err := s.authUC.AdminUnbindCredential(ctx, in.GetUserId(), in.GetProvider()); err != nil {
		return nil, err
	}
	return &systempb.AdminUnbindCredentialResponse{}, nil
}

func (s *AuthService) ResetPassword(ctx context.Context, in *systempb.ResetPasswordRequest) (*systempb.ResetPasswordResponse, error) {
	if err := s.authUC.ResetPassword(ctx, in.GetProvider(), in.GetState(), in.GetNewPassword(), in); err != nil {
		return nil, err
	}
	return &systempb.ResetPasswordResponse{}, nil
}

func (s *AuthService) ChangePassword(ctx context.Context, in *systempb.ChangePasswordRequest) (*systempb.ChangePasswordResponse, error) {
	if err := s.authUC.ChangePassword(ctx, *in.OldPassword, *in.NewPassword); err != nil {
		return nil, err
	}
	return &systempb.ChangePasswordResponse{}, nil
}

func (s *AuthService) GetSecurityStatus(ctx context.Context, in *systempb.GetSecurityStatusRequest) (*systempb.GetSecurityStatusResponse, error) {
	status, err := s.authUC.GetSecurityStatus(ctx)
	if err != nil {
		return nil, err
	}
	return &systempb.GetSecurityStatusResponse{
		HasPassword: status.HasPassword,
		Credentials: utils.SliceMap(status.Credentials, mapSecurityCredential),
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, in *systempb.LogoutRequest) (*systempb.LogoutResponse, error) {
	if err := s.authUC.Logout(ctx); err != nil {
		return nil, err
	}
	return &systempb.LogoutResponse{}, nil
}

func (s *AuthService) ImportAccounts(ctx context.Context, in *systempb.ImportAccountsRequest) (*systempb.ImportAccountsResponse, error) {
	out, err := s.authUC.ImportAccounts(ctx, &ImportAccountsIn{
		Rows: utils.SliceMap(in.GetRows(), func(r *systempb.ImportAccountRow) string { return r.GetPhone() }),
	})
	if err != nil {
		return nil, err
	}
	results := make([]*systempb.AccountImportResult, 0, len(out.Results))
	for _, r := range out.Results {
		results = append(results, &systempb.AccountImportResult{
			RowIndex: r.RowIndex,
			Phone:    r.Phone,
			UserId:   r.UserID,
			Status:   r.Status,
			Reason:   r.Reason,
		})
	}
	return &systempb.ImportAccountsResponse{
		Created: out.Created,
		Skipped: out.Skipped,
		Invalid: out.Invalid,
		Results: results,
	}, nil
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

// Private -------------------------------------------------------------------------------------------------------------

func callerIP(ctx context.Context) string {
	// The sender's network origin for the sms budget's source dimension:
	// the edge-terminated X-Forwarded-For's first hop (the deploy law has
	// the edge overwrite, not append, so the first hop is the client),
	// falling back to the connection's own address when no edge fronts
	// the service (dev). Feeds rate keys only, never authorization.
	if kt, ok := transport.FromServerContext(ctx); ok {
		if xff := kt.RequestHeader().Get("X-Forwarded-For"); xff != "" {
			return strings.TrimSpace(strings.SplitN(xff, ",", 2)[0])
		}
		if ra, ok := kt.(interface{ RemoteAddr() string }); ok {
			return ra.RemoteAddr()
		}
	}
	return ""
}

func mapSecurityCredential(c *SecurityCredential) *systempb.SecurityCredential {
	return &systempb.SecurityCredential{Kind: c.Kind, MaskedIdentifier: c.Identifier}
}

func toWireWechatProfile(p *wechatProfile) *systempb.WeChatProfile {
	if p == nil {
		return nil
	}
	return &systempb.WeChatProfile{
		Nickname:  wrapperspb.String(p.Nickname),
		AvatarUrl: wrapperspb.String(p.AvatarURL),
	}
}
