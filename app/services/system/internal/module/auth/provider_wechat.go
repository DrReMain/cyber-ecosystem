package auth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"cyber-ecosystem/shared-go/utils"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/conf"
)

const (
	wechatKind         = "wechat_mp"
	wechatEntryURL     = "https://open.weixin.qq.com/connect/oauth2/authorize"
	wechatExchangePath = "/sns/oauth2/access_token"
	wechatTimeout      = 5 * time.Second
)

type wechatProvider struct {
	client *http.Client
}

var wechatDefaultClient = &http.Client{Timeout: wechatTimeout}

func NewWechatProvider() wechatProvider { return wechatProvider{} }

func (p wechatProvider) clientOrDefault() *http.Client {
	if p.client != nil {
		return p.client
	}
	return wechatDefaultClient
}

type wechatTokenResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
	OpenID  string `json:"openid"`
}

func (wechatProvider) AuthorizeURL(row *conf.Auth_Provider, redirectURI, state string) string {
	q := url.Values{}
	q.Set("appid", row.GetWechat().GetAppid())
	q.Set("redirect_uri", redirectURI)
	q.Set("response_type", "code")
	q.Set("scope", "snsapi_base")
	q.Set("state", state)
	return wechatEntryURL + "?" + q.Encode() + "#wechat_redirect"
}

func (p wechatProvider) Exchange(ctx context.Context, row *conf.Auth_Provider, payload identityPayload) (*exchangedIdentity, error) {
	code := payload.GetWechat().GetCode()
	if code == "" {
		return nil, errorspb.ErrorGeneralErrorValidationFailed("").WithCause(fmt.Errorf("wechat payload required"))
	}
	base := row.GetEndpoint()
	if base == "" {
		base = "https://api.weixin.qq.com"
	}
	q := url.Values{}
	q.Set("appid", row.GetWechat().GetAppid())
	q.Set("secret", row.GetWechat().GetSecret())
	q.Set("code", code)
	q.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+wechatExchangePath+"?"+q.Encode(), nil)
	if err != nil {
		return nil, errorspb.ErrorGeneralErrorInternal("").WithCause(fmt.Errorf("wechat exchange: endpoint invalid"))
	}
	resp, err := p.clientOrDefault().Do(req)
	if err != nil {
		if inner := errors.Unwrap(err); inner != nil {
			err = inner
		}
		return nil, errorspb.ErrorGeneralErrorInternal("").WithCause(fmt.Errorf("wechat exchange: transport: %w", err))
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, errorspb.ErrorGeneralErrorInternal("").WithCause(fmt.Errorf("wechat exchange: status %d", resp.StatusCode))
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if err != nil {
		return nil, errorspb.ErrorGeneralErrorInternal("").WithCause(err)
	}
	out, err := utils.Unmarshal[wechatTokenResponse](body)
	if err != nil {
		return nil, errorspb.ErrorGeneralErrorInternal("").WithCause(err)
	}
	if out.ErrCode != 0 || out.OpenID == "" {
		return nil, systempb.ErrorSystemIdentityExchangeFailed("").WithCause(
			fmt.Errorf("wechat exchange: errcode %d %s", out.ErrCode, out.ErrMsg))
	}
	return &exchangedIdentity{Kind: wechatKind, Identifier: out.OpenID}, nil
}
