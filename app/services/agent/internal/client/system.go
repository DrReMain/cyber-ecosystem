package client

import (
	"context"
	"log/slog"

	connectrpc "connectrpc.com/connect"

	"cyber-ecosystem/shared-go/kratos/security"
	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"
	"cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
	systempb "cyber-ecosystem/gen/go/cyber/system/v1"
	systemv1connect "cyber-ecosystem/gen/go/cyber/system/v1/v1connect"

	"cyber-ecosystem/app/services/agent/internal/conf"
	"cyber-ecosystem/app/services/agent/internal/module/agentconfig"
	"cyber-ecosystem/app/services/agent/internal/shared"
)

// Adapter -------------------------------------------------------------------------------------------------------------

type systemClient struct {
	log        *slog.Logger
	introspect systemv1connect.IntrospectServiceClient
	users      systemv1connect.UserServiceClient
}

func NewSystemClient(c *conf.Remote, logger *slog.Logger) (*systemClient, func(), error) {
	conn, err := connect.DialInsecure(context.Background(),
		connect.WithEndpoint(c.GetSystemConnect()),
		connect.WithMiddleware(standardMiddleware(logger)...),
		connect.WithInterceptors(sessionCookieInterceptor()),
	)
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() { _ = conn.Close() }
	return &systemClient{
		log:        logger.With("module", "client/system"),
		introspect: systemv1connect.NewIntrospectServiceClient(conn.HTTPClient(), conn.BaseURL(), conn.ClientOptions()...),
		users:      systemv1connect.NewUserServiceClient(conn.HTTPClient(), conn.BaseURL(), conn.ClientOptions()...),
	}, cleanup, nil
}

// Method --------------------------------------------------------------------------------------------------------------

func (c *systemClient) AuthenticateSession(ctx context.Context, token string) (*security.Subject, error) {
	resp, err := c.introspect.VerifySession(withSessionToken(ctx, token), connectrpc.NewRequest(&systempb.VerifySessionRequest{}))
	if err != nil {
		return nil, err
	}
	return &security.Subject{
		UserID:    utils.Deref(utils.Unwrap[string](resp.Msg.GetUserId()), ""),
		TenantID:  utils.Deref(utils.Unwrap[string](resp.Msg.GetTenantId()), ""),
		SessionID: utils.Deref(utils.Unwrap[string](resp.Msg.GetSessionId()), ""),
		Extra:     map[string]any{shared.ExtraEmail: utils.Deref(utils.Unwrap[string](resp.Msg.GetEmail()), "")},
	}, nil
}

func (c *systemClient) CheckGrants(ctx context.Context, subject *security.Subject, operation string) (*kauthz.Decision, error) {
	resp, err := c.introspect.CheckOperation(withSessionToken(ctx, subject.SessionID), connectrpc.NewRequest(&systempb.CheckOperationRequest{Operation: operation}))
	if err != nil {
		return nil, err
	}
	return &kauthz.Decision{Allowed: resp.Msg.GetAllowed(), Reason: resp.Msg.GetReason()}, nil
}

func (c *systemClient) FindByID(ctx context.Context, id string) (*agentconfig.UserBrief, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok {
		return nil, nil
	}
	resp, err := c.users.GetUser(withSessionToken(ctx, subject.SessionID), connectrpc.NewRequest(&systempb.GetUserRequest{Id: id}))
	if err != nil {
		if errorspb.IsGeneralErrorPermissionDenied(err) {
			c.log.Debug("user hydration skipped: viewer ungranted", "user_id", id)
		} else {
			c.log.Warn("user hydration degraded", "user_id", id, "error", err)
		}
		return nil, nil
	}
	u := resp.Msg.GetUser()
	if u == nil {
		return nil, nil
	}
	return &agentconfig.UserBrief{
		ID:     utils.Deref(utils.Unwrap[string](u.GetId()), ""),
		Email:  utils.Unwrap[string](u.GetEmail()),
		Avatar: utils.Unwrap[string](u.GetAvatar()),
	}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

type sessionTokenKey struct{}

func withSessionToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, sessionTokenKey{}, token)
}

func sessionCookieInterceptor() connectrpc.Interceptor {
	return connectrpc.UnaryInterceptorFunc(func(next connectrpc.UnaryFunc) connectrpc.UnaryFunc {
		return func(ctx context.Context, req connectrpc.AnyRequest) (connectrpc.AnyResponse, error) {
			if token, ok := ctx.Value(sessionTokenKey{}).(string); ok && token != "" {
				req.Header().Set("Cookie", "session_token="+token)
			}
			return next(ctx, req)
		}
	})
}
