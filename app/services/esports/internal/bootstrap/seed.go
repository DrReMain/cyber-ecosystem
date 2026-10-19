package bootstrap

import (
	"context"
	"fmt"
	"log/slog"

	"cyber-ecosystem/app/services/esports/internal/ent/game"
	"cyber-ecosystem/app/services/esports/internal/ent/test"
	"cyber-ecosystem/app/services/esports/internal/platform"
)

var seedGames = []struct {
	Code        string
	Name        string
	Category    string
	Description string
}{
	{"lol", "英雄联盟", "moba", "MOBA竞技游戏"},
	{"dota2", "Dota 2", "moba", "MOBA竞技游戏"},
	{"honor", "王者荣耀", "moba", "MOBA竞技游戏"},
	{"csgo", "CS:GO", "fps", "第一人称射击游戏"},
	{"valorant", "Valorant", "fps", "第一人称射击游戏"},
	{"pubg", "PUBG", "fps", "第一人称射击游戏"},
	{"apex", "Apex Legends", "fps", "第一人称射击游戏"},
	{"overwatch", "守望先锋", "fps", "第一人称射击游戏"},
	{"fortnite", "Fortnite", "fps", "大逃杀射击游戏"},
}

var seedTests = []string{
	"red_dot",
	"flash_aiming",
	"multi_target",
	"target_recognition",
	"tracking_moving",
	"moving_target",
	"skill_trajectory",
	"dual_zone",
	"number_stars",
	"offense_defense",
	"character_counter",
	"resource_match",
	"economic_sim",
}

type Seed struct {
	log  *slog.Logger
	plat *platform.Platform
}

func NewSeed(p *platform.Platform, logger *slog.Logger) (*Seed, error) {
	return &Seed{log: logger.With("module", "bootstrap/seed"), plat: p}, nil
}

// Run ------------------------------------------------------------------------------------------------------------------

func (s *Seed) Run(ctx context.Context) error {
	client := s.plat.GetClient(ctx)
	for _, g := range seedGames {
		exists, err := client.Game.Query().Where(game.CodeEQ(g.Code)).Exist(ctx)
		if err != nil {
			return s.plat.HandleEntError(fmt.Errorf("seed: query game %q: %w", g.Code, err))
		}
		if exists {
			continue
		}
		if _, err := client.Game.Create().
			SetCode(g.Code).
			SetName(g.Name).
			SetCategory(g.Category).
			SetDescription(g.Description).
			Save(ctx); err != nil {
			return s.plat.HandleEntError(fmt.Errorf("seed: create game %q: %w", g.Code, err))
		}
		s.log.Info("seeded game", "code", g.Code)
	}
	for _, code := range seedTests {
		exists, err := client.Test.Query().Where(test.CodeEQ(code)).Exist(ctx)
		if err != nil {
			return s.plat.HandleEntError(fmt.Errorf("seed: query test %q: %w", code, err))
		}
		if exists {
			continue
		}
		row, err := client.Test.Create().
			SetCode(code).
			// Placeholder until the E3 metric sheet calibrates; the code is
			// the honest name while no calibrated one exists.
			SetName(code).
			Save(ctx)
		if err != nil {
			return s.plat.HandleEntError(fmt.Errorf("seed: create test %q: %w", code, err))
		}
		if _, err := client.TestConfig.Create().
			SetTestCode(code).
			Save(ctx); err != nil {
			return s.plat.HandleEntError(fmt.Errorf("seed: create test config %q: %w", code, err))
		}
		s.log.Info("seeded test", "code", code, "id", row.ID)
	}
	return nil
}
