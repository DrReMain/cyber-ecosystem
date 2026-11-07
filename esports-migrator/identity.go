package main

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rs/xid"
)

// runGuard halts when the dissolved organization layer turns out to have
// rows. The layer's deletion rests on "no production school data exists";
// a non-empty table voids that premise and needs a human decision, not a
// guess — silently skipping related tables is how ghost foreign keys are
// born.
func runGuard(ctx context.Context, env *environment) error {
	school, schoolReport, err := schoolRowCounts(ctx, env.src)
	if err != nil {
		return err
	}
	fmt.Printf("school rows: %d, school_report rows: %d\n", school, schoolReport)
	if school > 0 || schoolReport > 0 {
		if !env.allowSchoolResidue {
			return fmt.Errorf("organization layer not empty (school=%d, school_report=%d) — the no-school-data premise is void, halting for a decision (ruling D-MIG-1(a) proceeds with --allow-school-residue)", school, schoolReport)
		}
		// Ruling D-MIG-1(a): the residue is test data plus 22 rows of 7
		// real users, all belonging to a product layer the new version
		// dissolved — nothing migrates, and the sealed source dump keeps
		// every row retrievable. The bypass is explicit and the counts
		// stay in the run log as evidence.
		fmt.Printf("D-MIG-1(a) bypass: %d school + %d school_report rows ruled not-migrated (org layer dissolved); the sealed source dump retains every row\n", school, schoolReport)
	}
	return nil
}

// userPlan is the decided destination of one legacy user row, built
// entirely before any write so anomalies surface pre-flight.
type userPlan struct {
	src *srcUser

	account             string
	passwordHash        string // "" permanently closes the password plane (the auto-registration contract)
	bcryptCarried       bool
	forcedReset         bool
	nickname            *string
	nicknameTruncated   bool
	birthYear           *int32
	birthMonth          *int32
	ageRangeOnly        bool // legacy 4-band age with no birth date: unmappable to the 8-band set, profile stays incomplete
	ageBand             *string
	avatarURL           *string
	wechatRegionDropped bool
	schoolBound         bool
	lockedOut           bool // live identity with a closed password plane and no provider credential
	accountRenamed      bool // D-MIG-2: lost its collision group, keeps the `<account>#<legacy_id>` handle
}

func (p *userPlan) srcID() string { return strconv.FormatInt(p.src.ID, 10) }
func (p *userPlan) live() bool    { return !p.src.IsDeleted }
func (p *userPlan) enabled() bool { return p.live() && p.src.Status == 1 }
func (p *userPlan) deletedAt() any {
	if p.live() {
		return nil
	}
	return p.src.UpdatedAt.UTC()
}

type census struct {
	total               int
	live                int
	deleted             int
	bcryptCarried       int
	forcedReset         int
	nicknameTruncated   int
	wechatRegionDropped int
	schoolBound         int
	ageRangeOnly        int
	lockedOut           int
	algorithmCounts     map[string]int
	lockedOutExamples   []string
	renamedAccounts     int
	renamedExamples     []string
}

func buildPlans(users []*srcUser) ([]*userPlan, *census, error) {
	c := &census{algorithmCounts: map[string]int{}}
	plans := make([]*userPlan, 0, len(users))
	now := time.Now()

	for _, u := range users {
		c.total++
		c.algorithmCounts[u.PasswordAlgorithm]++

		p := &userPlan{src: u}
		if u.IsDeleted {
			c.deleted++
		} else {
			c.live++
		}

		// Account handle: the legacy login key with the strongest identity
		// meaning wins; a random handle is the last resort (the live
		// system's own convention for provider-born identities).
		switch {
		case nullStr(u.Username) != "":
			p.account = nullStr(u.Username)
		case nullStr(u.Phone) != "":
			p.account = nullStr(u.Phone)
		case nullStr(u.OpenID) != "":
			p.account = nullStr(u.OpenID)
		default:
			p.account = xid.New().String()
		}

		// Password: only a digest the new system can actually verify is
		// worth carrying (it upgrades on next login). MD5 and MD5-wrapped
		// bcrypt can never verify here, so their password plane is closed
		// and the owner signs in through phone/wechat or re-registers.
		if u.PasswordAlgorithm == "BCRYPT" && u.Password.Valid && u.Password.String != "" {
			p.passwordHash = u.Password.String
			p.bcryptCarried = true
			c.bcryptCarried++
		} else {
			p.forcedReset = true
			c.forcedReset++
		}

		if nick := strings.TrimSpace(nullStr(u.Nickname)); nick != "" {
			runes := []rune(nick)
			if len(runes) > 12 {
				nick = string(runes[:12])
				p.nicknameTruncated = true
				c.nicknameTruncated++
			}
			p.nickname = &nick
		}

		if u.BirthDate.Valid {
			y := clampInt32(u.BirthDate.Time.Year())
			m := clampInt32(int(u.BirthDate.Time.Month()))
			p.birthYear, p.birthMonth = &y, &m
			p.ageBand = ageBandOf(&y, &m, now)
		} else if nullStr(u.AgeRange) != "" {
			p.ageRangeOnly = true
			c.ageRangeOnly++
		}

		if nullStr(u.AvatarURL) != "" {
			v := nullStr(u.AvatarURL)
			p.avatarURL = &v
		}

		if nullStr(u.WechatCountry) != "" || nullStr(u.WechatProvince) != "" || nullStr(u.WechatCity) != "" {
			p.wechatRegionDropped = true
			c.wechatRegionDropped++
		}
		if u.SchoolID.Valid && u.SchoolID.Int64 != 0 {
			p.schoolBound = true
			c.schoolBound++
		}

		if p.live() && nullStr(u.Phone) == "" && nullStr(u.OpenID) == "" && p.forcedReset {
			p.lockedOut = true
			c.lockedOut++
			if len(c.lockedOutExamples) < 5 {
				c.lockedOutExamples = append(c.lockedOutExamples, p.srcID())
			}
		}

		plans = append(plans, p)
	}
	resolveAccountCollisions(plans, c)
	return orderRenamedLast(plans), c, nil
}

// resolveAccountCollisions applies ruling D-MIG-2's mechanical rule to
// every group of legacy rows deriving the same account handle: exactly one
// row keeps the original account, every other row is renamed
// `<account>#<legacy_id>`. Zero product judgement, zero dropped rows:
// a renamed handle is never a login key (member logins ride the
// phone/wechat credentials), soft-deleted identities have no login
// surface at all, and migration_map keeps every legacy id's trail.
func resolveAccountCollisions(plans []*userPlan, c *census) {
	byAccount := make(map[string][]*userPlan, len(plans))
	for _, p := range plans {
		byAccount[p.account] = append(byAccount[p.account], p)
	}
	for _, group := range byAccount {
		if len(group) < 2 {
			continue
		}
		winner := group[0]
		for _, p := range group[1:] {
			if keepsOriginalAccount(p, winner) {
				winner = p
			}
		}
		for _, p := range group {
			if p == winner {
				continue
			}
			p.account = fmt.Sprintf("%s#%s", p.account, p.srcID())
			p.accountRenamed = true
			c.renamedAccounts++
			if len(c.renamedExamples) < 5 {
				c.renamedExamples = append(c.renamedExamples, p.srcID())
			}
		}
	}
}

// keepsOriginalAccount is true when a outranks b as the holder of a
// collision group's original account: a live identity beats a deleted one
// (a dead row's handle is pure bookkeeping), then the most recently
// updated row wins, then the smaller legacy id keeps the order total.
func keepsOriginalAccount(a, b *userPlan) bool {
	if a.live() != b.live() {
		return a.live()
	}
	if !a.src.UpdatedAt.Equal(b.src.UpdatedAt) {
		return a.src.UpdatedAt.After(b.src.UpdatedAt)
	}
	return a.src.ID < b.src.ID
}

// orderRenamedLast moves every renamed (losing) plan behind all others.
// The credential inserts are first-wins on conflict, so a group winner's
// phone/wechat credential must land before its renamed mates': the
// original account's owner is also the credential's owner.
func orderRenamedLast(plans []*userPlan) []*userPlan {
	ordered := make([]*userPlan, 0, len(plans))
	renamed := make([]*userPlan, 0)
	for _, p := range plans {
		if p.accountRenamed {
			renamed = append(renamed, p)
		} else {
			ordered = append(ordered, p)
		}
	}
	return append(ordered, renamed...)
}

func nullStr(s sql.NullString) string {
	if s.Valid {
		return strings.TrimSpace(s.String)
	}
	return ""
}

// Bounds-guarded narrowing in place of the platform's ConvNum seam: this
// module deliberately imports nothing from the tree, so the guard lives
// here (the capability-package precedent).
func clampInt32(v int) int32 {
	if v > math.MaxInt32 {
		return math.MaxInt32
	}
	if v < math.MinInt32 {
		return math.MinInt32
	}
	return int32(v)
}

func runIdentity(ctx context.Context, env *environment) error {
	users, err := fetchUsers(ctx, env.src)
	if err != nil {
		return err
	}
	plans, c, err := buildPlans(users)
	if err != nil {
		return err
	}
	printCensus(c, len(plans))

	if env.dryRun {
		fmt.Println("dry-run: no writes performed")
		return nil
	}
	if err := env.ensureTargets(); err != nil {
		return err
	}
	if err := ensureMapTable(ctx, env.esp); err != nil {
		return err
	}

	sysIns, sysSkip, credIns, memIns := 0, 0, 0, 0
	err = withTx(ctx, env.sys, func(tx *sql.Tx) error {
		for _, p := range plans {
			inserted, err := insertSystemUser(ctx, tx, p)
			if err != nil {
				return err
			}
			if inserted {
				sysIns++
			} else {
				sysSkip++
			}
			if !p.live() {
				continue // dead identities keep no credentials and no live seat
			}
			if nullStr(p.src.Phone) != "" {
				inserted, err := insertSystemCredential(ctx, tx, kindSMS, nullStr(p.src.Phone), p.srcID(), p.src.CreatedAt)
				if err != nil {
					return err
				}
				if inserted {
					credIns++
				}
			}
			if nullStr(p.src.OpenID) != "" {
				inserted, err := insertSystemCredential(ctx, tx, kindWechat, nullStr(p.src.OpenID), p.srcID(), p.src.CreatedAt)
				if err != nil {
					return err
				}
				if inserted {
					credIns++
				}
			}
			inserted, err = insertSystemMembership(ctx, tx, p.srcID(), p.src.CreatedAt)
			if err != nil {
				return err
			}
			if inserted {
				memIns++
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	espIns, espSkip := 0, 0
	for _, p := range plans {
		done, err := applyEsportsIdentity(ctx, env.esp, p)
		if err != nil {
			return err
		}
		if done {
			espIns++
		} else {
			espSkip++
		}
	}

	fmt.Printf("system: users %d inserted / %d already-present, credentials %d, memberships %d\n", sysIns, sysSkip, credIns, memIns)
	fmt.Printf("esports: profiles+map %d written / %d already-mapped\n", espIns, espSkip)
	return nil
}

func withTx(ctx context.Context, db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// ensureMapTable creates the tool's own bookkeeping table. It holds the
// persisted legacy→new id mapping — migration evidence covering every
// legacy row, live and soft-deleted alike — and deliberately lives outside
// the service's schema: the product must not carry one-shot migration
// structure, and the table survives in the database as evidence after this
// tool is deleted.
func ensureMapTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS migration_map (
			id varchar(32) NOT NULL PRIMARY KEY,
			created_at timestamptz NOT NULL,
			updated_at timestamptz NOT NULL,
			old_pk varchar(64) NOT NULL,
			entity varchar(16) NOT NULL,
			new_id varchar(20) NOT NULL,
			CONSTRAINT migration_map_entity_old_pk_key UNIQUE (entity, old_pk)
		)`)
	if err != nil {
		return fmt.Errorf("ensure migration_map: %w", err)
	}
	return nil
}

// applyEsportsIdentity writes the profile row and its migration_map marker
// in one transaction; the map row is the pair's commit point, so a re-run
// after a crash resumes exactly where the map stops. The legacy snowflake
// id doubles as the new user id (carried verbatim, never re-keyed).
//
// Column list pinned to the esports service's Atlas schema for "profile";
// a schema change there must re-run the rehearsal's sampling pass.
func applyEsportsIdentity(ctx context.Context, db *sql.DB, p *userPlan) (bool, error) {
	var exists bool
	err := db.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM migration_map WHERE entity = $1 AND old_pk = $2)`,
		entityUser, p.srcID()).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("query map for %s: %w", p.srcID(), err)
	}
	if exists {
		return false, nil
	}

	err = withTx(ctx, db, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, `
			INSERT INTO profile (id, created_at, updated_at, tenant_id, deleted_at,
				user_id, nickname, birth_year, birth_month, sex, age_band, avatar_url,
				guardian_consented_at, games)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NULL, NULL)
			ON CONFLICT (user_id) WHERE deleted_at IS NULL DO NOTHING`,
			xid.New().String(), p.src.CreatedAt.UTC(), p.src.UpdatedAt.UTC(), migrTenant, p.deletedAt(),
			p.srcID(), p.nickname, birthCol(p.birthYear), birthCol(p.birthMonth),
			sexOf(p.src.WechatSex), p.ageBand, p.avatarURL)
		if err != nil {
			return fmt.Errorf("insert profile for %s: %w", p.srcID(), err)
		}
		if n, _ := res.RowsAffected(); n == 0 && p.live() {
			// A live member already owning a profile row without a map row
			// means someone else created it — not ours to overwrite.
			return fmt.Errorf("profile for live user %s already exists outside the map — halting for a decision", p.srcID())
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO migration_map (id, created_at, updated_at, old_pk, entity, new_id)
			VALUES ($1, now(), now(), $2, $3, $4)
			ON CONFLICT (entity, old_pk) DO NOTHING`,
			xid.New().String(), p.srcID(), entityUser, p.srcID()); err != nil {
			return fmt.Errorf("insert map row for %s: %w", p.srcID(), err)
		}
		return nil
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func birthCol(v *int32) any {
	if v == nil {
		return nil
	}
	return int64(*v)
}

func sexOf(v sql.NullInt64) string {
	switch v.Int64 {
	case 1:
		return "male"
	case 2:
		return "female"
	default:
		return "undetermined" // legacy 0/NULL both mean never said
	}
}

// derivedAge and ageBandOf mirror the live write path's derivation
// verbatim (module/profile in the esports service). The band vocabulary is
// frozen product wording — under12|12-14|15-17|18-22|23-25|26-29|30-40|
// over40 — and band_test.go pins the boundaries; the reconcile phase
// re-derives a sample in SQL as the independent drift check.
func derivedAge(year, month *int32, now time.Time) *int {
	if year == nil || month == nil {
		return nil
	}
	age := now.Year() - int(*year)
	if int(now.Month()) < int(*month) {
		age--
	}
	if age < 0 {
		age = 0
	}
	return &age
}

func ageBandOf(year, month *int32, now time.Time) *string {
	age := derivedAge(year, month, now)
	if age == nil {
		return nil
	}
	var band string
	switch a := *age; {
	case a < 12:
		band = "under12"
	case a <= 14:
		band = "12-14"
	case a <= 17:
		band = "15-17"
	case a <= 22:
		band = "18-22"
	case a <= 25:
		band = "23-25"
	case a <= 29:
		band = "26-29"
	case a <= 40:
		band = "30-40"
	default:
		band = "over40"
	}
	return &band
}

func printCensus(c *census, planned int) {
	fmt.Printf("identity plan: %d rows (%d live / %d deleted)\n", planned, c.live, c.deleted)
	fmt.Printf("  password: bcrypt carried %d, forced reset %d (algorithms: %s)\n",
		c.bcryptCarried, c.forcedReset, algorithmSummary(c.algorithmCounts))
	fmt.Printf("  profile: nickname truncated %d, age-range-only (incomplete state) %d\n", c.nicknameTruncated, c.ageRangeOnly)
	fmt.Printf("  dropped fields: wechat region %d, school binding %d (org layer dissolved)\n", c.wechatRegionDropped, c.schoolBound)
	if c.lockedOut > 0 {
		fmt.Printf("  ⚠ locked out (password closed, no phone/openid): %d — e.g. %s; these owners cannot sign in and need the reset/announce path\n",
			c.lockedOut, strings.Join(c.lockedOutExamples, ", "))
	}
	if c.renamedAccounts > 0 {
		fmt.Printf("  account collisions resolved (D-MIG-2): %d rows renamed to `<account>#<legacy_id>` — e.g. legacy ids %s\n",
			c.renamedAccounts, strings.Join(c.renamedExamples, ", "))
	}
}

func algorithmSummary(counts map[string]int) string {
	parts := make([]string, 0, len(counts))
	for algo, n := range counts {
		if algo == "" {
			algo = "(unset)"
		}
		parts = append(parts, fmt.Sprintf("%s=%d", algo, n))
	}
	sort.Strings(parts)
	return strings.Join(parts, ", ")
}
