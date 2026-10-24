package bootstrap

import (
	"context"
	"fmt"
	"log/slog"

	"cyber-ecosystem/app/services/esports/internal/ent"
	"cyber-ecosystem/app/services/esports/internal/ent/game"
	"cyber-ecosystem/app/services/esports/internal/ent/matchconfig"
	"cyber-ecosystem/app/services/esports/internal/ent/questionbank"
	"cyber-ecosystem/app/services/esports/internal/ent/test"
	"cyber-ecosystem/app/services/esports/internal/ent/testconfig"
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
	Code         string
	Name         string
	Category     string
	Ability      string
	EngineParams map[string]string
}{
	{"red_dot", "随机红点触摸测试", "universal", "reactionSpeed", map[string]string{
		"target_size_px":     "20",
		"target_lifetime_ms": "1500",
		"spawn_padding_px":   "50",
		"unit_score":         "10",
	}},
	{"number_stars", "数字星空", "universal", "reactionSpeed", nil},
	{"dual_zone", "双区同步躲避", "universal", "reactionSpeed", map[string]string{
		"shape_size_px":          "36",
		"bar_thickness_px":       "24",
		"bar_ratio":              "0.35",
		"obstacle_speed_px_s":    "120",
		"spawn_interval_min_ms":  "800",
		"spawn_interval_span_ms": "700",
		"base_score_time":        "30",
		"difficulty_weight":      "1.15",
		"unit_score":             "15",
	}},
	{"target_recognition", "目标识别测试", "universal", "reactionSpeed", map[string]string{
		"spawn_interval_ms":        "700",
		"spawn_padding_px":         "100",
		"click_detection_range_px": "25",
		"invalid_click_margin_px":  "2",
		"unit_score":               "20",
	}},
	{"flash_aiming", "闪现瞄准测试", "fps", "aimingAccuracy", map[string]string{
		"spawn_interval_ms":       "800",
		"target_size_min_px":      "30",
		"target_size_span_px":     "50",
		"target_lifetime_min_ms":  "500",
		"target_lifetime_span_ms": "1000",
		"center_hit_deviation_px": "20",
		"unit_score":              "10",
	}},
	{"moving_target", "移动靶射击测试", "fps", "trackingStability", map[string]string{
		"sample_interval_ms":   "50",
		"pattern_duration_ms":  "5000",
		"target_size_px":       "60",
		"margin_px":            "100",
		"perfect_deviation_px": "20",
		"good_deviation_px":    "30",
		"pattern_deviation_px": "80",
		"unit_score":           "10",
	}},
	{"tracking_moving", "追踪移动目标测试", "fps", "trackingStability", map[string]string{
		"hit_threshold_px":            "45",
		"dot_size_px":                 "40",
		"spawn_padding_px":            "40",
		"crosshair_speed_px_s":        "400",
		"movement_sample_interval_ms": "100",
		"respawn_delay_max_ms":        "1000",
		"unit_score":                  "10",
	}},
	{"skill_trajectory", "技能轨迹测试", "universal", "skillReleaseAccuracy", map[string]string{
		"total_rounds":      "5",
		"display_time_ms":   "3000",
		"target_size_px":    "60",
		"palette_size_px":   "40",
		"match_distance_px": "40",
		"unit_score":        "8",
	}},
	{"offense_defense", "攻防决策测试", "moba", "decisionEfficiency", map[string]string{
		"total_rounds":      "10",
		"time_per_round_ms": "2000",
		"feedback_ms":       "800",
		"unit_score":        "25",
	}},
	{"resource_match", "资源配对测试", "moba", "resourceAllocation", map[string]string{
		"minion_count":           "25",
		"monster_count":          "12",
		"dragon_count":           "5",
		"collision_cap":          "10",
		"player_speed_px_s":      "180",
		"enemy_speed_px_s":       "100",
		"player_size_px":         "40",
		"enemy_size_px":          "40",
		"resource_size_px":       "30",
		"collect_radius_px":      "25",
		"enemy_count":            "5",
		"on_field_count":         "6",
		"spawn_min_gap_px":       "100",
		"trajectory_interval_ms": "200",
		"unit_score":             "20",
	}},
	{"character_counter", "角色Counter测试", "moba", "resourceAllocation", nil},
	{"economic_sim", "经济模拟测试", "moba", "resourceAllocation", nil},
	{"multi_target", "多目标点击测试", "fps", "aimingAccuracy", map[string]string{
		"spawn_interval_min_ms":  "500",
		"spawn_interval_span_ms": "400",
		"target_lifetime_ms":     "2500",
		"max_on_screen":          "8",
		"tier_value_high":        "100",
		"tier_value_medium":      "50",
		"tier_value_low":         "25",
		"tier_value_trap":        "-50",
		"unit_score":             "10",
	}},
}

var seedMatchConfigs = []struct {
	Genre       string
	Ability     string
	RenderOrder int
}{
	// The genre capability orders (fps/moba, four abilities each); a
	// channel scenario axis reuses this table later with no schema change.
	{"fps", "reactionSpeed", 1},
	{"fps", "aimingAccuracy", 2},
	{"fps", "trackingStability", 3},
	{"fps", "skillReleaseAccuracy", 4},
	{"moba", "resourceAllocation", 1},
	{"moba", "decisionEfficiency", 2},
	{"moba", "reactionSpeed", 3},
	{"moba", "skillReleaseAccuracy", 4},
}

var seedQuestions = []questionSeed{
	{
		TestCode: "offense_defense",
		Scenario: "OD_001",
		Payload: map[string]any{
			"category":    "中期",
			"description": "敌方中单消失，我方下路正在推进",
			"choices": []map[string]string{
				{"choice": "attack", "text": "敌方中单可能来gank，不安全"},
				{"choice": "defense", "text": "被动防守无法获得优势"},
				{"choice": "retreat", "text": "✓ 安全撤退，避免被gank"},
			},
		},
		Answer: map[string]any{"complexity": "simple", "difficulty": 1, "optimal_choice": "retreat"},
	},
	{
		TestCode: "offense_defense",
		Scenario: "OD_002",
		Payload: map[string]any{
			"category":    "早期",
			"description": "我方打野在附近，敌人塔下残血",
			"choices": []map[string]string{
				{"choice": "attack", "text": "✓ 配合打野越塔击杀"},
				{"choice": "defense", "text": "错失击杀机会"},
				{"choice": "retreat", "text": "浪费打野gank机会"},
			},
		},
		Answer: map[string]any{"complexity": "medium", "difficulty": 2, "optimal_choice": "attack"},
	},
	{
		TestCode: "offense_defense",
		Scenario: "OD_003",
		Payload: map[string]any{
			"category":    "中期",
			"description": "敌方团队正在打龙，我方队伍分散",
			"choices": []map[string]string{
				{"choice": "attack", "text": "团队分散无法参战"},
				{"choice": "defense", "text": "✓ 防守等待时机，避免团灭"},
				{"choice": "retreat", "text": "放弃龙损失太大"},
			},
		},
		Answer: map[string]any{"complexity": "complex", "difficulty": 3, "optimal_choice": "defense"},
	},
	{
		TestCode: "offense_defense",
		Scenario: "OD_004",
		Payload: map[string]any{
			"category":    "早期",
			"description": "敌方打野上路gank，我方有视野且闪现未用",
			"choices": []map[string]string{
				{"choice": "attack", "text": "打野在gank无法进攻"},
				{"choice": "defense", "text": "被gank风险高"},
				{"choice": "retreat", "text": "✓ 使用视野优势及时撤退"},
			},
		},
		Answer: map[string]any{"complexity": "medium", "difficulty": 2, "optimal_choice": "retreat"},
	},
	{
		TestCode: "offense_defense",
		Scenario: "OD_005",
		Payload: map[string]any{
			"category":    "中期",
			"description": "我方在中路防御塔处5打3人数优势",
			"choices": []map[string]string{
				{"choice": "attack", "text": "✓ 人数优势推进获胜"},
				{"choice": "defense", "text": "浪费人数优势"},
				{"choice": "retreat", "text": "错失优势期"},
			},
		},
		Answer: map[string]any{"complexity": "simple", "difficulty": 1, "optimal_choice": "attack"},
	},
	{
		TestCode: "offense_defense",
		Scenario: "OD_006",
		Payload: map[string]any{
			"category":    "早期",
			"description": "敌方刺客正在游走支援你的线路，无视野",
			"choices": []map[string]string{
				{"choice": "attack", "text": "无视野容易被击杀"},
				{"choice": "defense", "text": "被动等待不利"},
				{"choice": "retreat", "text": "✓ 无视野情况下保守撤退"},
			},
		},
		Answer: map[string]any{"complexity": "medium", "difficulty": 2, "optimal_choice": "retreat"},
	},
	{
		TestCode: "offense_defense",
		Scenario: "OD_007",
		Payload: map[string]any{
			"category":    "后期",
			"description": "我方刚赢下团战，大龙刷新，敌方存活2人",
			"choices": []map[string]string{
				{"choice": "attack", "text": "人数劣势打龙风险大"},
				{"choice": "defense", "text": "✓ 守住优势防止翻盘"},
				{"choice": "retreat", "text": "浪费团战胜利"},
			},
		},
		Answer: map[string]any{"complexity": "complex", "difficulty": 3, "optimal_choice": "defense"},
	},
	{
		TestCode: "offense_defense",
		Scenario: "OD_008",
		Payload: map[string]any{
			"category":    "中期",
			"description": "敌方射手无闪现，我方辅助控制技能已好",
			"choices": []map[string]string{
				{"choice": "attack", "text": "✓ 抓住技能差击杀ADC"},
				{"choice": "defense", "text": "错失击杀机会"},
				{"choice": "retreat", "text": "浪费优势时机"},
			},
		},
		Answer: map[string]any{"complexity": "simple", "difficulty": 1, "optimal_choice": "attack"},
	},
	{
		TestCode: "offense_defense",
		Scenario: "OD_009",
		Payload: map[string]any{
			"category":    "后期",
			"description": "敌方团队正在推我方基地，我方可以偷家",
			"choices": []map[string]string{
				{"choice": "attack", "text": "偷家速度不够快"},
				{"choice": "defense", "text": "✓ 回防保护基地"},
				{"choice": "retreat", "text": "放弃基地必败"},
			},
		},
		Answer: map[string]any{"complexity": "complex", "difficulty": 3, "optimal_choice": "defense"},
	},
	{
		TestCode: "offense_defense",
		Scenario: "OD_010",
		Payload: map[string]any{
			"category":    "后期",
			"description": "后期，双方在大龙处对峙，我方团战能力更强",
			"choices": []map[string]string{
				{"choice": "attack", "text": "✓ 团战能力强主动开团"},
				{"choice": "defense", "text": "后期机会难得"},
				{"choice": "retreat", "text": "错失取胜机会"},
			},
		},
		Answer: map[string]any{"complexity": "complex", "difficulty": 3, "optimal_choice": "attack"},
	},
	{
		TestCode: "offense_defense",
		Scenario: "OD_011",
		Payload: map[string]any{
			"category":    "中期",
			"description": "我方上单正在单带，敌方正在包围他",
			"choices": []map[string]string{
				{"choice": "attack", "text": "队友被包围无法进攻"},
				{"choice": "defense", "text": "✓ 防守或支援上单"},
				{"choice": "retreat", "text": "放弃队友不利"},
			},
		},
		Answer: map[string]any{"complexity": "medium", "difficulty": 2, "optimal_choice": "defense"},
	},
	{
		TestCode: "offense_defense",
		Scenario: "OD_012",
		Payload: map[string]any{
			"category":    "早期",
			"description": "敌方防御塔血量很低，但敌方即将复活",
			"choices": []map[string]string{
				{"choice": "attack", "text": "敌方复活后危险"},
				{"choice": "defense", "text": "塔已残血不需防守"},
				{"choice": "retreat", "text": "✓ 撤退避免被反杀"},
			},
		},
		Answer: map[string]any{"complexity": "medium", "difficulty": 2, "optimal_choice": "retreat"},
	},
	{
		TestCode: "offense_defense",
		Scenario: "OD_013",
		Payload: map[string]any{
			"category":    "后期",
			"description": "我方拥有远古巨龙buff，敌方防守",
			"choices": []map[string]string{
				{"choice": "attack", "text": "✓ 远古龙buff碾压性推进"},
				{"choice": "defense", "text": "浪费龙buff时间"},
				{"choice": "retreat", "text": "放弃绝佳机会"},
			},
		},
		Answer: map[string]any{"complexity": "simple", "difficulty": 1, "optimal_choice": "attack"},
	},
	{
		TestCode: "offense_defense",
		Scenario: "OD_014",
		Payload: map[string]any{
			"category":    "中期",
			"description": "敌方优势核心位置未知，我方想要强开团战",
			"choices": []map[string]string{
				{"choice": "attack", "text": "核心位置未知风险大"},
				{"choice": "defense", "text": "✓ 谨慎等待敌方暴露位置"},
				{"choice": "retreat", "text": "过于保守"},
			},
		},
		Answer: map[string]any{"complexity": "complex", "difficulty": 3, "optimal_choice": "defense"},
	},
	{
		TestCode: "offense_defense",
		Scenario: "OD_015",
		Payload: map[string]any{
			"category":    "早期",
			"description": "在我方野区抓到敌方打野，队友在附近",
			"choices": []map[string]string{
				{"choice": "attack", "text": "✓ 人多击杀落单打野"},
				{"choice": "defense", "text": "错失击杀机会"},
				{"choice": "retreat", "text": "浪费优势"},
			},
		},
		Answer: map[string]any{"complexity": "simple", "difficulty": 1, "optimal_choice": "attack"},
	},
	{
		TestCode: "offense_defense",
		Scenario: "OD_016",
		Payload: map[string]any{
			"category":    "中期",
			"description": "敌方有你的视野，他们正在设置埋伏",
			"choices": []map[string]string{
				{"choice": "attack", "text": "已被发现会中埋伏"},
				{"choice": "defense", "text": "被动等待不利"},
				{"choice": "retreat", "text": "✓ 撤离危险区域"},
			},
		},
		Answer: map[string]any{"complexity": "medium", "difficulty": 2, "optimal_choice": "retreat"},
	},
	{
		TestCode: "offense_defense",
		Scenario: "OD_017",
		Payload: map[string]any{
			"category":    "后期",
			"description": "我方高地水晶已破，超级兵正在推进",
			"choices": []map[string]string{
				{"choice": "attack", "text": "兵线压力无法进攻"},
				{"choice": "defense", "text": "✓ 清理超级兵线"},
				{"choice": "retreat", "text": "放弃防守必败"},
			},
		},
		Answer: map[string]any{"complexity": "simple", "difficulty": 1, "optimal_choice": "defense"},
	},
	{
		TestCode: "offense_defense",
		Scenario: "OD_018",
		Payload: map[string]any{
			"category":    "中期",
			"description": "敌方交出关键大招，我方技能冷却完毕",
			"choices": []map[string]string{
				{"choice": "attack", "text": "✓ 抓住技能差发起团战"},
				{"choice": "defense", "text": "错失优势窗口"},
				{"choice": "retreat", "text": "浪费机会"},
			},
		},
		Answer: map[string]any{"complexity": "medium", "difficulty": 2, "optimal_choice": "attack"},
	},
	{
		TestCode: "offense_defense",
		Scenario: "OD_019",
		Payload: map[string]any{
			"category":    "早期",
			"description": "前期阶段，敌方打野等级领先",
			"choices": []map[string]string{
				{"choice": "attack", "text": "等级劣势会被反杀"},
				{"choice": "defense", "text": "前期不需过度防守"},
				{"choice": "retreat", "text": "✓ 前期发育避免战斗"},
			},
		},
		Answer: map[string]any{"complexity": "simple", "difficulty": 1, "optimal_choice": "retreat"},
	},
	{
		TestCode: "offense_defense",
		Scenario: "OD_020",
		Payload: map[string]any{
			"category":    "中期",
			"description": "我方实力更强但打完目标后血量较低",
			"choices": []map[string]string{
				{"choice": "attack", "text": "血量低容易被反杀"},
				{"choice": "defense", "text": "已经拿到目标"},
				{"choice": "retreat", "text": "✓ 安全撤退巩固优势"},
			},
		},
		Answer: map[string]any{"complexity": "complex", "difficulty": 3, "optimal_choice": "retreat"},
	},
	{
		TestCode: "character_counter",
		Scenario: "Tank|Fighter",
		Payload:  map[string]any{"counter": "Tank", "countered": "Fighter"},
		Answer:   map[string]any{"counter_score": 20},
	},
	{
		TestCode: "character_counter",
		Scenario: "Tank|Assassin",
		Payload:  map[string]any{"counter": "Tank", "countered": "Assassin"},
		Answer:   map[string]any{"counter_score": 20},
	},
	{
		TestCode: "character_counter",
		Scenario: "Fighter|Marksman",
		Payload:  map[string]any{"counter": "Fighter", "countered": "Marksman"},
		Answer:   map[string]any{"counter_score": 20},
	},
	{
		TestCode: "character_counter",
		Scenario: "Fighter|Mage",
		Payload:  map[string]any{"counter": "Fighter", "countered": "Mage"},
		Answer:   map[string]any{"counter_score": 20},
	},
	{
		TestCode: "character_counter",
		Scenario: "Mage|Tank",
		Payload:  map[string]any{"counter": "Mage", "countered": "Tank"},
		Answer:   map[string]any{"counter_score": 20},
	},
	{
		TestCode: "character_counter",
		Scenario: "Mage|Marksman",
		Payload:  map[string]any{"counter": "Mage", "countered": "Marksman"},
		Answer:   map[string]any{"counter_score": 20},
	},
	{
		TestCode: "character_counter",
		Scenario: "Marksman|Fighter",
		Payload:  map[string]any{"counter": "Marksman", "countered": "Fighter"},
		Answer:   map[string]any{"counter_score": 20},
	},
	{
		TestCode: "character_counter",
		Scenario: "Marksman|Support",
		Payload:  map[string]any{"counter": "Marksman", "countered": "Support"},
		Answer:   map[string]any{"counter_score": 20},
	},
	{
		TestCode: "character_counter",
		Scenario: "Support|Assassin",
		Payload:  map[string]any{"counter": "Support", "countered": "Assassin"},
		Answer:   map[string]any{"counter_score": 20},
	},
	{
		TestCode: "character_counter",
		Scenario: "Support|Mage",
		Payload:  map[string]any{"counter": "Support", "countered": "Mage"},
		Answer:   map[string]any{"counter_score": 20},
	},
	{
		TestCode: "character_counter",
		Scenario: "Assassin|Marksman",
		Payload:  map[string]any{"counter": "Assassin", "countered": "Marksman"},
		Answer:   map[string]any{"counter_score": 20},
	},
	{
		TestCode: "character_counter",
		Scenario: "Assassin|Mage",
		Payload:  map[string]any{"counter": "Assassin", "countered": "Mage"},
		Answer:   map[string]any{"counter_score": 20},
	},
}

type questionSeed struct {
	TestCode string
	Scenario string
	Payload  map[string]any
	Answer   map[string]any
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
		// keep their native mark. Durations, unit labels, and introductions
		// remain metric-sheet data; engine params materialize for the
		// engine-backed games as the same KV the client boot pack and the
		// server scorers read, values equal to the compiled metric-sheet
		// constants so the config path goes live without moving a needle.
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
			create := client.TestConfig.Create().SetTestCode(t.Code)
			if t.EngineParams != nil {
				create = create.SetEngineParams(t.EngineParams)
			}
			if _, err := create.Save(ctx); err != nil {
				return s.plat.HandleEntError(fmt.Errorf("seed: create test config %q: %w", t.Code, err))
			}
			s.log.Info("seeded test", "code", t.Code, "id", row.ID)
			continue
		}
		// Upsert keeps already-seeded rows in step with adopted baseline
		// data; admin-tuned values will move out of this table's seed path
		// entirely. The same law holds for engine params: a nil entry hands
		// the row's KV off untouched, so only the games whose params this
		// table owns get refreshed at boot.
		if _, err := row.Update().
			SetName(t.Name).
			SetNillableCategory(strPtr(t.Category)).
			SetNillableAbility(strPtr(t.Ability)).
			Save(ctx); err != nil {
			return s.plat.HandleEntError(fmt.Errorf("seed: update test %q: %w", t.Code, err))
		}
		if t.EngineParams == nil {
			continue
		}
		cfg, cfgErr := client.TestConfig.Query().Where(testconfig.TestCodeEQ(t.Code)).Only(ctx)
		if cfgErr != nil && !ent.IsNotFound(cfgErr) {
			return s.plat.HandleEntError(fmt.Errorf("seed: query test config %q: %w", t.Code, cfgErr))
		}
		if ent.IsNotFound(cfgErr) {
			if _, err := client.TestConfig.Create().
				SetTestCode(t.Code).
				SetEngineParams(t.EngineParams).
				Save(ctx); err != nil {
				return s.plat.HandleEntError(fmt.Errorf("seed: create test config %q: %w", t.Code, err))
			}
			continue
		}
		if _, err := cfg.Update().SetEngineParams(t.EngineParams).Save(ctx); err != nil {
			return s.plat.HandleEntError(fmt.Errorf("seed: update test config %q: %w", t.Code, err))
		}
	}

	for _, mc := range seedMatchConfigs {
		// Same keep-in-step law as the tests: create when absent, refresh
		// render_order when the adopted baseline moves.
		row, err := client.MatchConfig.Query().
			Where(matchconfig.GenreEQ(mc.Genre), matchconfig.AbilityEQ(mc.Ability)).
			Only(ctx)
		if err != nil && !ent.IsNotFound(err) {
			return s.plat.HandleEntError(fmt.Errorf("seed: query match config %s/%s: %w", mc.Genre, mc.Ability, err))
		}
		if row == nil {
			if _, err := client.MatchConfig.Create().
				SetGenre(mc.Genre).
				SetAbility(mc.Ability).
				SetRenderOrder(mc.RenderOrder).
				Save(ctx); err != nil {
				// Multi-replica boot race: see the game loop's arm — the
				// winning replica created the row too.
				if ent.IsConstraintError(err) {
					if _, reErr := client.MatchConfig.Query().
						Where(matchconfig.GenreEQ(mc.Genre), matchconfig.AbilityEQ(mc.Ability)).
						Only(ctx); reErr == nil {
						continue
					}
				}
				return s.plat.HandleEntError(fmt.Errorf("seed: create match config %s/%s: %w", mc.Genre, mc.Ability, err))
			}
			s.log.Info("seeded match config", "genre", mc.Genre, "ability", mc.Ability)
			continue
		}
		if _, err := row.Update().
			SetRenderOrder(mc.RenderOrder).
			Save(ctx); err != nil {
			return s.plat.HandleEntError(fmt.Errorf("seed: update match config %s/%s: %w", mc.Genre, mc.Ability, err))
		}
	}

	for _, q := range seedQuestions {
		// Bank rows follow the tests' keep-in-step law: create when absent,
		// refresh content when the adopted baseline moves (admin-tuned rows
		// will leave the seed path entirely, same as test seeds). Content is
		// the legacy init.sql bank verbatim (experience data) — the legacy
		// offense_scenario + offense_choice_explanation tables consolidate
		// into payload (rendering face) and answer (judging key).
		row, err := client.QuestionBank.Query().
			Where(questionbank.TestCodeEQ(q.TestCode), questionbank.ScenarioCodeEQ(q.Scenario)).
			Only(ctx)
		if err != nil && !ent.IsNotFound(err) {
			return s.plat.HandleEntError(fmt.Errorf("seed: query question %s/%s: %w", q.TestCode, q.Scenario, err))
		}
		if row == nil {
			if _, err := client.QuestionBank.Create().
				SetTestCode(q.TestCode).
				SetScenarioCode(q.Scenario).
				SetPayload(q.Payload).
				SetAnswer(q.Answer).
				Save(ctx); err != nil {
				// Multi-replica boot race: see the game loop's arm.
				if ent.IsConstraintError(err) {
					if _, reErr := client.QuestionBank.Query().
						Where(questionbank.TestCodeEQ(q.TestCode), questionbank.ScenarioCodeEQ(q.Scenario)).
						Only(ctx); reErr == nil {
						continue
					}
				}
				return s.plat.HandleEntError(fmt.Errorf("seed: create question %s/%s: %w", q.TestCode, q.Scenario, err))
			}
			s.log.Info("seeded question", "test", q.TestCode, "scenario", q.Scenario)
			continue
		}
		if _, err := row.Update().
			SetPayload(q.Payload).
			SetAnswer(q.Answer).
			Save(ctx); err != nil {
			return s.plat.HandleEntError(fmt.Errorf("seed: update question %s/%s: %w", q.TestCode, q.Scenario, err))
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
