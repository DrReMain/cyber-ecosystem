package auth

import (
	"context"

	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/conf"
)

type identityPayload interface {
	GetWechat() *systempb.WeChatIdentityPayload
}

type exchangedIdentity struct {
	Kind       string
	Identifier string
}

type identityProvider interface {
	Exchange(ctx context.Context, row *conf.Auth_Provider, payload identityPayload) (*exchangedIdentity, error)
	AuthorizeURL(row *conf.Auth_Provider, redirectURI, state string) string
}

var identityProviders = map[string]identityProvider{}

func registerIdentityProvider(kind string, p identityProvider) {
	// A duplicate kind would silently override the first registration; a
	// boot-time panic turns that into a build mistake caught immediately.
	if _, dup := identityProviders[kind]; dup {
		panic("auth: duplicate identity provider kind: " + kind)
	}
	identityProviders[kind] = p
}
