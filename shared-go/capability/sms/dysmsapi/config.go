package dysmsapi

import (
	"fmt"
	"net/http"
	"time"
)

// Config carries the channel credentials and endpoint; per-send business
// values (sign, template, params) ride the Message — one credential set
// serves every template the caller owns.
type Config struct {
	AccessKeyID     string
	AccessKeySecret string
	// Endpoint defaults to https://dysmsapi.aliyuncs.com (tests override
	// here with an httptest URL).
	Endpoint string
	// RegionID defaults to cn-hangzhou.
	RegionID string
	// HTTPClient defaults to a client with a 5s timeout.
	HTTPClient *http.Client
}

const (
	defaultEndpoint = "https://dysmsapi.aliyuncs.com"
	defaultRegion   = "cn-hangzhou"
	defaultTimeout  = 5 * time.Second
)

// New validates the credential pair — a half-configured channel fails the
// boot, not the first send.
func New(cfg Config) (sender *Sender, err error) {
	if cfg.AccessKeyID == "" || cfg.AccessKeySecret == "" {
		return nil, fmt.Errorf("dysmsapi: access key id and secret are both required")
	}
	if cfg.Endpoint == "" {
		cfg.Endpoint = defaultEndpoint
	}
	if cfg.RegionID == "" {
		cfg.RegionID = defaultRegion
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: defaultTimeout}
	}
	return &Sender{cfg: cfg}, nil
}
