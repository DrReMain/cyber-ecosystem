package condition

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
)

type ipRangeConfig struct {
	IPs  []string `json:"ips"`
	nets []*net.IPNet
}

// IPRangePlugin checks if the request IP is in the allowed list.
type IPRangePlugin struct{}

func (p *IPRangePlugin) Type() string { return "ip_range" }

func (p *IPRangePlugin) Evaluate(ctx context.Context, config string) (bool, error) {
	cfg, err := p.parseConfig(config)
	if err != nil {
		return false, err
	}
	clientIP := ClientIPFromContext(ctx)
	if clientIP == "" {
		return false, fmt.Errorf("missing client ip")
	}
	ip := net.ParseIP(clientIP)
	if ip == nil {
		return false, fmt.Errorf("invalid client ip: %q", clientIP)
	}
	for _, network := range cfg.nets {
		if network.Contains(ip) {
			return true, nil
		}
	}
	return false, nil
}

func (p *IPRangePlugin) ValidateConfig(config string) error {
	_, err := p.parseConfig(config)
	return err
}

func (p *IPRangePlugin) parseConfig(raw string) (*ipRangeConfig, error) {
	if raw == "" {
		return nil, fmt.Errorf("ip_range config is required")
	}
	var cfg ipRangeConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal ip_range config: %w", err)
	}
	if len(cfg.IPs) == 0 {
		return nil, fmt.Errorf("ip_range config has no ips")
	}
	nets := make([]*net.IPNet, 0, len(cfg.IPs))
	for _, cidr := range cfg.IPs {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, fmt.Errorf("invalid CIDR %q: %w", cidr, err)
		}
		nets = append(nets, network)
	}
	cfg.nets = nets
	return &cfg, nil
}
