package platform

import (
	"fmt"
	"strings"

	"cyber-ecosystem/shared-go/capability/payment"
	"cyber-ecosystem/shared-go/capability/payment/mock"
	"cyber-ecosystem/shared-go/capability/payment/wechat"

	"cyber-ecosystem/app/services/esports/internal/conf"
)

type Payments struct {
	defaultName string
	notifyBase  string
	backends    map[string]*payment.Payment
	channels    map[string][]payment.Channel
}

func NewPayments(c *conf.Payment) (*Payments, func(), error) {
	if c == nil || c.GetDefaultProvider() == "" || len(c.GetProviders()) == 0 {
		return nil, nil, fmt.Errorf("payment config is required: default_provider and at least one provider")
	}
	p := &Payments{
		// One backend per configured provider, coexisting — orders route
		// by their provider field, never by a singleton; construction is
		// total, a malformed block fails the boot.
		defaultName: c.GetDefaultProvider(),
		notifyBase:  strings.TrimRight(c.GetNotifyBaseUrl(), "/"),
		backends:    make(map[string]*payment.Payment, len(c.GetProviders())),
		channels:    make(map[string][]payment.Channel, len(c.GetProviders())),
	}
	for _, prov := range c.GetProviders() {
		if prov.GetName() == "" {
			return nil, nil, fmt.Errorf("payment provider: name is required")
		}
		if _, exists := p.backends[prov.GetName()]; exists {
			return nil, nil, fmt.Errorf("payment provider %q: configured twice", prov.GetName())
		}
		var (
			backend *payment.Payment
			channel []payment.Channel
			err     error
		)
		switch prov.GetKind() {
		case "mock":
			// Dev backend; its channel values mirror wechat's set, so a
			// consumer's choice never changes when the platform swaps it in.
			backend, channel = mock.New(), mock.Channels()
		case "wechat":
			// WeChat's servers call back into the notify base; without one
			// the orders would carry a callback URL no one can reach, so
			// the boot refuses rather than mint broken notify URLs.
			if p.notifyBase == "" {
				return nil, nil, fmt.Errorf("payment provider %q: wechat requires a non-empty notify_base_url", prov.GetName())
			}
			backend, err = wechat.New(toWechatConfig(prov.GetWechat()))
			channel = wechat.Channels()
		default:
			err = fmt.Errorf("payment provider %q: unknown kind %q", prov.GetName(), prov.GetKind())
		}
		if err != nil {
			return nil, nil, fmt.Errorf("payment provider %q: %w", prov.GetName(), err)
		}
		p.backends[prov.GetName()] = backend
		p.channels[prov.GetName()] = channel
	}
	if _, err := p.backend(p.defaultName); err != nil {
		return nil, nil, fmt.Errorf("payment config: default_provider: %w", err)
	}
	return p, func() {}, nil
}

func (p *Payments) backend(name string) (*payment.Payment, error) {
	if b, ok := p.backends[name]; ok {
		return b, nil
	}
	// A payment sentinel already, so callers fold it through the error
	// handler unchanged.
	return nil, fmt.Errorf("%w: provider %q is not configured", payment.ErrUnavailable, name)
}

func (p *Payments) DefaultProvider() string { return p.defaultName }

func (p *Payments) NotifyBaseURL() string { return p.notifyBase }

func (p *Payments) ChannelsFor(name string) ([]payment.Channel, error) {
	if _, err := p.backend(name); err != nil {
		return nil, err
	}
	return p.channels[name], nil
}

func toWechatConfig(w *conf.Payment_Wechat) wechat.Config {
	return wechat.Config{
		MchID:                 w.GetMchId(),
		MchSerial:             w.GetMchSerial(),
		AppID:                 w.GetAppId(),
		APIv3Key:              []byte(w.GetApiv3Key()),
		PrivateKeyPEM:         []byte(w.GetPrivateKeyPem()),
		WechatpayPublicKeyID:  w.GetPublicKeyId(),
		WechatpayPublicKeyPEM: []byte(w.GetPublicKeyPem()),
		BaseURL:               w.GetBaseUrl(),
	}
}
