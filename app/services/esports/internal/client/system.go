package client

import (
	"context"
	"log/slog"
	"time"

	connectrpc "connectrpc.com/connect"

	"cyber-ecosystem/shared-go/kratos/security"
	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"
	"cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	systempb "cyber-ecosystem/gen/go/cyber/system/v1"
	systemv1connect "cyber-ecosystem/gen/go/cyber/system/v1/v1connect"

	"cyber-ecosystem/app/services/esports/internal/conf"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

// Adapter -------------------------------------------------------------------------------------------------------------

type systemClient struct {
	log        *slog.Logger
	introspect systemv1connect.IntrospectServiceClient
}

func NewSystemClient(c *conf.Remote, logger *slog.Logger) (*systemClient, func(), error) {
	conn, err := connect.DialInsecure(context.Background(),
		connect.WithEndpoint(c.GetSystemConnect()),
		connect.WithMiddleware(standardMiddleware(logger)...),
		connect.WithInterceptors(connect.SessionCookieInterceptor()),
	)
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() { _ = conn.Close() }
	return &systemClient{
		log:        logger.With("module", "client/system"),
		introspect: systemv1connect.NewIntrospectServiceClient(conn.HTTPClient(), conn.BaseURL(), conn.ClientOptions()...),
	}, cleanup, nil
}

const (
	verifyTimeout = 300 * time.Millisecond // a hung engine fails fast here, not at the inbound server ceiling
	checkTimeout  = 300 * time.Millisecond // fail-closed ceiling for a hung engine
)

// Method --------------------------------------------------------------------------------------------------------------

func (c *systemClient) AuthenticateSession(ctx context.Context, token string) (*security.Subject, error) {
	ctx, cancel := context.WithTimeout(ctx, verifyTimeout)
	defer cancel()
	resp, err := c.introspect.VerifySession(connect.WithSessionToken(ctx, token), connectrpc.NewRequest(&systempb.VerifySessionRequest{}))
	if err != nil {
		return nil, err
	}
	return &security.Subject{
		UserID:     utils.Deref(utils.Unwrap[string](resp.Msg.GetUserId()), ""),
		TenantID:   utils.Deref(utils.Unwrap[string](resp.Msg.GetTenantId()), ""),
		Population: utils.Deref(utils.Unwrap[string](resp.Msg.GetPopulation()), ""),
		SessionID:  utils.Deref(utils.Unwrap[string](resp.Msg.GetSessionId()), ""),
		Extra:      map[string]any{shared.ExtraAccount: utils.Deref(utils.Unwrap[string](resp.Msg.GetAccount()), "")},
	}, nil
}

func (c *systemClient) CheckGrants(ctx context.Context, subject *security.Subject, operation string) (*kauthz.Decision, error) {
	ctx, cancel := context.WithTimeout(ctx, checkTimeout)
	defer cancel()
	resp, err := c.introspect.CheckOperation(connect.WithSessionToken(ctx, subject.SessionID), connectrpc.NewRequest(&systempb.CheckOperationRequest{Operation: operation}))
	if err != nil {
		return nil, err
	}
	return &kauthz.Decision{Allowed: resp.Msg.GetAllowed(), Reason: resp.Msg.GetReason()}, nil
}
