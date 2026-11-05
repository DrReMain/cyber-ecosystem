package client

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	connectrpc "connectrpc.com/connect"

	"cyber-ecosystem/shared-go/kratos/security"
	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"
	"cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
	systempb "cyber-ecosystem/gen/go/cyber/system/v1"
	systemv1connect "cyber-ecosystem/gen/go/cyber/system/v1/v1connect"

	"cyber-ecosystem/app/services/esports/internal/conf"
	"cyber-ecosystem/app/services/esports/internal/module/member"
	"cyber-ecosystem/app/services/esports/internal/module/profile"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

const (
	verifyTimeout = 300 * time.Millisecond // a hung engine fails fast here, not at the inbound server ceiling
	checkTimeout  = 300 * time.Millisecond // fail-closed ceiling for a hung engine

	smsProviderKey    = "sms"    // the system config's own instance key
	wechatProviderKey = "wechat" // ditto, the MP web-authorization row
	smsCredKind       = "sms"    // the exchange's registered credential families
	wechatCredKind    = "wechat_mp"
)

// Adapter -------------------------------------------------------------------------------------------------------------

type systemClient struct {
	log        *slog.Logger
	introspect systemv1connect.IntrospectServiceClient
	users      systemv1connect.UserServiceClient
	auth       systemv1connect.AuthServiceClient
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
		users:      systemv1connect.NewUserServiceClient(conn.HTTPClient(), conn.BaseURL(), conn.ClientOptions()...),
		auth:       systemv1connect.NewAuthServiceClient(conn.HTTPClient(), conn.BaseURL(), conn.ClientOptions()...),
	}, cleanup, nil
}

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
		Plane:      utils.Deref(utils.Unwrap[string](resp.Msg.GetPlane()), ""),
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

func (c *systemClient) ListRoster(ctx context.Context, q *member.RosterQuery) (*member.RosterPage, error) {
	ctx, err := withCallerSession(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := c.users.ListUsersByPopulation(ctx, connectrpc.NewRequest(&systempb.ListUsersByPopulationRequest{
		Population: &q.Population,
		Enabled:    q.Enabled,
		UserIds:    q.UserIDs,
		OrderBy:    q.OrderBy,
		Page:       q.Page,
	}))
	if err != nil {
		return nil, err
	}
	return &member.RosterPage{
		PageResponse: resp.Msg.GetPage(),
		Users:        utils.SliceMap(resp.Msg.GetList(), mapUser),
	}, nil
}

func (c *systemClient) ResolveCredential(ctx context.Context, cred *profile.MergeCredential) (*profile.CredentialFacts, error) {
	ctx, err := withCallerSession(ctx)
	if err != nil {
		return nil, err
	}
	req := &systempb.ResolveCredentialRequest{}
	if cred.Phone != nil {
		key := smsProviderKey
		req.Provider = &key
		req.Credential = &systempb.ResolveCredentialRequest_Phone{Phone: *cred.Phone}
	} else {
		key := wechatProviderKey
		req.Provider = &key
		state := cred.State
		req.State = &state
		req.Credential = &systempb.ResolveCredentialRequest_Wechat{
			Wechat: &systempb.WeChatIdentityPayload{Code: cred.WeChatCode},
		}
	}
	resp, err := c.auth.ResolveCredential(ctx, connectrpc.NewRequest(req))
	if err != nil {
		return nil, err
	}
	owner := resp.Msg.GetOwner()
	if owner == nil {
		return nil, nil
	}
	return mapCredentialFacts(owner), nil
}

func (c *systemClient) SecurityStatus(ctx context.Context) (*profile.SecurityFacts, error) {
	ctx, err := withCallerSession(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := c.auth.GetSecurityStatus(ctx, connectrpc.NewRequest(&systempb.GetSecurityStatusRequest{}))
	if err != nil {
		return nil, err
	}
	facts := &profile.SecurityFacts{HasPassword: resp.Msg.GetHasPassword(), CredentialCount: len(resp.Msg.GetCredentials())}
	for _, cred := range resp.Msg.GetCredentials() {
		switch cred.GetKind() {
		case smsCredKind:
			masked := cred.GetMaskedIdentifier()
			facts.PhoneMask = &masked
		case wechatCredKind:
			facts.Wechat = true
		}
	}
	return facts, nil
}

func (c *systemClient) WechatOpenID(ctx context.Context) (*string, error) {
	// Me-face by construction: the forwarded session is the order creator's
	// own, so the read can only ever name the payer's binding.
	ctx, err := withCallerSession(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := c.auth.GetWechatOpenId(ctx, connectrpc.NewRequest(&systempb.GetWechatOpenIdRequest{}))
	if err != nil {
		return nil, err
	}
	// Unwrap boxes the typed-nil wrapper into a non-nil interface and would
	// surface &"" — here nil-vs-empty is semantic: nil is the unbound payer.
	if v := resp.Msg.GetOpenid(); v != nil {
		id := v.GetValue()
		return &id, nil
	}
	return nil, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mapCredentialFacts(owner *systempb.CredentialOwner) *profile.CredentialFacts {
	facts := &profile.CredentialFacts{
		UserID:    owner.GetUserId(),
		Enabled:   owner.GetEnabled(),
		Mergeable: owner.GetMergeable(),
	}
	for _, cred := range owner.GetCredentials() {
		switch cred.GetKind() {
		case smsCredKind:
			masked := cred.GetMaskedIdentifier()
			facts.PhoneMask = &masked
		case wechatCredKind:
			facts.Wechat = true
		}
	}
	return facts
}

func withCallerSession(ctx context.Context) (context.Context, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("no subject in context"))
	}
	return connect.WithSessionToken(ctx, subject.SessionID), nil
}

func mapUser(u *systempb.User) *member.RosterUser {
	return &member.RosterUser{
		ID:          utils.Deref(utils.Unwrap[string](u.GetId()), ""),
		Account:     utils.Unwrap[string](u.GetAccount()),
		Avatar:      utils.Unwrap[string](u.GetAvatar()),
		Enabled:     utils.Deref(utils.Unwrap[bool](u.GetEnabled()), false),
		CreatedAt:   utils.Deref(utils.FromTimestamp(u.GetCreatedAt()), time.Time{}),
		PhoneMasked: utils.Unwrap[string](u.GetPhoneMasked()),
		WechatBound: utils.Deref(utils.Unwrap[bool](u.GetWechatBound()), false),
	}
}
