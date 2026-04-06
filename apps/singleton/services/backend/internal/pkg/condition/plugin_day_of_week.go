package condition

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type dayOfWeekConfig struct {
	Days []string `json:"days"`
}

// DayOfWeekPlugin checks if today is in the allowed days.
type DayOfWeekPlugin struct{}

func (p *DayOfWeekPlugin) Type() string { return "day_of_week" }

func (p *DayOfWeekPlugin) Evaluate(_ context.Context, config string) (bool, error) {
	cfg, err := p.parseConfig(config)
	if err != nil {
		return false, err
	}
	today := strings.ToLower(time.Now().Weekday().String())
	for _, d := range cfg.Days {
		if strings.EqualFold(d, today) {
			return true, nil
		}
	}
	return false, nil
}

func (p *DayOfWeekPlugin) ValidateConfig(config string) error {
	_, err := p.parseConfig(config)
	return err
}

func (p *DayOfWeekPlugin) parseConfig(raw string) (*dayOfWeekConfig, error) {
	if raw == "" {
		return nil, fmt.Errorf("day_of_week config is required")
	}
	var cfg dayOfWeekConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal day_of_week config: %w", err)
	}
	if len(cfg.Days) == 0 {
		return nil, fmt.Errorf("day_of_week config has no days")
	}
	return &cfg, nil
}
