package main

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rs/xid"
)

// The entitlement cut translates the legacy unlock planes into the grant
// ledger. The legacy store has two eras, and they leave different evidence:
//
//   - Era-1 "global unlock": unlock() set is_unlocked with a method and,
//     for payment, a 30-day window; code unlocks were permanent. The flow
//     always wrote the code string on the activation record, so the source
//     artifact is recoverable.
//   - Era-2 "family unlock": unlockReport() set one of three boolean flags
//     and wrote an activation record WITHOUT the code string or method —
//     the family is only visible in the flags, the code identity is gone.
//
// The frozen translation ruling: old-code activations land as membership
// rows (the member plane), family flags land as report_type rows, and the
// flags are the primary truth — the ledger keeps one row per (user, scope),
// so records annotate (grant time, source) but never multiply rows.
const (
	entityEntitlement    = "entitlement"
	entityActivationCode = "activation_code"

	scopeMembership = "membership"
	scopeReportType = "report_type"

	familyCognitive = "cognitive"
	familyCareer    = "career"
	familySchool    = "school"

	// Era-2 rows: the code string never reached the record, so no artifact
	// id exists to name. A fixed sentinel keeps source non-empty without
	// inventing an attribution.
	sourceLegacy        = "legacy"
	sourceLegacyPayment = "legacy-payment"
)

// farFuture is the expiry sentinel for legacy codes that never expire: the
// new column is NOT NULL, so "no deadline" is written as a date no batch
// will outlive, and the census counts every sentinel so the rehearsal sees
// how often the substitution fired.
var farFuture = time.Date(9999, time.December, 31, 23, 59, 59, 0, time.UTC)

// srcUnlock carries the unlock-plane columns of the legacy user row.
type srcUnlock struct {
	ID                int64
	CreatedAt         time.Time
	IsDeleted         bool
	IsUnlocked        bool
	UnlockMethod      sql.NullString
	UnlockedAt        sql.NullTime
	UnlockExpiresAt   sql.NullTime
	UnlockedCognitive bool
	UnlockedCareer    bool
	UnlockedSchool    bool
}

// srcRecord is one activation_records row — the redemption flow. Era-1
// flows wrote method and code string; the era-2 flow wrote neither.
type srcRecord struct {
	ID             int64
	UserID         int64
	ActivationCode sql.NullString
	UnlockMethod   sql.NullString
	ActivatedAt    time.Time
	IsDeleted      bool
}

// eraGlobal reports the era-1 flows: those set unlock_method on the record.
func (r *srcRecord) eraGlobal() bool {
	m := nullStr(r.UnlockMethod)
	return m == "activation_code" || m == "payment"
}

// srcCode is one legacy activation_code row — the support artifact, possibly
// multi-use, bound to a (test) school whose binding the dissolved
// organization layer takes with it.
type srcCode struct {
	ID         int64
	Code       string
	ReportType sql.NullString
	Quantity   int64
	UsedCount  int64
	ExpiresAt  sql.NullTime
	CreatedBy  int64
	CreatedAt  time.Time
	IsDeleted  bool
}

func fetchUnlocks(ctx context.Context, db *sql.DB) ([]*srcUnlock, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, created_at, is_deleted, is_unlocked,
		unlock_method, unlocked_at, unlock_expires_at, unlocked_cognitive, unlocked_career, unlocked_school
		FROM `+"`user`"+` ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("query legacy unlock plane: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []*srcUnlock
	for rows.Next() {
		u := &srcUnlock{}
		var isDeleted, isUnlocked, cognitive, career, school int64
		err := rows.Scan(&u.ID, &u.CreatedAt, &isDeleted, &isUnlocked,
			&u.UnlockMethod, &u.UnlockedAt, &u.UnlockExpiresAt, &cognitive, &career, &school)
		if err != nil {
			return nil, fmt.Errorf("scan legacy unlock %d: %w", u.ID, err)
		}
		u.IsDeleted = isDeleted != 0
		u.IsUnlocked = isUnlocked != 0
		u.UnlockedCognitive = cognitive != 0
		u.UnlockedCareer = career != 0
		u.UnlockedSchool = school != 0
		out = append(out, u)
	}
	return out, rows.Err()
}

func fetchRecords(ctx context.Context, db *sql.DB) ([]*srcRecord, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, user_id, activation_code, unlock_method,
		activated_at, is_deleted FROM activation_records ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("query legacy activation_records: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []*srcRecord
	for rows.Next() {
		r := &srcRecord{}
		var isDeleted int64
		err := rows.Scan(&r.ID, &r.UserID, &r.ActivationCode, &r.UnlockMethod, &r.ActivatedAt, &isDeleted)
		if err != nil {
			return nil, fmt.Errorf("scan legacy activation_record %d: %w", r.ID, err)
		}
		r.IsDeleted = isDeleted != 0
		out = append(out, r)
	}
	return out, rows.Err()
}

func fetchCodes(ctx context.Context, db *sql.DB) ([]*srcCode, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, code, report_type, quantity, used_count,
		expires_at, created_by, created_at, is_deleted FROM activation_code ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("query legacy activation_code: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []*srcCode
	for rows.Next() {
		c := &srcCode{}
		var isDeleted int64
		err := rows.Scan(&c.ID, &c.Code, &c.ReportType, &c.Quantity, &c.UsedCount,
			&c.ExpiresAt, &c.CreatedBy, &c.CreatedAt, &isDeleted)
		if err != nil {
			return nil, fmt.Errorf("scan legacy activation_code %d: %w", c.ID, err)
		}
		c.IsDeleted = isDeleted != 0
		out = append(out, c)
	}
	return out, rows.Err()
}

// entRow is one ledger row to write; oldPK is the synthetic migration_map
// key (the ledger has no legacy counterpart row to name).
type entRow struct {
	oldPK    string
	kind     string
	scopeRef *string
	expires  *time.Time
	source   string
	grantAt  time.Time
}

type entUserPlan struct {
	src  *srcUnlock
	rows []entRow
}

type codePlan struct {
	src        *srcCode
	expires    time.Time
	sentinel   bool
	redeemedBy *string
	redeemedAt *time.Time
}

type entCensus struct {
	usersLive    int
	usersDeleted int

	unlockTotal     int
	unlockByMethod  map[string]int
	unlockWindowed  int
	sourceCode      int
	sourcePayment   int
	sourceFallback  int
	grantFromRecord int
	grantFromUnlock int
	grantFromCreate int

	flagCognitive     int
	flagCareer        int
	flagSchoolDropped int
	flagNoRecord      int

	recordsEra1   int
	recordsEra2   int
	recordsDead   int
	recordsOrphan int

	codesTotal            int
	codesMigrate          int
	codesDroppedDead      int
	codesDroppedGlobal    int
	codesDroppedExhausted int
	codesDroppedExpired   int
	codesSentinelExpiry   int

	trailSingle  int
	trailMulti   int
	trailMissing int

	deletedWithEntitlements int
}

func runEntitlement(ctx context.Context, env *environment) error {
	unlocks, err := fetchUnlocks(ctx, env.src)
	if err != nil {
		return err
	}
	records, err := fetchRecords(ctx, env.src)
	if err != nil {
		return err
	}
	codes, err := fetchCodes(ctx, env.src)
	if err != nil {
		return err
	}

	userIDs := map[int64]struct{}{}
	for _, u := range unlocks {
		userIDs[u.ID] = struct{}{}
	}
	recsEra1ByUser := map[int64][]*srcRecord{}
	recsEra2ByUser := map[int64][]*srcRecord{}
	recsByCode := map[string][]*srcRecord{}
	c := &entCensus{unlockByMethod: map[string]int{}}
	for _, r := range records {
		if r.IsDeleted {
			c.recordsDead++
			continue
		}
		if _, ok := userIDs[r.UserID]; !ok {
			c.recordsOrphan++
			continue
		}
		if r.eraGlobal() {
			c.recordsEra1++
			recsEra1ByUser[r.UserID] = append(recsEra1ByUser[r.UserID], r)
		} else {
			c.recordsEra2++
			recsEra2ByUser[r.UserID] = append(recsEra2ByUser[r.UserID], r)
		}
		if code := nullStr(r.ActivationCode); code != "" {
			recsByCode[code] = append(recsByCode[code], r)
		}
	}

	plans, codePlans, err := buildEntitlementPlans(unlocks, recsEra1ByUser, recsEra2ByUser, recsByCode, codes, c)
	if err != nil {
		return err
	}
	printEntCensus(c)

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

	entIns, entSkip, entAdopted := 0, 0, 0
	for _, p := range plans {
		if len(p.rows) == 0 {
			continue
		}
		if _, ok := mapped[strconv.FormatInt(p.src.ID, 10)]; !ok {
			return fmt.Errorf("live user %d absent from migration_map — the identity phase has not covered this store, halting", p.src.ID)
		}
		ins, skip, adopted, err := applyEsportsEntitlement(ctx, env.esp, p)
		if err != nil {
			return err
		}
		entIns += ins
		entSkip += skip
		entAdopted += adopted
	}

	codeIns, codeSkip := 0, 0
	for _, cp := range codePlans {
		done, err := applyEsportsCode(ctx, env.esp, cp)
		if err != nil {
			return err
		}
		if done {
			codeIns++
		} else {
			codeSkip++
		}
	}

	fmt.Printf("entitlement: rows %d written / %d already-mapped / %d adopted (same-scope grant predated the map)\n", entIns, entSkip, entAdopted)
	fmt.Printf("codes: %d written / %d already-mapped\n", codeIns, codeSkip)
	return nil
}

func buildEntitlementPlans(unlocks []*srcUnlock, era1, era2 map[int64][]*srcRecord, recsByCode map[string][]*srcRecord, codes []*srcCode, c *entCensus) ([]*entUserPlan, []*codePlan, error) {
	now := time.Now()
	plans := make([]*entUserPlan, 0, len(unlocks))

	for _, u := range unlocks {
		if u.IsDeleted {
			c.usersDeleted++
			if u.IsUnlocked || u.UnlockedCognitive || u.UnlockedCareer {
				c.deletedWithEntitlements++
			}
			continue // dead identities keep no grants; their history stays in the persisted map and the archived old stack
		}
		c.usersLive++
		p := &entUserPlan{src: u}

		if u.IsUnlocked {
			c.unlockTotal++
			method := nullStr(u.UnlockMethod)
			if method == "" {
				method = "(unset)"
			}
			c.unlockByMethod[method]++

			row := entRow{kind: scopeMembership, oldPK: "unlock:" + strconv.FormatInt(u.ID, 10)}
			// Code-era unlocks were permanent (the flow never set the user
			// column); only the payment branch wrote a window.
			if u.UnlockExpiresAt.Valid {
				v := u.UnlockExpiresAt.Time.UTC()
				row.expires = &v
				c.unlockWindowed++
			}
			row.source, c.sourceCode, c.sourcePayment, c.sourceFallback = membershipSource(u, era1[u.ID], c.sourceCode, c.sourcePayment, c.sourceFallback)
			row.grantAt = grantTime(
				[]sql.NullTime{u.UnlockedAt, earliestRecord(era1[u.ID]), nullTime(u.CreatedAt)},
				[]*int{&c.grantFromUnlock, &c.grantFromRecord, &c.grantFromCreate})
			p.rows = append(p.rows, row)
		}

		for _, fam := range []struct {
			on    bool
			name  string
			count *int
		}{
			{u.UnlockedCognitive, familyCognitive, &c.flagCognitive},
			{u.UnlockedCareer, familyCareer, &c.flagCareer},
		} {
			if !fam.on {
				continue
			}
			*fam.count++
			if len(era2[u.ID]) == 0 {
				c.flagNoRecord++
			}
			// The legacy family gate was the boolean flag — it never
			// lapsed, so the migrated grant is permanent; the new redeem
			// path's code-expiry windows govern new redemptions only.
			ref := fam.name
			p.rows = append(p.rows, entRow{
				oldPK:    "flag:" + strconv.FormatInt(u.ID, 10) + ":" + fam.name,
				kind:     scopeReportType,
				scopeRef: &ref,
				source:   sourceLegacy,
				grantAt: grantTime(
					[]sql.NullTime{earliestRecord(era2[u.ID]), u.UnlockedAt, nullTime(u.CreatedAt)},
					[]*int{&c.grantFromRecord, &c.grantFromUnlock, &c.grantFromCreate}),
			})
		}
		if u.UnlockedSchool {
			// Default-1 for every legacy user and unattributable to any
			// redemption (the era-2 record never named the family), the
			// school flag is default state, not a grant; the family's
			// content plane left with the dissolved organization layer.
			c.flagSchoolDropped++
		}

		plans = append(plans, p)
	}

	codePlans := make([]*codePlan, 0, len(codes))
	for _, code := range codes {
		c.codesTotal++
		if !code.IsDeleted && nullStr(code.ReportType) != "" &&
			nullStr(code.ReportType) != familyCognitive && nullStr(code.ReportType) != familyCareer && nullStr(code.ReportType) != familySchool {
			return nil, nil, fmt.Errorf("code %d (%s) carries report_type %q outside the family set — halting for a decision instead of guessing",
				code.ID, code.Code, nullStr(code.ReportType))
		}
		switch {
		case code.IsDeleted:
			c.codesDroppedDead++
			continue
		case nullStr(code.ReportType) == "":
			// Global-era codes: the new model has no global unlock to
			// redeem — dropped and counted so the rehearsal can escalate
			// if real handed-out stock turns up.
			c.codesDroppedGlobal++
			continue
		case code.UsedCount >= code.Quantity:
			c.codesDroppedExhausted++
			continue
		case code.ExpiresAt.Valid && !code.ExpiresAt.Time.After(now):
			c.codesDroppedExpired++
			continue
		}

		cp := &codePlan{src: code}
		if code.ExpiresAt.Valid {
			cp.expires = code.ExpiresAt.Time.UTC()
		} else {
			cp.expires = farFuture
			cp.sentinel = true
			c.codesSentinelExpiry++
		}
		if code.UsedCount > 0 {
			// The redeem trail only lands when exactly one live record
			// names the code: multi-use codes have no single owner, and
			// the ledger rows carry the full history regardless.
			switch rs := recsByCode[code.Code]; len(rs) {
			case 1:
				by := strconv.FormatInt(rs[0].UserID, 10)
				at := rs[0].ActivatedAt.UTC()
				cp.redeemedBy, cp.redeemedAt = &by, &at
				c.trailSingle++
			case 0:
				c.trailMissing++
			default:
				c.trailMulti++
			}
		}
		c.codesMigrate++
		codePlans = append(codePlans, cp)
	}
	return plans, codePlans, nil
}

// membershipSource names the granting artifact: the earliest era-1 record's
// code string when the flow left one (the unlock flow refused an already
// unlocked user, so the first era-1 record IS the global unlock; later
// same-shaped rows are drift and must not overwrite the attribution), a
// payment marker for the gateway-less payment branch, and a plain sentinel
// when no record survives.
func membershipSource(u *srcUnlock, recs []*srcRecord, nCode, nPayment, nFallback int) (string, int, int, int) {
	if nullStr(u.UnlockMethod) == "payment" {
		nPayment++
		return sourceLegacyPayment, nCode, nPayment, nFallback
	}
	var earliest *srcRecord
	for _, r := range recs {
		if earliest == nil || r.ActivatedAt.Before(earliest.ActivatedAt) {
			earliest = r
		}
	}
	if earliest != nil && nullStr(earliest.ActivationCode) != "" {
		nCode++
		return nullStr(earliest.ActivationCode), nCode, nPayment, nFallback
	}
	nFallback++
	return sourceLegacy, nCode, nPayment, nFallback
}

// grantTime answers the first present timestamp in the chain, counting
// which tier supplied it — grant times are approximations by necessity (the
// legacy row kept one clock for every plane), and the census states how
// coarse each one is. The final fallback is unreachable in practice: the
// legacy creation column is NOT NULL.
func grantTime(chain []sql.NullTime, counters []*int) time.Time {
	for i, ct := range chain {
		if ct.Valid {
			*counters[i]++
			return ct.Time.UTC()
		}
	}
	return time.Now().UTC()
}

func nullTime(t time.Time) sql.NullTime {
	return sql.NullTime{Time: t, Valid: true}
}

func earliestRecord(recs []*srcRecord) sql.NullTime {
	if len(recs) == 0 {
		return sql.NullTime{}
	}
	earliest := recs[0].ActivatedAt
	for _, r := range recs[1:] {
		if r.ActivatedAt.Before(earliest) {
			earliest = r.ActivatedAt
		}
	}
	return sql.NullTime{Time: earliest, Valid: true}
}

func loadMappedUsers(ctx context.Context, db *sql.DB) (map[string]struct{}, error) {
	rows, err := db.QueryContext(ctx, `SELECT old_pk FROM migration_map WHERE entity = $1`, entityUser)
	if err != nil {
		return nil, fmt.Errorf("load mapped users: %w", err)
	}
	defer func() { _ = rows.Close() }()
	out := map[string]struct{}{}
	for rows.Next() {
		var pk string
		if err := rows.Scan(&pk); err != nil {
			return nil, fmt.Errorf("scan mapped user: %w", err)
		}
		out[pk] = struct{}{}
	}
	return out, rows.Err()
}

// applyEsportsEntitlement writes one user's ledger rows and their map
// markers in one transaction; the map rows are the pair's commit point.
// A row whose unique scope already exists without a map row is adopted:
// same (user, scope) grant, one row — the map just starts claiming it.
//
// Column list pinned to the esports service's Atlas schema for
// "entitlement"; a schema change there must re-run the rehearsal's
// sampling pass.
func applyEsportsEntitlement(ctx context.Context, db *sql.DB, p *entUserPlan) (ins, skip, adopted int, err error) {
	for _, row := range p.rows {
		var exists bool
		err = db.QueryRowContext(ctx,
			`SELECT EXISTS (SELECT 1 FROM migration_map WHERE entity = $1 AND old_pk = $2)`,
			entityEntitlement, row.oldPK).Scan(&exists)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("query map for %s: %w", row.oldPK, err)
		}
		if exists {
			skip++
			continue
		}
		userID := strconv.FormatInt(p.src.ID, 10)
		err = withTx(ctx, db, func(tx *sql.Tx) error {
			res, err := tx.ExecContext(ctx, `
				INSERT INTO entitlement (id, created_at, updated_at, tenant_id, user_id,
					scope_kind, scope_ref, expires_at, source, status, terminated_at)
				VALUES ($1, $2, $2, $3, $4, $5, $6, $7, $8, 'active', NULL)
				ON CONFLICT (tenant_id, user_id, scope_kind, scope_ref) DO NOTHING`,
				xid.New().String(), row.grantAt, migrTenant, userID,
				row.kind, row.scopeRef, row.expires, row.source)
			if err != nil {
				return fmt.Errorf("insert entitlement %s: %w", row.oldPK, err)
			}
			if n, _ := res.RowsAffected(); n == 0 {
				adopted++
			}
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO migration_map (id, created_at, updated_at, old_pk, entity, new_id)
				VALUES ($1, now(), now(), $2, $3, $4)
				ON CONFLICT (entity, old_pk) DO NOTHING`,
				xid.New().String(), row.oldPK, entityEntitlement, userID); err != nil {
				return fmt.Errorf("insert map row for %s: %w", row.oldPK, err)
			}
			return nil
		})
		if err != nil {
			return 0, 0, 0, err
		}
		ins++
	}
	return ins, skip, adopted, nil
}

// applyEsportsCode writes one unredeemed legacy code as a bare code row —
// batch_id NULL is the schema's own home for rows predating the batch
// model, and the redeem path treats a null batch as live. A code string
// colliding with a new-stack mint is never adopted silently: the collision
// halts for a decision.
func applyEsportsCode(ctx context.Context, db *sql.DB, cp *codePlan) (bool, error) {
	oldPK := strconv.FormatInt(cp.src.ID, 10)
	var exists bool
	err := db.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM migration_map WHERE entity = $1 AND old_pk = $2)`,
		entityActivationCode, oldPK).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("query map for code %s: %w", oldPK, err)
	}
	if exists {
		return false, nil
	}

	id := xid.New().String()
	err = withTx(ctx, db, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, `
			INSERT INTO activation_code (id, created_at, updated_at, tenant_id, code,
				report_type, quantity, used_count, expires_at, created_by,
				redeemed_by, redeemed_at, batch_id)
			VALUES ($1, $2, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NULL)
			ON CONFLICT (code) DO NOTHING`,
			id, cp.src.CreatedAt.UTC(), migrTenant, cp.src.Code, nullStr(cp.src.ReportType),
			cp.src.Quantity, cp.src.UsedCount, cp.expires, strconv.FormatInt(cp.src.CreatedBy, 10),
			cp.redeemedBy, cp.redeemedAt)
		if err != nil {
			return fmt.Errorf("insert activation_code %s: %w", cp.src.Code, err)
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return fmt.Errorf("code %q already exists outside the map — a new-stack mint colliding with a legacy string halts for a decision", cp.src.Code)
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO migration_map (id, created_at, updated_at, old_pk, entity, new_id)
			VALUES ($1, now(), now(), $2, $3, $4)
			ON CONFLICT (entity, old_pk) DO NOTHING`,
			xid.New().String(), oldPK, entityActivationCode, id); err != nil {
			return fmt.Errorf("insert map row for code %s: %w", oldPK, err)
		}
		return nil
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func printEntCensus(c *entCensus) {
	fmt.Printf("entitlement plan: %d users (%d live / %d deleted; %d deleted users held grants — kept in map+archive only)\n",
		c.usersLive+c.usersDeleted, c.usersLive, c.usersDeleted, c.deletedWithEntitlements)
	fmt.Printf("  era-1 global unlocks: %d (methods: %s; windowed %d) — sources: code %d / payment %d / fallback %d\n",
		c.unlockTotal, kvSummary(c.unlockByMethod), c.unlockWindowed, c.sourceCode, c.sourcePayment, c.sourceFallback)
	fmt.Printf("  era-2 family flags: cognitive %d, career %d, school %d dropped (default-1 unattributable, org layer dissolved); flags without any record %d\n",
		c.flagCognitive, c.flagCareer, c.flagSchoolDropped, c.flagNoRecord)
	fmt.Printf("  grant time tiers: from records %d / from unlock column %d / from row creation %d\n",
		c.grantFromRecord, c.grantFromUnlock, c.grantFromCreate)
	fmt.Printf("  records: era-1 %d, era-2 %d, dead %d dropped, orphan %d\n",
		c.recordsEra1, c.recordsEra2, c.recordsDead, c.recordsOrphan)
	fmt.Printf("  codes: %d total → %d migrate; dropped: dead %d / global-type %d / exhausted %d / expired %d; NULL-expiry sentinel %d; redeem trail: single %d / multi %d / used-no-record %d\n",
		c.codesTotal, c.codesMigrate, c.codesDroppedDead, c.codesDroppedGlobal,
		c.codesDroppedExhausted, c.codesDroppedExpired, c.codesSentinelExpiry,
		c.trailSingle, c.trailMulti, c.trailMissing)
	if c.recordsOrphan > 0 {
		fmt.Printf("  ⚠ orphan activation records (user row absent): %d\n", c.recordsOrphan)
	}
}

func kvSummary(m map[string]int) string {
	parts := make([]string, 0, len(m))
	for k, v := range m {
		parts = append(parts, fmt.Sprintf("%s=%d", k, v))
	}
	sort.Strings(parts)
	return strings.Join(parts, ", ")
}
