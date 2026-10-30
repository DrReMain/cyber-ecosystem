package auth

import (
	"context"
	"fmt"

	"cyber-ecosystem/shared-go/kratos/security"
	"cyber-ecosystem/shared-go/utils"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/conf"
)

type identityPayload interface {
	GetWechat() *systempb.WeChatIdentityPayload
}

type smsPayload interface {
	GetSms() *systempb.SmsIdentityPayload
}

type exchangedIdentity struct {
	Kind       string
	Identifier string
	Profile    *wechatProfile
}

type wechatProfile struct {
	// The person behind the openid as the oauth leg learned them — wechat's
	// raw values, pre-sanitisation. Nil means the userinfo leg degraded (or
	// the kind reports no person at all); degraded never fails the exchange
	// that owns it.
	Nickname  string
	AvatarURL string
}

type IdentityProvider interface {
	Exchange(ctx context.Context, row *conf.Auth_Provider, payload any) (*exchangedIdentity, error)
}

type redirectProvider interface {
	IdentityProvider
	AuthorizeURL(row *conf.Auth_Provider, redirectURI, state string) string
}

type codeIssuer interface {
	IssueCode(ctx context.Context, row *conf.Auth_Provider, phone string) (state string, err error)
}

func NewIdentityProviders(wechatP wechatProvider, smsP *smsProvider) map[string]IdentityProvider {
	return map[string]IdentityProvider{
		wechatKind: wechatP,
		smsKind:    smsP,
	}
}

type identityStateValue struct {
	// The provider whose flow minted the state.
	Provider string `json:"provider"`
	// The minting session's user on authenticated rungs; empty on the
	// public login rungs. Bind faces refuse a state whose BindUser is not
	// the caller; login faces ignore the binding (the exchanged code is
	// the login proof).
	BindUser string `json:"bind_user,omitempty"`
}

func mintIdentityState(ctx context.Context, trp TokenRP, providerKey string) (string, error) {
	// The one state-minting path every provider shares: Begin mints it
	// for redirect flows, IssueCode for direct-code ones. A mint on an
	// authenticated rung (bind flows) carries the session; the public
	// login rungs mint anonymous.
	val := identityStateValue{Provider: providerKey}
	if subject, ok := security.SubjectFromCtx(ctx); ok {
		val.BindUser = subject.UserID
	}
	state, err := utils.RandomToken()
	if err != nil {
		return "", err
	}
	if err := trp.Set(ctx, identityStatePrefix+state, utils.MustMarshal(val), identityStateTTL); err != nil {
		return "", err
	}
	return state, nil
}

func loadIdentityState(ctx context.Context, trp TokenRP, providerKey, state string) (identityStateValue, error) {
	stored, err := trp.GetDel(ctx, identityStatePrefix+state)
	if err != nil {
		if errorspb.IsInfraErrorCacheMiss(err) {
			return identityStateValue{}, systempb.ErrorSystemIdentityExchangeFailed("").WithCause(fmt.Errorf("state not recognized"))
		}
		return identityStateValue{}, err
	}
	// A miss, a foreign provider, and an unparseable payload all answer
	// the same refusal.
	val, uerr := utils.Unmarshal[identityStateValue](stored)
	if uerr != nil || val.Provider != providerKey {
		return identityStateValue{}, systempb.ErrorSystemIdentityExchangeFailed("").WithCause(fmt.Errorf("state not recognized"))
	}
	return val, nil
}
