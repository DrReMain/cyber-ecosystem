package auth

import (
	"context"

	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/conf"
)

// Identity exchange providers. The kind vocabulary is code — one file per
// kind (provider_<kind>.go), init-registered, duplicate kinds panic at
// boot; the file list IS the registry. Instances are conf rows: a project
// picks its methods by keeping provider files and conf rows, and deleting
// a kind removes its registration with the file.

// identityPayload is the request face's oneof member — every exchange
// request carries exactly one typed payload.
type identityPayload interface {
	GetWechat() *systempb.WeChatIdentityPayload
}

// exchangedIdentity is the verified external identity: what the kind
// exchanged for, globally unique within the kind.
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
