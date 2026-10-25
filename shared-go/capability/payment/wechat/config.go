package wechat

import (
	stderrors "errors"
	"fmt"

	"cyber-ecosystem/shared-go/capability/payment"
)

// The channel's primary API host; the documented fallback host stays a
// deploy-plane concern — the capability takes the base URL it is given.
const defaultBaseURL = "https://api.mch.weixin.qq.com"

// Config is everything the APIv3 backend needs at construction. The five
// identity fields are required in both verification modes; BaseURL defaults
// to the primary host. The wechatpay public-key pair picks the primary
// verification mode: configured, response and notification signatures
// verify against that fixed key under its PUB_KEY_ID (the only mode new
// merchants get); left empty, platform certificates bootstrap lazily from
// the channel's certificate endpoint on first use — both keys can coexist
// in the verifier table during a mode switch.
type Config struct {
	MchID     string
	MchSerial string
	// PrivateKeyPEM is the merchant API certificate's private key.
	PrivateKeyPEM []byte
	// APIv3Key is the 32-byte symmetric key that seals notification and
	// certificate payloads.
	APIv3Key []byte
	AppID    string
	BaseURL  string

	WechatpayPublicKeyID  string
	WechatpayPublicKeyPEM []byte
}

func (c Config) validate() error {
	switch {
	case c.MchID == "", c.MchSerial == "", c.AppID == "":
		return stderrors.New("MchID, MchSerial and AppID are required")
	case len(c.APIv3Key) != 32:
		return fmt.Errorf("APIv3Key must be exactly 32 bytes, got %d", len(c.APIv3Key))
	case (len(c.WechatpayPublicKeyID) == 0) != (len(c.WechatpayPublicKeyPEM) == 0):
		return stderrors.New("WechatpayPublicKeyID and WechatpayPublicKeyPEM come as a pair")
	default:
		return nil
	}
}

// New assembles the WeChat backend into the capability container: one
// transport core shared by the three faces, so gateway calls, notification
// verification and statement downloads trust one key table and one replay
// window.
func New(cfg Config) (*payment.Payment, error) {
	c, err := newClient(cfg)
	if err != nil {
		return nil, err
	}
	return &payment.Payment{
		Gateway:  &gateway{c: c, appID: cfg.AppID},
		Notifier: &notifier{c: c},
		Reports:  &reports{c: c},
	}, nil
}
