package client

import (
	"context"
	"log/slog"
	"time"

	connectrpc "connectrpc.com/connect"

	"cyber-ecosystem/shared-go/kratos/security"
	krauth "cyber-ecosystem/shared-go/kratos/security/auth"
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

const (
	verifyTimeout  = 300 * time.Millisecond // a hung engine fails fast here, not at the inbound server ceiling
	checkTimeout   = 300 * time.Millisecond // fail-closed ceiling for a hung engine
	hydrateTimeout = 500 * time.Millisecond // degradable: a timeout renders absent fields, never a failed row
)

// Method --------------------------------------------------------------------------------------------------------------

func (c *systemClient) AuthenticateSession(ctx context.Context, token string) (*security.Subject, error) {
	ctx, cancel := context.WithTimeout(ctx, verifyTimeout)
	defer cancel()
	resp, err := c.introspect.VerifySession(withSessionToken(ctx, token), connectrpc.NewRequest(&systempb.VerifySessionRequest{}))
	if err != nil {
		return nil, err
	}
	return &security.Subject{
		UserID:     utils.Deref(utils.Unwrap[string](resp.Msg.GetUserId()), ""),
		TenantID:   utils.Deref(utils.Unwrap[string](resp.Msg.GetTenantId()), ""),
		Population: utils.Deref(utils.Unwrap[string](resp.Msg.GetPopulation()), ""),
		SessionID:  utils.Deref(utils.Unwrap[string](resp.Msg.GetSessionId()), ""),
		Extra:      map[string]any{shared.ExtraEmail: utils.Deref(utils.Unwrap[string](resp.Msg.GetEmail()), "")},
	}, nil
}

func (c *systemClient) CheckGrants(ctx context.Context, subject *security.Subject, operation string) (*kauthz.Decision, error) {
	ctx, cancel := context.WithTimeout(ctx, checkTimeout)
	defer cancel()
	resp, err := c.introspect.CheckOperation(withSessionToken(ctx, subject.SessionID), connectrpc.NewRequest(&systempb.CheckOperationRequest{Operation: operation}))
	if err != nil {
		return nil, err
	}
	return &kauthz.Decision{Allowed: resp.Msg.GetAllowed(), Reason: resp.Msg.GetReason()}, nil
}

func (c *systemClient) Hydrate(ctx context.Context, ids []string) map[string]*agentconfig.UserBrief {
	briefs := make(map[string]*agentconfig.UserBrief, len(ids))
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || len(ids) == 0 {
		return briefs
	}
	decision, err := c.CheckGrants(ctx, subject, systemv1connect.UserServiceBatchGetUsersProcedure)
	if err != nil {
		c.log.Warn("hydration eligibility degraded", "error", err)
		return briefs
	}
	if !decision.Allowed {
		c.log.Debug("hydration skipped: viewer ungranted")
		return briefs
	}
	ctx, cancel := context.WithTimeout(ctx, hydrateTimeout)
	defer cancel()
	resp, err := c.users.BatchGetUsers(withSessionToken(ctx, subject.SessionID), connectrpc.NewRequest(&systempb.BatchGetUsersRequest{Ids: ids}))
	if err != nil {
		if errorspb.IsGeneralErrorPermissionDenied(err) {
			c.log.Debug("user hydration skipped: viewer ungranted")
		} else {
			c.log.Warn("user hydration degraded", "error", err)
		}
		return briefs
	}
	for id, u := range resp.Msg.GetUsers() {
		briefs[id] = &agentconfig.UserBrief{
			ID:     utils.Deref(utils.Unwrap[string](u.GetId()), ""),
			Email:  utils.Unwrap[string](u.GetEmail()),
			Avatar: utils.Unwrap[string](u.GetAvatar()),
		}
	}
	return briefs
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
				req.Header().Set("Cookie", krauth.SessionCookieName+"="+token)
			}
			return next(ctx, req)
		}
	})
}
