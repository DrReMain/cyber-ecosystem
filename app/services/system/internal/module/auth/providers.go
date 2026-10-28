package auth

import (
	"context"

	"cyber-ecosystem/shared-go/utils"

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
}

type IdentityProvider interface {
	Exchange(ctx context.Context, row *conf.Auth_Provider, payload identityPayload) (*exchangedIdentity, error)
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

func mintIdentityState(ctx context.Context, trp TokenRP, providerKey string) (string, error) {
	// The one state-minting path every provider shares: Begin mints it
	// for redirect flows, IssueCode for direct-code ones.
	state, err := utils.RandomToken()
	if err != nil {
		return "", err
	}
	if err := trp.Set(ctx, identityStatePrefix+state, []byte(providerKey), identityStateTTL); err != nil {
		return "", err
	}
	return state, nil
}
