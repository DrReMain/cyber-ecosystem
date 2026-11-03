package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// The reconcile cut is the closing assertion pass: it re-derives every
// translation rule through a different route than the writers used and
// fails loudly on any disagreement. Three layers, cheapest first:
//
//   - count identities: the same population counted in source SQL and in
//     target SQL, scoped through migration_map so pre-existing target rows
//     (or a missing phase) break the identity instead of hiding in it;
//   - full-column invariants that live on one store or re-derive from the
//     source's exact ordering (grade is a function of the authoritative
//     score; attempt ordinals; the age band from the birth columns at the
//     map's own timestamp);
//   - per-field sampling: random source rows compared field by field
//     against the mapped target row, JSON payloads compared canonically so
//     whitespace and key order never mask a real diff.
//
// The phase is read-only everywhere and fails listing every violated check
// — a rehearsal that cannot say "all green" in one line is not green.
const (
	// Derivation-instant timezone: the writers derived age bands from
	// time.Now() on the operator's machine (team locale Asia/Shanghai); the
	// map row's created_at re-anchors that instant, rendered in the same
	// zone so the (year, month) the derivation saw is reproduced.
	derivationTZ = "Asia/Shanghai"

	expectedEntities = "activation_code,entitlement,profile_games,test_record,user"
)

type reconciler struct {
	sample int
	env    *environment
	// ids is the migrated user id set as a Postgres array literal — the
	// join spine for system-side counts (the map lives in the esports
	// database, the credential and membership rows in the system one).
	ids string

	passed int
	failed []string
}

// pair is one (label → count) side of a count identity.
type pair struct {
	label string
	n     int64
}

func (r *reconciler) check(id string, ok bool, detail string) {
	if ok {
		r.passed++
		fmt.Printf("reconcile: PASS %s — %s\n", id, detail)
		return
	}
	r.failed = append(r.failed, id)
	fmt.Printf("reconcile: FAIL %s — %s\n", id, detail)
}

// identity asserts every side of a count identity carries the same number.
func (r *reconciler) identity(id string, sides ...pair) {
	ok := true
	for _, s := range sides[1:] {
		if s.n != sides[0].n {
			ok = false
		}
	}
	parts := make([]string, 0, len(sides))
	for _, s := range sides {
		parts = append(parts, fmt.Sprintf("%s=%d", s.label, s.n))
	}
	r.check(id, ok, strings.Join(parts, " "))
}

func (r *reconciler) count(id string, db *sql.DB, q string, args ...any) int64 {
	var n int64
	if err := db.QueryRowContext(context.Background(), q, args...).Scan(&n); err != nil {
		r.check(id, false, "query error: "+err.Error())
		return -1
	}
	return n
}

// zero asserts a violation count is exactly zero.
func (r *reconciler) zero(id, q string, db *sql.DB, args ...any) {
	n := r.count(id, db, q, args...)
	r.check(id, n == 0, fmt.Sprintf("violations=%d", n))
}

func runReconcile(ctx context.Context, env *environment) error {
	if err := env.ensureTargets(); err != nil {
		return err
	}
	r := &reconciler{sample: env.sample, env: env}

	ids, err := legacyIDList(ctx, env.esp)
	if err != nil {
		return err
	}
	r.ids = ids

	// The map is the join spine: unknown entities mean someone wrote
	// bookkeeping rows this pass does not know how to verify.
	var entities sql.NullString
	err = env.esp.QueryRowContext(ctx,
		`SELECT string_agg(DISTINCT entity, ',' ORDER BY entity) FROM migration_map`).Scan(&entities)
	if err != nil {
		return fmt.Errorf("read map entities: %w", err)
	}
	known := strings.Split(expectedEntities, ",")
	unknown := 0
	for _, e := range strings.Split(entities.String, ",") {
		if e == "" {
			continue
		}
		found := false
		for _, k := range known {
			if e == k {
				found = true
			}
		}
		if !found {
			unknown++
		}
	}
	r.check("map.entities", unknown == 0,
		fmt.Sprintf("map entities %q, unknown=%d", entities.String, unknown))

	for _, e := range known {
		n := r.count("map.counts", env.esp, `SELECT COUNT(*) FROM migration_map WHERE entity = $1`, e)
		fmt.Printf("reconcile: map %-15s rows: %d\n", e, n)
	}

	if err := r.identityFamily(ctx); err != nil {
		return err
	}
	if err := r.entitlementFamily(ctx); err != nil {
		return err
	}
	if err := r.behaviorFamily(ctx); err != nil {
		return err
	}

	fmt.Printf("reconcile: %d checks, %d passed, %d failed\n", r.passed+len(r.failed), r.passed, len(r.failed))
	if len(r.failed) > 0 {
		return fmt.Errorf("reconcile failed: %s", strings.Join(r.failed, ", "))
	}
	return nil
}

// legacyIDList pulls the migrated user id set as a Postgres array literal.
func legacyIDList(ctx context.Context, esp *sql.DB) (string, error) {
	rows, err := esp.QueryContext(ctx, `SELECT new_id FROM migration_map WHERE entity = $1`, entityUser)
	if err != nil {
		return "", fmt.Errorf("load mapped ids: %w", err)
	}
	defer func() { _ = rows.Close() }()
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return "", err
		}
		ids = append(ids, id)
	}
	return "{" + strings.Join(ids, ",") + "}", rows.Err()
}

// scopedUser limits an esports query to users the identity phase claimed;
// the id list always binds $1, any further parameter starts at $2.
const scopedUser = `user_id IN (SELECT unnest($1::text[]))`

func (r *reconciler) identityFamily(ctx context.Context) error {
	src, sys, esp := r.env.src, r.env.sys, r.env.esp

	srcUsers := r.count("identity.user-count", src, "SELECT COUNT(*) FROM `user`")
	mapUsers := r.count("identity.user-count", esp, `SELECT COUNT(*) FROM migration_map WHERE entity = $1`, entityUser)
	sysUsers := r.count("identity.user-count", sys, `SELECT COUNT(*) FROM "user" WHERE id IN (SELECT unnest($1::text[]))`, r.ids)
	r.identity("identity.user-count",
		pair{"src", srcUsers}, pair{"map", mapUsers}, pair{"system", sysUsers})

	srcLive := r.count("identity.live-dead", src, "SELECT COUNT(*) FROM `user` WHERE is_deleted = 0")
	profLive := r.count("identity.live-dead", esp, `SELECT COUNT(*) FROM profile WHERE deleted_at IS NULL AND `+scopedUser, r.ids)
	profDead := r.count("identity.live-dead", esp, `SELECT COUNT(*) FROM profile WHERE deleted_at IS NOT NULL AND `+scopedUser, r.ids)
	r.identity("identity.live", pair{"src", srcLive}, pair{"profile", profLive})
	r.identity("identity.dead", pair{"src", srcUsers - srcLive}, pair{"profile", profDead})

	// Credentials: one per DISTINCT identifier among live users — the
	// writers' ON CONFLICT (kind, identifier) folds duplicates the same way.
	srcPhones := r.count("identity.credentials-sms", src,
		"SELECT COUNT(DISTINCT phone) FROM `user` WHERE is_deleted = 0 AND phone IS NOT NULL AND TRIM(phone) <> ''")
	srcOpenIDs := r.count("identity.credentials-wechat", src,
		"SELECT COUNT(DISTINCT open_id) FROM `user` WHERE is_deleted = 0 AND open_id IS NOT NULL AND TRIM(open_id) <> ''")
	sysSMS := r.count("identity.credentials-sms", sys,
		`SELECT COUNT(*) FROM credential WHERE kind = 'sms' AND user_id IN (SELECT unnest($1::text[]))`, r.ids)
	sysWechat := r.count("identity.credentials-wechat", sys,
		`SELECT COUNT(*) FROM credential WHERE kind = 'wechat_mp' AND user_id IN (SELECT unnest($1::text[]))`, r.ids)
	r.identity("identity.credentials-sms", pair{"src-distinct-phone", srcPhones}, pair{"credential", sysSMS})
	r.identity("identity.credentials-wechat", pair{"src-distinct-openid", srcOpenIDs}, pair{"credential", sysWechat})

	// Seat: every live legacy user holds exactly one default/member seat.
	sysMembership := r.count("identity.membership", sys,
		`SELECT COUNT(*) FROM membership WHERE tenant_id = 'default' AND population = 'member'
		AND user_id IN (SELECT unnest($1::text[]))`, r.ids)
	r.identity("identity.membership", pair{"src-live", srcLive}, pair{"membership", sysMembership})

	srcEnabled := r.count("identity.enabled", src, "SELECT COUNT(*) FROM `user` WHERE is_deleted = 0 AND status = 1")
	sysEnabled := r.count("identity.enabled", sys,
		`SELECT COUNT(*) FROM "user" WHERE enabled AND id IN (SELECT unnest($1::text[]))`, r.ids)
	r.identity("identity.enabled", pair{"src-live-status1", srcEnabled}, pair{"system-enabled", sysEnabled})

	// Password plane: only verifiable bcrypt carries; everything else closed.
	srcBcrypt := r.count("identity.password-plane", src,
		"SELECT COUNT(*) FROM `user` WHERE password_algorithm = 'BCRYPT' AND password IS NOT NULL AND password <> ''")
	sysHash := r.count("identity.password-plane", sys,
		`SELECT COUNT(*) FROM "user" WHERE password_hash IS NOT NULL AND password_hash <> ''
		AND id IN (SELECT unnest($1::text[]))`, r.ids)
	r.check("identity.password-plane", srcBcrypt == sysHash && srcBcrypt+(srcUsers-srcBcrypt) == srcUsers,
		fmt.Sprintf("src-bcrypt=%d system-hash=%d forced-reset=%d of %d users", srcBcrypt, sysHash, srcUsers-srcBcrypt, srcUsers))

	srcNick := r.count("identity.nickname-count", src,
		"SELECT COUNT(*) FROM `user` WHERE nickname IS NOT NULL AND TRIM(nickname) <> ''")
	profNick := r.count("identity.nickname-count", esp,
		`SELECT COUNT(*) FROM profile WHERE nickname IS NOT NULL AND `+scopedUser, r.ids)
	r.identity("identity.nickname-count", pair{"src-nonempty", srcNick}, pair{"profile", profNick})

	r.zero("identity.account-filled", `SELECT COUNT(*) FROM "user"
		WHERE (account IS NULL OR account = '') AND id IN (SELECT unnest($1::text[]))`, sys, r.ids)
	r.zero("identity.credential-orphans", `SELECT COUNT(*) FROM credential
		WHERE kind IN ('sms', 'wechat_mp') AND user_id NOT IN (SELECT unnest($1::text[]))`, sys, r.ids)
	r.zero("identity.profile-unmapped", `SELECT COUNT(*) FROM profile p
		WHERE NOT EXISTS (SELECT 1 FROM migration_map m WHERE m.entity = 'user' AND m.new_id = p.user_id)`, esp)
	r.zero("identity.ageband-null-iff-birth", `
		SELECT COUNT(*) FROM profile p JOIN migration_map m ON m.entity = 'user' AND m.new_id = p.user_id
		WHERE (p.age_band IS NULL) <> (p.birth_year IS NULL OR p.birth_month IS NULL)`, esp)

	// Age band: re-derived in pure SQL from the birth columns at the map's
	// own timestamp — the derivation instant the writer actually used.
	r.zero("identity.ageband-rederive", `
		WITH inst AS (
			SELECT m.new_id AS uid,
				EXTRACT(YEAR FROM m.created_at AT TIME ZONE '`+derivationTZ+`')::int AS y,
				EXTRACT(MONTH FROM m.created_at AT TIME ZONE '`+derivationTZ+`')::int AS mo
			FROM migration_map m WHERE m.entity = 'user'
		), aged AS (
			SELECT p.user_id,
				i.y - p.birth_year - CASE WHEN i.mo < p.birth_month THEN 1 ELSE 0 END AS age,
				p.age_band
			FROM profile p JOIN inst i ON i.uid = p.user_id
			WHERE p.birth_year IS NOT NULL AND p.birth_month IS NOT NULL
		)
		SELECT COUNT(*) FROM aged WHERE age_band <> CASE
			WHEN age < 12 THEN 'under12' WHEN age <= 14 THEN '12-14' WHEN age <= 17 THEN '15-17'
			WHEN age <= 22 THEN '18-22' WHEN age <= 25 THEN '23-25' WHEN age <= 29 THEN '26-29'
			WHEN age <= 40 THEN '30-40' ELSE 'over40' END`, esp)

	// Sex mapping, full column through Go: the vocabulary CASE mirrors the
	// writer's sexOf on every migrated row.
	srows, err := src.QueryContext(ctx, "SELECT id, wechat_sex FROM `user`")
	if err != nil {
		return fmt.Errorf("load src sex: %w", err)
	}
	wantSex := map[string]string{}
	for srows.Next() {
		var id int64
		var sex sql.NullInt64
		if err := srows.Scan(&id, &sex); err != nil {
			_ = srows.Close()
			return err
		}
		wantSex[strconv.FormatInt(id, 10)] = sexOf(sex)
	}
	_ = srows.Close()
	if err := srows.Err(); err != nil {
		return err
	}
	trows, err := esp.QueryContext(ctx, `SELECT user_id, sex FROM profile WHERE `+scopedUser, r.ids)
	if err != nil {
		return fmt.Errorf("load target sex: %w", err)
	}
	sexBad, sexChecked := 0, 0
	for trows.Next() {
		var uid, sex string
		if err := trows.Scan(&uid, &sex); err != nil {
			_ = trows.Close()
			return err
		}
		sexChecked++
		if wantSex[uid] != sex {
			sexBad++
		}
	}
	_ = trows.Close()
	if err := trows.Err(); err != nil {
		return err
	}
	r.check("identity.sex-map", sexBad == 0 && sexChecked == len(wantSex),
		fmt.Sprintf("rows=%d mismatched=%d", sexChecked, sexBad))

	return r.sampleIdentity(ctx)
}

// sampleIdentity compares every translated identity field on a random
// source sample against the mapped target rows.
func (r *reconciler) sampleIdentity(ctx context.Context) error {
	src, sys, esp := r.env.src, r.env.sys, r.env.esp

	rows, err := src.QueryContext(ctx,
		"SELECT "+userColumns+" FROM `user` ORDER BY RAND() LIMIT ?", r.sample)
	if err != nil {
		return fmt.Errorf("sample src users: %w", err)
	}
	var users []*srcUser
	for rows.Next() {
		u := &srcUser{}
		var isDeleted int64
		if err := rows.Scan(&u.ID, &u.Username, &u.Password, &u.PasswordAlgorithm, &u.Phone, &u.OpenID,
			&u.WechatSex, &u.WechatCountry, &u.WechatProvince, &u.WechatCity, &u.Nickname,
			&u.AgeRange, &u.BirthDate, &u.AvatarURL, &u.SchoolID, &u.Status, &isDeleted,
			&u.CreatedAt, &u.UpdatedAt); err != nil {
			_ = rows.Close()
			return err
		}
		u.IsDeleted = isDeleted != 0
		users = append(users, u)
	}
	_ = rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	sysBad, profBad, nickBad := 0, 0, 0
	for _, u := range users {
		id := strconv.FormatInt(u.ID, 10)

		var account, hash string
		var enabled bool
		var sCreated, sUpdated time.Time
		var sDeleted sql.NullTime
		err := sys.QueryRowContext(ctx, `
			SELECT account, COALESCE(password_hash, ''), enabled, created_at, updated_at, deleted_at
			FROM "user" WHERE id = $1`, id).
			Scan(&account, &hash, &enabled, &sCreated, &sUpdated, &sDeleted)
		if err != nil {
			return fmt.Errorf("sample system user %s: %w", id, err)
		}
		bcryptOK := u.PasswordAlgorithm == "BCRYPT" && u.Password.Valid && u.Password.String != ""
		bad := !sCreated.Equal(u.CreatedAt.UTC()) ||
			!sUpdated.Equal(u.UpdatedAt.UTC()) ||
			enabled != (!u.IsDeleted && u.Status == 1) ||
			(hash != "") != bcryptOK ||
			strings.TrimSpace(account) == ""
		if u.IsDeleted {
			bad = bad || !sDeleted.Valid || !sDeleted.Time.Equal(u.UpdatedAt.UTC())
		} else {
			bad = bad || sDeleted.Valid
		}
		if bad {
			sysBad++
		}

		var nick sql.NullString
		var pCreated, pUpdated time.Time
		var pDeleted sql.NullTime
		var birthY, birthM sql.NullInt64
		var avatar sql.NullString
		err = esp.QueryRowContext(ctx, `
			SELECT nickname, created_at, updated_at, deleted_at, birth_year, birth_month, avatar_url
			FROM profile WHERE user_id = $1`, id).
			Scan(&nick, &pCreated, &pUpdated, &pDeleted, &birthY, &birthM, &avatar)
		if err != nil {
			return fmt.Errorf("sample profile %s: %w", id, err)
		}
		avatarWant := nullStr(u.AvatarURL)
		avatarGot := ""
		if avatar.Valid {
			avatarGot = strings.TrimSpace(avatar.String)
		}
		pbad := !pCreated.Equal(u.CreatedAt.UTC()) ||
			!pUpdated.Equal(u.UpdatedAt.UTC()) ||
			u.IsDeleted != pDeleted.Valid ||
			birthY.Valid != u.BirthDate.Valid ||
			avatarGot != avatarWant
		if u.BirthDate.Valid {
			pbad = pbad || !birthY.Valid ||
				int(birthY.Int64) != u.BirthDate.Time.Year() ||
				int(birthM.Int64) != int(u.BirthDate.Time.Month())
		}
		if pbad {
			profBad++
		}

		// Nickname: trimmed, rune-capped at 12.
		wantNick := ""
		if trimmed := strings.TrimSpace(u.Nickname); trimmed != "" {
			runes := []rune(trimmed)
			if len(runes) > 12 {
				runes = runes[:12]
			}
			wantNick = string(runes)
		}
		gotNick := ""
		if nick.Valid {
			gotNick = nick.String
		}
		if wantNick != gotNick {
			nickBad++
		}
	}
	r.check("identity.sample-system", sysBad == 0, fmt.Sprintf("%d sampled, %d mismatched", len(users), sysBad))
	r.check("identity.sample-profile", profBad == 0, fmt.Sprintf("%d sampled, %d mismatched", len(users), profBad))
	r.check("identity.sample-nickname", nickBad == 0, fmt.Sprintf("%d sampled, %d mismatched", len(users), nickBad))
	return nil
}

func (r *reconciler) entitlementFamily(ctx context.Context) error {
	src, esp := r.env.src, r.env.esp

	srcUnlocked := r.count("ent.membership-scope", src,
		"SELECT COUNT(*) FROM `user` WHERE is_deleted = 0 AND is_unlocked = 1")
	tgtMembership := r.count("ent.membership-scope", esp,
		`SELECT COUNT(*) FROM entitlement WHERE scope_kind = 'membership' AND scope_ref IS NULL AND `+scopedUser, r.ids)
	r.identity("ent.membership-scope", pair{"src-unlocked-live", srcUnlocked}, pair{"entitlement", tgtMembership})

	srcWindowed := r.count("ent.membership-windowed", src,
		"SELECT COUNT(*) FROM `user` WHERE is_deleted = 0 AND is_unlocked = 1 AND unlock_expires_at IS NOT NULL")
	tgtWindowed := r.count("ent.membership-windowed", esp,
		`SELECT COUNT(*) FROM entitlement WHERE scope_kind = 'membership' AND expires_at IS NOT NULL AND `+scopedUser, r.ids)
	r.identity("ent.membership-windowed", pair{"src-window", srcWindowed}, pair{"expires-set", tgtWindowed})

	for _, fam := range []struct{ col, name string }{
		{"unlocked_cognitive", familyCognitive},
		{"unlocked_career", familyCareer},
	} {
		srcFlag := r.count("ent.family-"+fam.name, src,
			"SELECT COUNT(*) FROM `user` WHERE is_deleted = 0 AND "+fam.col+" = 1")
		tgtFlag := r.count("ent.family-"+fam.name, esp,
			`SELECT COUNT(*) FROM entitlement WHERE scope_kind = 'report_type' AND scope_ref = $2 AND `+scopedUser, r.ids, fam.name)
		r.identity("ent.family-"+fam.name, pair{"src-flag", srcFlag}, pair{"entitlement", tgtFlag})
	}

	srcPayment := r.count("ent.payment-sentinel", src,
		"SELECT COUNT(*) FROM `user` WHERE is_deleted = 0 AND is_unlocked = 1 AND unlock_method = 'payment'")
	tgtPayment := r.count("ent.payment-sentinel", esp,
		`SELECT COUNT(*) FROM entitlement WHERE scope_kind = 'membership' AND source = $2 AND `+scopedUser, r.ids, sourceLegacyPayment)
	r.identity("ent.payment-sentinel", pair{"src-payment", srcPayment}, pair{"source=legacy_payment", tgtPayment})

	r.zero("ent.family-school-none", `SELECT COUNT(*) FROM entitlement WHERE scope_ref = 'school' AND `+scopedUser, esp, r.ids)
	r.zero("ent.flag-sentinel", `SELECT COUNT(*) FROM entitlement
		WHERE scope_kind = 'report_type' AND source <> $2 AND `+scopedUser, esp, r.ids, sourceLegacy)
	r.zero("ent.dead-user-grants", `SELECT COUNT(*) FROM entitlement e
		JOIN profile p ON p.user_id = e.user_id
		WHERE p.deleted_at IS NOT NULL AND e.`+scopedUser, esp, r.ids)
	r.zero("ent.status-active", `SELECT COUNT(*) FROM entitlement WHERE status <> 'active' AND `+scopedUser, esp, r.ids)

	// Codes: the kept set re-derived in source SQL, set-compared against the
	// target's claimed rows — both directions, so a dropped code leaking in
	// or a kept code missing both fail.
	kept, err := sourceKeptCodes(ctx, src)
	if err != nil {
		return err
	}
	rows, err := esp.QueryContext(ctx, `SELECT code FROM activation_code
		WHERE id IN (SELECT new_id FROM migration_map WHERE entity = 'activation_code')`)
	if err != nil {
		return err
	}
	targetCodes := map[string]struct{}{}
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			_ = rows.Close()
			return err
		}
		targetCodes[c] = struct{}{}
	}
	_ = rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	keptSet := map[string]struct{}{}
	for _, c := range kept {
		keptSet[c] = struct{}{}
	}
	missing, leaked := 0, 0
	for c := range keptSet {
		if _, ok := targetCodes[c]; !ok {
			missing++
		}
	}
	for c := range targetCodes {
		if _, ok := keptSet[c]; !ok {
			leaked++
		}
	}
	r.check("ent.codes-kept", missing == 0 && leaked == 0 && len(targetCodes) == len(keptSet),
		fmt.Sprintf("src-kept=%d target=%d missing=%d leaked=%d", len(keptSet), len(targetCodes), missing, leaked))

	// Sentinel expiry: legacy NULL expiry wrote the far-future sentinel.
	srcNullExpiry := r.count("ent.codes-sentinel-expiry", src,
		keptCodeCount("expires_at IS NULL"))
	tgtSentinel := r.count("ent.codes-sentinel-expiry", esp,
		`SELECT COUNT(*) FROM activation_code WHERE expires_at = $1
		AND id IN (SELECT new_id FROM migration_map WHERE entity = 'activation_code')`, farFuture)
	r.identity("ent.codes-sentinel-expiry", pair{"src-null-expiry", srcNullExpiry}, pair{"sentinel", tgtSentinel})

	// Redeem trail: used>0 and exactly one live record naming the code.
	srcTrail := r.count("ent.codes-redeem-trail", src, keptCodeCount(`used_count > 0 AND
		(SELECT COUNT(*) FROM activation_records rr
			WHERE rr.activation_code = activation_code.code AND rr.is_deleted = 0) = 1`))
	tgtTrail := r.count("ent.codes-redeem-trail", esp,
		`SELECT COUNT(*) FROM activation_code WHERE redeemed_by IS NOT NULL
		AND id IN (SELECT new_id FROM migration_map WHERE entity = 'activation_code')`)
	r.identity("ent.codes-redeem-trail", pair{"src-single-record", srcTrail}, pair{"redeemed", tgtTrail})

	return r.sampleCodes(ctx)
}

// keptCodeWhere re-derives the code-keep rules in source SQL: live,
// family-typed, not exhausted, not expired. The NOW() boundary matches the
// writer's plan-time comparison — a code expiring between the phases shows
// up as a one-code drift the operator resolves.
const keptCodeWhere = `WHERE is_deleted = 0 AND report_type IS NOT NULL AND report_type <> ''
		AND used_count < quantity AND (expires_at IS NULL OR expires_at > NOW())`

func keptCodeCount(cond string) string {
	return "SELECT COUNT(*) FROM activation_code " + keptCodeWhere + " AND " + cond
}

// sourceKeptCodes re-derives the kept code list under the same rules.
func sourceKeptCodes(ctx context.Context, src *sql.DB) ([]string, error) {
	rows, err := src.QueryContext(ctx, "SELECT code FROM activation_code "+keptCodeWhere)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *reconciler) sampleCodes(ctx context.Context) error {
	src, esp := r.env.src, r.env.esp

	rows, err := src.QueryContext(ctx, `SELECT code, report_type, quantity, used_count, expires_at, created_by, created_at
		FROM activation_code `+keptCodeWhere+` ORDER BY RAND() LIMIT ?`, r.sample)
	if err != nil {
		return err
	}
	type codeRow struct {
		code                      string
		rtype                     sql.NullString
		quantity, used, createdBy int64
		expires                   sql.NullTime
		createdAt                 time.Time
	}
	var sample []codeRow
	for rows.Next() {
		var c codeRow
		if err := rows.Scan(&c.code, &c.rtype, &c.quantity, &c.used, &c.expires, &c.createdBy, &c.createdAt); err != nil {
			_ = rows.Close()
			return err
		}
		sample = append(sample, c)
	}
	_ = rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	if len(sample) == 0 {
		r.check("ent.codes-sample", true, "no kept codes to sample")
		return nil
	}

	bad := 0
	for _, c := range sample {
		var rtype string
		var quantity, used int64
		var expires sql.NullTime
		var createdBy string
		var createdAt time.Time
		var batch sql.NullString
		err := esp.QueryRowContext(ctx, `
			SELECT report_type, quantity, used_count, expires_at, created_by, created_at, batch_id
			FROM activation_code WHERE code = $1`, c.code).
			Scan(&rtype, &quantity, &used, &expires, &createdBy, &createdAt, &batch)
		if err != nil {
			return fmt.Errorf("sample code %s: %w", c.code, err)
		}
		miss := rtype != nullStr(c.rtype) || quantity != c.quantity || used != c.used ||
			createdBy != strconv.FormatInt(c.createdBy, 10) || !createdAt.Equal(c.createdAt.UTC()) ||
			batch.Valid
		if c.expires.Valid {
			miss = miss || !expires.Valid || !expires.Time.Equal(c.expires.Time.UTC())
		} else {
			miss = miss || !expires.Valid || !expires.Time.Equal(farFuture)
		}
		if miss {
			bad++
		}
	}
	r.check("ent.codes-sample", bad == 0, fmt.Sprintf("%d sampled, %d mismatched", len(sample), bad))
	return nil
}

func (r *reconciler) behaviorFamily(ctx context.Context) error {
	src, esp := r.env.src, r.env.esp

	srcRecords := r.count("beh.record-count", src,
		"SELECT COUNT(*) FROM test_record WHERE test_code <> ?", retiredTestCode)
	mapRecords := r.count("beh.record-count", esp,
		`SELECT COUNT(*) FROM migration_map WHERE entity = $1`, entityTestRecord)
	tgtRecords := r.count("beh.record-count", esp,
		`SELECT COUNT(*) FROM test_record WHERE id IN (SELECT new_id FROM migration_map WHERE entity = 'test_record')`)
	r.identity("beh.record-count",
		pair{"src-minus-retired", srcRecords}, pair{"map", mapRecords}, pair{"target", tgtRecords})

	// Per-code counts, both stores grouped.
	srcByCode := r.groupByCode(ctx, src, "SELECT test_code, COUNT(*) FROM test_record WHERE test_code <> ? GROUP BY test_code", retiredTestCode)
	tgtByCode := r.groupByCode(ctx, esp, `SELECT test_code, COUNT(*) FROM test_record
		WHERE id IN (SELECT new_id FROM migration_map WHERE entity = 'test_record') GROUP BY test_code`)
	codeOK := len(srcByCode) == len(tgtByCode)
	for code, n := range srcByCode {
		if tgtByCode[code] != n {
			codeOK = false
		}
	}
	r.check("beh.record-by-code", codeOK,
		fmt.Sprintf("src %d codes / target %d codes", len(srcByCode), len(tgtByCode)))

	srcDead := r.count("beh.deleted-carried", src,
		"SELECT COUNT(*) FROM test_record WHERE test_code <> ? AND is_deleted = 1", retiredTestCode)
	tgtDead := r.count("beh.deleted-carried", esp,
		`SELECT COUNT(*) FROM test_record WHERE deleted_at IS NOT NULL
		AND id IN (SELECT new_id FROM migration_map WHERE entity = 'test_record')`)
	r.identity("beh.deleted-carried", pair{"src-deleted", srcDead}, pair{"target-deleted", tgtDead})

	// Store invariant, full column: grade is a function of the score.
	r.zero("beh.grade-invariant", `SELECT COUNT(*) FROM test_record
		WHERE authoritative_score IS NOT NULL AND grade <> CASE
			WHEN authoritative_score >= 95 THEN 'S' WHEN authoritative_score >= 85 THEN 'A'
			WHEN authoritative_score >= 60 THEN 'B' ELSE 'C' END`, esp)
	r.zero("beh.record-orphans", `SELECT COUNT(*) FROM test_record tr
		WHERE NOT EXISTS (SELECT 1 FROM migration_map m WHERE m.entity = 'test_record' AND m.new_id = tr.id)`, esp)

	// Fact-stream routing: '{}' rows are exactly the no-facts rows.
	noFacts := "(detail_data IS NULL OR TRIM(detail_data) IN ('', '{}', 'null')) AND (summary_data IS NULL OR TRIM(summary_data) IN ('', '{}', 'null'))"
	srcNoFacts := r.count("beh.fact-empty", src,
		"SELECT COUNT(*) FROM test_record WHERE test_code <> ? AND "+noFacts, retiredTestCode)
	tgtEmpty := r.count("beh.fact-empty", esp,
		`SELECT COUNT(*) FROM test_record WHERE raw_detail = '{}'::jsonb
		AND id IN (SELECT new_id FROM migration_map WHERE entity = 'test_record')`)
	r.identity("beh.fact-empty", pair{"src-no-facts", srcNoFacts}, pair{"target-{}", tgtEmpty})

	if err := r.checkOrdinals(ctx); err != nil {
		return err
	}
	if err := r.checkGames(ctx); err != nil {
		return err
	}
	return r.sampleRecords(ctx)
}

func (r *reconciler) groupByCode(ctx context.Context, db *sql.DB, q string, args ...any) map[string]int64 {
	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		r.check("beh.record-by-code", false, "query error: "+err.Error())
		return nil
	}
	out := map[string]int64{}
	for rows.Next() {
		var code string
		var n int64
		if err := rows.Scan(&code, &n); err != nil {
			_ = rows.Close()
			r.check("beh.record-by-code", false, "scan error: "+err.Error())
			return nil
		}
		out[code] = n
	}
	_ = rows.Close()
	if err := rows.Err(); err != nil {
		r.check("beh.record-by-code", false, "rows error: "+err.Error())
		return nil
	}
	return out
}

// checkOrdinals re-derives every attempt ordinal from the exact source
// ordering the writer used (user, test, completed_at with NULLs first, id)
// and compares against the target column — full column, no sampling.
func (r *reconciler) checkOrdinals(ctx context.Context) error {
	src, esp := r.env.src, r.env.esp

	rows, err := src.QueryContext(ctx, `SELECT id, user_id, test_code, completed_at
		FROM test_record WHERE test_code <> ? ORDER BY user_id, test_code, completed_at, id`, retiredTestCode)
	if err != nil {
		return err
	}
	type ordKey struct{ user, code string }
	want := map[string]int64{}
	seen := map[ordKey]int64{}
	for rows.Next() {
		var id, user int64
		var code string
		var completed sql.NullTime
		if err := rows.Scan(&id, &user, &code, &completed); err != nil {
			_ = rows.Close()
			return err
		}
		k := ordKey{strconv.FormatInt(user, 10), code}
		seen[k]++
		want[strconv.FormatInt(id, 10)] = seen[k]
	}
	_ = rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	trows, err := esp.QueryContext(ctx, `SELECT id, attempt_no FROM test_record
		WHERE id IN (SELECT new_id FROM migration_map WHERE entity = 'test_record')`)
	if err != nil {
		return err
	}
	bad, checked := 0, 0
	for trows.Next() {
		var id string
		var n int64
		if err := trows.Scan(&id, &n); err != nil {
			_ = trows.Close()
			return err
		}
		checked++
		if want[id] != n {
			bad++
		}
	}
	_ = trows.Close()
	if err := trows.Err(); err != nil {
		return err
	}
	r.check("beh.ordinal-rederive", bad == 0 && checked == len(want),
		fmt.Sprintf("rows=%d mismatched=%d", checked, bad))
	return nil
}

// checkGames pins the preference fold: every user with a live preference
// row owns a games array, and its length re-derives in source SQL through
// the dedup identities (catalog by id, custom by lowercased name).
func (r *reconciler) checkGames(ctx context.Context) error {
	src, esp := r.env.src, r.env.esp

	srcUsers := r.count("beh.games-coverage", src,
		"SELECT COUNT(DISTINCT user_id) FROM user_game_preference WHERE is_deleted = 0")
	tgtGames := r.count("beh.games-coverage", esp,
		`SELECT COUNT(*) FROM profile WHERE games IS NOT NULL AND `+scopedUser, r.ids)
	r.identity("beh.games-coverage", pair{"src-pref-users", srcUsers}, pair{"profile-games", tgtGames})

	rows, err := src.QueryContext(ctx, `SELECT user_id,
		COUNT(DISTINCT CASE WHEN game_id IS NOT NULL AND game_id <> 0 THEN game_id END)
		+ COUNT(DISTINCT CASE WHEN (game_id IS NULL OR game_id = 0) AND TRIM(game_name) <> '' THEN LOWER(TRIM(game_name)) END)
		FROM user_game_preference WHERE is_deleted = 0 GROUP BY user_id`)
	if err != nil {
		return err
	}
	wantLen := map[string]int64{}
	for rows.Next() {
		var user, n int64
		if err := rows.Scan(&user, &n); err != nil {
			_ = rows.Close()
			return err
		}
		wantLen[strconv.FormatInt(user, 10)] = n
	}
	_ = rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	trows, err := esp.QueryContext(ctx, `SELECT user_id, jsonb_array_length(games) FROM profile
		WHERE games IS NOT NULL AND `+scopedUser, r.ids)
	if err != nil {
		return err
	}
	bad, checked := 0, 0
	for trows.Next() {
		var user string
		var n int64
		if err := trows.Scan(&user, &n); err != nil {
			_ = trows.Close()
			return err
		}
		checked++
		if wantLen[user] != n {
			bad++
		}
	}
	_ = trows.Close()
	if err := trows.Err(); err != nil {
		return err
	}
	r.check("beh.games-length", bad == 0 && checked == len(wantLen),
		fmt.Sprintf("users=%d mismatched=%d", checked, bad))
	return nil
}

// sampleRecords compares the carried payload and wall clocks on a random
// sample: raw_detail against the routed fact stream (canonical JSON), the
// millisecond epochs against the target timestamps.
func (r *reconciler) sampleRecords(ctx context.Context) error {
	src, esp := r.env.src, r.env.esp

	rows, err := src.QueryContext(ctx, `SELECT id, score, raw_score, summary_data, detail_data,
		start_time, end_time, completed_at, created_at FROM test_record
		WHERE test_code <> ? ORDER BY RAND() LIMIT ?`, retiredTestCode, r.sample)
	if err != nil {
		return err
	}
	type recRow struct {
		id              int64
		score           float64
		rawScore        sql.NullFloat64
		summary, detail []byte
		start, end      int64
		completed       sql.NullTime
		createdAt       time.Time
	}
	var sample []recRow
	for rows.Next() {
		var x recRow
		if err := rows.Scan(&x.id, &x.score, &x.rawScore, &x.summary, &x.detail,
			&x.start, &x.end, &x.completed, &x.createdAt); err != nil {
			_ = rows.Close()
			return err
		}
		sample = append(sample, x)
	}
	_ = rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	payloadBad, timeBad := 0, 0
	for _, x := range sample {
		id := strconv.FormatInt(x.id, 10)
		var rawDetail []byte
		var score float64
		var rawScore sql.NullFloat64
		var started, ended, createdAt time.Time
		err := esp.QueryRowContext(ctx, `
			SELECT raw_detail::text, authoritative_score, raw_score, started_at, ended_at, created_at
			FROM test_record WHERE id = $1`, id).
			Scan(&rawDetail, &score, &rawScore, &started, &ended, &createdAt)
		if err != nil {
			return fmt.Errorf("sample record %s: %w", id, err)
		}

		wantPayload := []byte("{}")
		switch {
		case hasFacts(x.detail):
			wantPayload = x.detail
		case hasFacts(x.summary):
			wantPayload = x.summary
		}
		if !jsonEqual(wantPayload, rawDetail) || score != x.score ||
			rawScore.Valid != x.rawScore.Valid || (x.rawScore.Valid && rawScore.Float64 != x.rawScore.Float64) {
			payloadBad++
		}

		wantCreated := x.createdAt.UTC()
		if x.completed.Valid {
			wantCreated = x.completed.Time.UTC()
		}
		if !started.Equal(time.UnixMilli(x.start).UTC()) ||
			!ended.Equal(time.UnixMilli(x.end).UTC()) || !createdAt.Equal(wantCreated) {
			timeBad++
		}
	}
	r.check("beh.sample-payload", payloadBad == 0, fmt.Sprintf("%d sampled, %d mismatched", len(sample), payloadBad))
	r.check("beh.sample-times", timeBad == 0, fmt.Sprintf("%d sampled, %d mismatched", len(sample), timeBad))
	return nil
}

// jsonEqual compares two JSON payloads canonically: whitespace and key
// order never mask a real difference, any real difference always fails.
func jsonEqual(a, b []byte) bool {
	var av, bv any
	if err := json.Unmarshal(a, &av); err != nil {
		return false
	}
	if err := json.Unmarshal(b, &bv); err != nil {
		return false
	}
	ak, _ := json.Marshal(av)
	bk, _ := json.Marshal(bv)
	return string(ak) == string(bk)
}
