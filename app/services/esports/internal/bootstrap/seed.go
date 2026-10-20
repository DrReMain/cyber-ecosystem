package bootstrap

import (
	"context"
	"fmt"
	"log/slog"

	"cyber-ecosystem/app/services/esports/internal/ent"
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

var seedTests = []struct {
	Code     string
	Name     string
	Category string
	Ability  string
}{
	{"red_dot", "随机红点触摸测试", "universal", "reactionSpeed"},
	{"number_stars", "数字星空", "universal", "reactionSpeed"},
	{"dual_zone", "双区同步躲避", "universal", "reactionSpeed"},
	{"target_recognition", "目标识别测试", "universal", "reactionSpeed"},
	{"flash_aiming", "闪现瞄准测试", "fps", "aimingAccuracy"},
	{"moving_target", "移动靶射击测试", "fps", "trackingStability"},
	{"tracking_moving", "追踪移动目标测试", "fps", "trackingStability"},
	{"skill_trajectory", "技能轨迹测试", "universal", "skillReleaseAccuracy"},
	{"offense_defense", "攻防决策测试", "moba", "decisionEfficiency"},
	{"resource_match", "资源配对测试", "moba", "resourceAllocation"},
	{"character_counter", "角色Counter测试", "moba", "resourceAllocation"},
	{"economic_sim", "经济模拟测试", "moba", "resourceAllocation"},
	{"multi_target", "多目标点击测试", "fps", "aimingAccuracy"},
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
		_, err := client.Game.Query().Where(game.CodeEQ(g.Code)).Only(ctx)
		if err == nil {
			continue
		}
		if !ent.IsNotFound(err) {
			return s.plat.HandleEntError(fmt.Errorf("seed: query game %q: %w", g.Code, err))
		}
		if _, err := client.Game.Create().
			SetCode(g.Code).
			SetName(g.Name).
			SetCategory(g.Category).
			SetDescription(g.Description).
			Save(ctx); err != nil {
			// Multi-replica boot race: a concurrent seed may have inserted
			// the unique row between the miss and this create — re-query
			// before failing (the same arm the test loop carries).
			if ent.IsConstraintError(err) {
				if _, reErr := client.Game.Query().Where(game.CodeEQ(g.Code)).Only(ctx); reErr == nil {
					continue
				}
			}
			return s.plat.HandleEntError(fmt.Errorf("seed: create game %q: %w", g.Code, err))
		}
		s.log.Info("seeded game", "code", g.Code)
	}
	for _, t := range seedTests {
		// Names and abilities are adopted verbatim from the legacy
		// interactive baseline (_state.js GAMES, an authority-table source —
		// multi_target included). Categories encode the prototype's pool law
		// (genre-order rule): lineups draw per-ability pools the genre
		// order runs, so families both orders draw (reactionSpeed,
		// skillReleaseAccuracy) are universal while single-genre families
		// keep their native mark. Durations, engine params, unit labels,
		// and introductions remain metric-sheet data.
		row, err := client.Test.Query().Where(test.CodeEQ(t.Code)).Only(ctx)
		if err != nil && !ent.IsNotFound(err) {
			return s.plat.HandleEntError(fmt.Errorf("seed: query test %q: %w", t.Code, err))
		}
		if row == nil {
			row, err = client.Test.Create().
				SetCode(t.Code).
				SetName(t.Name).
				SetNillableCategory(strPtr(t.Category)).
				SetNillableAbility(strPtr(t.Ability)).
				Save(ctx)
			if err != nil {
				// Multi-replica boot race: see the game loop's arm — the
				// winning replica created the config row too.
				if ent.IsConstraintError(err) {
					if _, reErr := client.Test.Query().Where(test.CodeEQ(t.Code)).Only(ctx); reErr == nil {
						continue
					}
				}
				return s.plat.HandleEntError(fmt.Errorf("seed: create test %q: %w", t.Code, err))
			}
			if _, err := client.TestConfig.Create().
				SetTestCode(t.Code).
				Save(ctx); err != nil {
				return s.plat.HandleEntError(fmt.Errorf("seed: create test config %q: %w", t.Code, err))
			}
			s.log.Info("seeded test", "code", t.Code, "id", row.ID)
			continue
		}
		// Upsert keeps already-seeded rows in step with adopted baseline
		// data; admin-tuned values will move out of this table's seed path
		// entirely.
		if _, err := row.Update().
			SetName(t.Name).
			SetNillableCategory(strPtr(t.Category)).
			SetNillableAbility(strPtr(t.Ability)).
			Save(ctx); err != nil {
			return s.plat.HandleEntError(fmt.Errorf("seed: update test %q: %w", t.Code, err))
		}
	}
	return nil
}

func strPtr(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}
