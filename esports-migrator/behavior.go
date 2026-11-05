package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rs/xid"
)

// The behavior cut carries the member-plane history: game preferences into
// the profile's games array, and every legacy test attempt into the fact
// stream. The store's philosophy is "facts only, every aggregate is
// recomputed" — so the transport preserves raw payloads verbatim and lets
// the recompute phase re-derive scores' companions. Two provenance tables
// stay out by evidence: login_record has no home in the new schema (P22
// ruled the collection face; the archived old stack is the record), and
// user_referral waits on P23's re-ruling.
const (
	entityProfileGames = "profile_games"
	entityTestRecord   = "test_record"

	// The retired test code: the D-tier ruling voids skill_prediction (dead
	// endpoint, no valid stock) — its rows are dropped and counted, never
	// carried into the new fact stream.
	retiredTestCode = "skill_prediction"
)

// gameEntry mirrors the profile row's games jsonb shape (game_id for
// catalog picks, custom_name for the unclassified ones — the ruling keeps
// them unclassified). The json tags are the wire contract.
type gameEntry struct {
	GameID     string `json:"game_id,omitempty"`
	CustomName string `json:"custom_name,omitempty"`
}

// hasFacts reports whether a legacy JSON payload carries a fact stream:
// empty objects and nulls are the legacy writers' "nothing here" — a
// summary-only row whose detail is '{}' must fall through to its summary,
// not crown the empty object as the facts and drop the real ones.
func hasFacts(payload []byte) bool {
	t := strings.TrimSpace(string(payload))
	return t != "" && t != "{}" && t != "null"
}

// bandFor mirrors the live scoring path's tier thresholds verbatim
// (module/test in the esports service): the score reads as percentile
// position, the tiers are the frozen top 5/15/40. Migrated rows carry the
// derived grade so the new invariant — grade is a function of the
// authoritative score — holds from row one; behavior_band_test.go pins the
// boundaries.
func bandFor(score float64) string {
	switch {
	case score >= 95.0:
		return "S"
	case score >= 85.0:
		return "A"
	case score >= 60.0:
		return "B"
	default:
		return "C"
	}
}

type srcPreference struct {
	ID            int64
	UserID        int64
	GameID        sql.NullInt64
	GameName      string
	PlayFrequency sql.NullString
	ServerRank    sql.NullString
	IsDeleted     bool
}

type srcTestRecord struct {
	ID           int64
	UserID       int64
	TestCode     string
	Score        float64
	RawScore     sql.NullFloat64
	SummaryData  []byte
	DetailData   []byte
	StartTime    int64
	EndTime      int64
	CompletedAt  sql.NullTime
	UserUploaded bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	IsDeleted    bool
}

func fetchPreferences(ctx context.Context, db *sql.DB) ([]*srcPreference, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, user_id, game_id, game_name,
		play_frequency, server_rank, is_deleted FROM user_game_preference ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("query legacy preferences: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []*srcPreference
	for rows.Next() {
		p := &srcPreference{}
		var isDeleted int64
		err := rows.Scan(&p.ID, &p.UserID, &p.GameID, &p.GameName, &p.PlayFrequency, &p.ServerRank, &isDeleted)
		if err != nil {
			return nil, fmt.Errorf("scan legacy preference %d: %w", p.ID, err)
		}
		p.IsDeleted = isDeleted != 0
		out = append(out, p)
	}
	return out, rows.Err()
}

func fetchTestRecords(ctx context.Context, db *sql.DB) ([]*srcTestRecord, error) {
	// PK order only: a server-side sort over these wide detail rows blew the
	// sort buffer on the production-shaped copy (MySQL 1038), and the cutover
	// host's buffer size is not ours to assume. The four-key order is rebuilt
	// below.
	rows, err := db.QueryContext(ctx, `SELECT id, user_id, test_code, score, raw_score,
		summary_data, detail_data, start_time, end_time, completed_at, user_uploaded,
		created_at, updated_at, is_deleted FROM test_record ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("query legacy test_record: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []*srcTestRecord
	for rows.Next() {
		r := &srcTestRecord{}
		var userUploaded, isDeleted int64
		err := rows.Scan(&r.ID, &r.UserID, &r.TestCode, &r.Score, &r.RawScore,
			&r.SummaryData, &r.DetailData, &r.StartTime, &r.EndTime, &r.CompletedAt,
			&userUploaded, &r.CreatedAt, &r.UpdatedAt, &isDeleted)
		if err != nil {
			return nil, fmt.Errorf("scan legacy test_record %d: %w", r.ID, err)
		}
		r.UserUploaded = userUploaded != 0
		r.IsDeleted = isDeleted != 0
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// Same order the SQL used to carry: user, test, completion time (NULLs
	// first, MySQL ascending), id last so the order is total.
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		switch {
		case a.UserID != b.UserID:
			return a.UserID < b.UserID
		case a.TestCode != b.TestCode:
			return a.TestCode < b.TestCode
		case a.CompletedAt.Valid != b.CompletedAt.Valid:
			return !a.CompletedAt.Valid
		case !a.CompletedAt.Time.Equal(b.CompletedAt.Time):
			return a.CompletedAt.Time.Before(b.CompletedAt.Time)
		default:
			return a.ID < b.ID
		}
	})
	return out, nil
}

func countRows(ctx context.Context, db *sql.DB, table string) (int64, error) {
	var n int64
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM `"+table+"`").Scan(&n); err != nil {
		return 0, fmt.Errorf("count %s: %w", table, err)
	}
	return n, nil
}

type gamesPlan struct {
	userID  int64
	entries []gameEntry
}

type recordPlan struct {
	src       *srcTestRecord
	userID    string
	attemptNo int64
	rawDetail []byte
}

type behaviorCensus struct {
	prefTotal       int
	prefDead        int
	prefCatalog     int
	prefCustom      int
	prefDeduped     int
	prefFreqDropped int
	gamesUsers      int

	recTotal          int
	recDeadCarried    int
	recRetired        int
	recDetailFacts    int
	recSummaryFacts   int
	recSummaryDropped int
	recNoFacts        int
	recSelfReported   int
	recUsers          int

	loginRows  int64
	referrRows int64
}

func runBehavior(ctx context.Context, env *environment) error {
	prefs, err := fetchPreferences(ctx, env.src)
	if err != nil {
		return err
	}
	records, err := fetchTestRecords(ctx, env.src)
	if err != nil {
		return err
	}
	loginRows, err := countRows(ctx, env.src, "login_record")
	if err != nil {
		return err
	}
	referralRows, err := countRows(ctx, env.src, "user_referral")
	if err != nil {
		return err
	}

	c := &behaviorCensus{loginRows: loginRows, referrRows: referralRows}
	games, recPlans := buildBehaviorPlans(prefs, records, c)
	printBehaviorCensus(c, len(recPlans))

	if env.dryRun {
		fmt.Println("dry-run: no writes performed")
		return nil
	}
	if err := env.ensureEsports(); err != nil {
		return err
	}
	if err := ensureMapTable(ctx, env.esp); err != nil {
		return err
	}
	mapped, err := loadMappedUsers(ctx, env.esp)
	if err != nil {
		return err
	}

	gamesIns, gamesSkip := 0, 0
	for _, g := range games {
		done, err := applyProfileGames(ctx, env.esp, g)
		if err != nil {
			return err
		}
		if done {
			gamesIns++
		} else {
			gamesSkip++
		}
	}

	recIns, recSkip := 0, 0
	for _, rp := range recPlans {
		if _, ok := mapped[rp.userID]; !ok {
			return fmt.Errorf("test_record %d references user %s absent from migration_map — the identity phase has not covered this store, halting", rp.src.ID, rp.userID)
		}
		done, err := applyTestRecord(ctx, env.esp, rp)
		if err != nil {
			return err
		}
		if done {
			recIns++
		} else {
			recSkip++
		}
	}

	fmt.Printf("profile games: %d written / %d already-mapped\n", gamesIns, gamesSkip)
	fmt.Printf("test records: %d written / %d already-mapped\n", recIns, recSkip)
	return nil
}

func buildBehaviorPlans(prefs []*srcPreference, records []*srcTestRecord, c *behaviorCensus) ([]*gamesPlan, []*recordPlan) {
	// Preferences fold into one games array per user, row order preserved,
	// identical identities deduped (a catalog id may not repeat; custom
	// names compare case-insensitively). Frequency and rank have no column
	// in the new shape — dropped and counted.
	byUser := map[int64]*gamesPlan{}
	for _, p := range prefs {
		c.prefTotal++
		if p.IsDeleted {
			c.prefDead++
			continue
		}
		if p.PlayFrequency.Valid || p.ServerRank.Valid {
			c.prefFreqDropped++
		}
		g := byUser[p.UserID]
		if g == nil {
			g = &gamesPlan{userID: p.UserID}
			byUser[p.UserID] = g
		}
		if p.GameID.Valid && p.GameID.Int64 != 0 {
			g.entries = append(g.entries, gameEntry{GameID: strconv.FormatInt(p.GameID.Int64, 10)})
			c.prefCatalog++
		} else {
			g.entries = append(g.entries, gameEntry{CustomName: strings.TrimSpace(p.GameName)})
			c.prefCustom++
		}
	}
	games := make([]*gamesPlan, 0, len(byUser))
	for _, g := range byUser {
		seenID := map[string]struct{}{}
		seenName := map[string]struct{}{}
		deduped := make([]gameEntry, 0, len(g.entries))
		for _, e := range g.entries {
			switch {
			case e.GameID != "":
				if _, dup := seenID[e.GameID]; dup {
					c.prefDeduped++
					continue
				}
				seenID[e.GameID] = struct{}{}
			case e.CustomName != "":
				key := strings.ToLower(e.CustomName)
				if _, dup := seenName[key]; dup {
					c.prefDeduped++
					continue
				}
				seenName[key] = struct{}{}
			default:
				continue // an empty custom name carries no identity; not even the map hears about it
			}
			deduped = append(deduped, e)
		}
		g.entries = deduped
		games = append(games, g)
	}
	sort.Slice(games, func(i, j int) bool { return games[i].userID < games[j].userID })
	c.gamesUsers = len(games)

	// Records: attempt ordinals run per (user, test) in completion order —
	// the legacy store kept no ordinals, and the claim index only bites on
	// session attempts, so the derived numbering is presentational. The
	// fact stream is the detail payload when the legacy row kept one, else
	// the summary aggregate (the B-tier tests' only surviving facts).
	ordinal := map[string]int64{}
	recPlans := make([]*recordPlan, 0, len(records))
	users := map[int64]struct{}{}
	for _, r := range records {
		c.recTotal++
		if r.TestCode == retiredTestCode {
			c.recRetired++
			continue
		}
		if r.IsDeleted {
			c.recDeadCarried++
		}
		if r.UserUploaded {
			c.recSelfReported++
		}
		key := strconv.FormatInt(r.UserID, 10) + "\x00" + r.TestCode
		ordinal[key]++
		rp := &recordPlan{src: r, userID: strconv.FormatInt(r.UserID, 10), attemptNo: ordinal[key]}
		switch {
		case hasFacts(r.DetailData):
			rp.rawDetail = r.DetailData
			c.recDetailFacts++
			if hasFacts(r.SummaryData) {
				c.recSummaryDropped++
			}
		case hasFacts(r.SummaryData):
			rp.rawDetail = r.SummaryData
			c.recSummaryFacts++
		default:
			rp.rawDetail = []byte("{}")
			c.recNoFacts++
		}
		users[r.UserID] = struct{}{}
		recPlans = append(recPlans, rp)
	}
	c.recUsers = len(users)
	return games, recPlans
}

func printBehaviorCensus(c *behaviorCensus, plannedRecords int) {
	fmt.Printf("preferences: %d rows (%d dead dropped) → %d users; catalog %d / custom %d; duplicates folded %d; frequency+rank dropped on %d rows\n",
		c.prefTotal, c.prefDead, c.gamesUsers, c.prefCatalog, c.prefCustom, c.prefDeduped, c.prefFreqDropped)
	fmt.Printf("test records: %d rows → %d migrate (%d users); soft-deleted carried %d; retired %s dropped %d\n",
		c.recTotal, plannedRecords, c.recUsers, c.recDeadCarried, retiredTestCode, c.recRetired)
	fmt.Printf("  fact stream: detail %d / summary-as-facts %d (summary beside detail dropped %d) / no-facts {} %d; self-reported flagged %d\n",
		c.recDetailFacts, c.recSummaryFacts, c.recSummaryDropped, c.recNoFacts, c.recSelfReported)
	fmt.Printf("  provenance: login_record %d rows stay cold (no new-schema home; archived old stack is the record); user_referral %d rows wait on P23\n",
		c.loginRows, c.referrRows)
}

// applyProfileGames fills the profile row's games array once — the map row
// is the claim, and a games value already set outside the map (a real edit)
// is adopted, never overwritten.
func applyProfileGames(ctx context.Context, db *sql.DB, g *gamesPlan) (bool, error) {
	userID := strconv.FormatInt(g.userID, 10)
	var exists bool
	err := db.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM migration_map WHERE entity = $1 AND old_pk = $2)`,
		entityProfileGames, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("query map for games %s: %w", userID, err)
	}
	if exists {
		return false, nil
	}
	payload, err := json.Marshal(g.entries)
	if err != nil {
		return false, fmt.Errorf("marshal games for %s: %w", userID, err)
	}

	err = withTx(ctx, db, func(tx *sql.Tx) error {
		// updated_at deliberately stays what identity carried: timestamps
		// travel verbatim as historical truth, and the fold's own moment
		// is recorded in migration_map.created_at, not on the row.
		res, err := tx.ExecContext(ctx, `
			UPDATE profile SET games = $1
			WHERE user_id = $2 AND games IS NULL`,
			payload, userID)
		if err != nil {
			return fmt.Errorf("update profile games for %s: %w", userID, err)
		}
		if n, _ := res.RowsAffected(); n == 0 {
			var hasProfile bool
			if err := tx.QueryRowContext(ctx,
				`SELECT EXISTS (SELECT 1 FROM profile WHERE user_id = $1)`, userID).Scan(&hasProfile); err != nil {
				return fmt.Errorf("re-check profile for %s: %w", userID, err)
			}
			if !hasProfile {
				return fmt.Errorf("profile for user %s missing — the identity phase has not covered this store, halting", userID)
			}
			// games already set outside the map: a real edit stands, the
			// map just starts claiming the user's preference migration.
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO migration_map (id, created_at, updated_at, old_pk, entity, new_id)
			VALUES ($1, now(), now(), $2, $3, $4)
			ON CONFLICT (entity, old_pk) DO NOTHING`,
			xid.New().String(), userID, entityProfileGames, userID); err != nil {
			return fmt.Errorf("insert map row for games %s: %w", userID, err)
		}
		return nil
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

// applyTestRecord carries one attempt verbatim: the snowflake id doubles as
// the new row id, wall clocks convert from the millisecond epochs, and the
// grade re-derives from the carried score so the store's invariant holds
// from row one. Column list pinned to the esports service's Atlas schema
// for "test_record"; a schema change there must re-run the rehearsal's
// sampling pass.
func applyTestRecord(ctx context.Context, db *sql.DB, rp *recordPlan) (bool, error) {
	oldPK := strconv.FormatInt(rp.src.ID, 10)
	var exists bool
	err := db.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM migration_map WHERE entity = $1 AND old_pk = $2)`,
		entityTestRecord, oldPK).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("query map for record %s: %w", oldPK, err)
	}
	if exists {
		return false, nil
	}

	var deletedAt any
	if rp.src.IsDeleted {
		deletedAt = rp.src.UpdatedAt.UTC()
	}
	createdAt := rp.src.CreatedAt.UTC()
	if rp.src.CompletedAt.Valid {
		createdAt = rp.src.CompletedAt.Time.UTC()
	}
	err = withTx(ctx, db, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, `
			INSERT INTO test_record (id, created_at, updated_at, tenant_id, deleted_at,
				test_code, user_id, attempt_no, session_id, started_at, ended_at,
				backgrounds, raw_detail, raw_score, authoritative_score, grade, valid, invalid_reason)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NULL, $9, $10, NULL, $11, $12, $13, $14, true, NULL)
			ON CONFLICT (id) DO NOTHING`,
			oldPK, createdAt, rp.src.UpdatedAt.UTC(), migrTenant, deletedAt,
			rp.src.TestCode, rp.userID, rp.attemptNo,
			time.UnixMilli(rp.src.StartTime).UTC(), time.UnixMilli(rp.src.EndTime).UTC(),
			rp.rawDetail, nullFloat(rp.src.RawScore), rp.src.Score, bandFor(rp.src.Score))
		if err != nil {
			return fmt.Errorf("insert test_record %s: %w", oldPK, err)
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return fmt.Errorf("test_record %s already exists outside the map — halting for a decision", oldPK)
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO migration_map (id, created_at, updated_at, old_pk, entity, new_id)
			VALUES ($1, now(), now(), $2, $3, $4)
			ON CONFLICT (entity, old_pk) DO NOTHING`,
			xid.New().String(), oldPK, entityTestRecord, oldPK); err != nil {
			return fmt.Errorf("insert map row for record %s: %w", oldPK, err)
		}
		return nil
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func nullFloat(v sql.NullFloat64) any {
	if !v.Valid {
		return nil
	}
	return v.Float64
}
