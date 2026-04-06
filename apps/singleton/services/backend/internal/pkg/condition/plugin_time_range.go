package condition

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type timeRangeConfig struct {
	Start        string `json:"start"`
	End          string `json:"end"`
	startMinutes int
	endMinutes   int
}

// TimeRangePlugin checks if current time is within [start, end] (HH:MM format).
type TimeRangePlugin struct{}

func (p *TimeRangePlugin) Type() string { return "time_range" }

func (p *TimeRangePlugin) Evaluate(_ context.Context, config string) (bool, error) {
	cfg, err := p.parseConfig(config)
	if err != nil {
		return false, err
	}
	now := time.Now()
	currentMinutes := now.Hour()*60 + now.Minute()
	return currentMinutes >= cfg.startMinutes && currentMinutes <= cfg.endMinutes, nil
}

func (p *TimeRangePlugin) ValidateConfig(config string) error {
	_, err := p.parseConfig(config)
	return err
}

func (p *TimeRangePlugin) parseConfig(raw string) (*timeRangeConfig, error) {
	if raw == "" {
		return nil, fmt.Errorf("time_range config is required")
	}
	var cfg timeRangeConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal time_range config: %w", err)
	}
	start, ok := parseHHMM(cfg.Start)
	if !ok {
		return nil, fmt.Errorf("invalid start time: %q", cfg.Start)
	}
	end, ok := parseHHMM(cfg.End)
	if !ok {
		return nil, fmt.Errorf("invalid end time: %q", cfg.End)
	}
	cfg.startMinutes = start
	cfg.endMinutes = end
	return &cfg, nil
}
