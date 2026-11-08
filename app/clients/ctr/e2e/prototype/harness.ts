import { execSync } from "node:child_process";
import { readFileSync } from "node:fs";
import { expect, type Page } from "@playwright/test";
import { normalizeLines } from "./text.mjs";

// The shared machinery of the prototype-fidelity specs: member seeding
// against the live dev stack (CTR_E2E=1), the namespaced session login,
// and the two-sided face collection/assertion. Home owns the frozen-face
// precedent; every later family (profile, report) rides the same laws:
// the oracle is the machine snapshot, never the implementation, and a
// standing ledgered divergence is a red line, not an adjusted test.

const SYSTEM_HTTP = "http://localhost:11001";
const PG = "docker exec -i cyber-ecosystem-postgresql-1 psql -U postgres -v ON_ERROR_STOP=1";
// The admin row's argon2id digest (password admin123456) — copied, not
// minted: a static literal keeps the seed deterministic with zero crypto
// in the runner; each member still owns its own rows.
const PASSWORD_HASH =
  "$argon2id$v=19$m=19456,t=2,p=1$YuFiYhO+1nNpBgTy6KxQUg$s3/toZlfTIuDY6DEMkPy8L3e6Suvm87RRXU1ivZTKrQ";

// The fps lineup the real draw freezes (match_config render_order with
// the catalog's code picks — the fullchain run's narrowed reaction slot).
const FPS_STEPS = [
  { ability: "reactionSpeed", test_code: "red_dot" },
  { ability: "aimingAccuracy", test_code: "flash_aiming" },
  { ability: "trackingStability", test_code: "moving_target" },
  { ability: "skillReleaseAccuracy", test_code: "skill_trajectory" },
];

// The moba lineup off the same source (match_config render_order
// 1-4: resourceAllocation, decisionEfficiency, reactionSpeed,
// skillReleaseAccuracy) with the catalog's own code picks — the
// member-overview preset's second session rides it.
const MOBA_STEPS = [
  { ability: "resourceAllocation", test_code: "character_counter" },
  { ability: "decisionEfficiency", test_code: "offense_defense" },
  { ability: "reactionSpeed", test_code: "red_dot" },
  { ability: "skillReleaseAccuracy", test_code: "skill_trajectory" },
];

// The cohort pools behind the tested preset's live percentiles. The
// engine benches each best against the pool of same-birth-year-same-sex
// authoritative records (the member's own record joins its own pool)
// through the z-CDF — pct = round(100·Φ((score−μ)/σ)) — so every pool
// below is engineered to land its ability's frozen percentile exactly,
// with a safe margin off the rounding edge (float noise cannot flip
// it). Ten bench rows plus the member's own clear the pool floor
// (cohortFloor 10) and stay under the medium-confidence ceiling
// (cohortCeil 50) — three report faces read these live: the digest's
// per-ability percentiles, the cognitive page's weighted dimensions
// (an ability with no pool answer sinks its whole dimension), and the
// locked single report's live judgement.
const COHORT_POOLS: Record<string, number[]> = {
  // reactionSpeed · member 75 → 78 (raw 78.01)
  red_dot: [40, 62, 63, 68, 69, 71, 72, 73, 73, 74],
  // aimingAccuracy · member 82 → 86 (raw 86.03)
  flash_aiming: [38, 45, 52, 59, 59, 66, 73, 73, 80, 87],
  // trackingStability · member 69 → 71 (raw 71.02)
  moving_target: [26, 35, 44, 53, 53, 62, 62, 71, 80, 89],
  // skillReleaseAccuracy · member 63 → 64 (raw 63.87)
  skill_trajectory: [32, 39, 46, 53, 53, 60, 67, 67, 74, 81],
};
// One id suffix per pool record so ten bench users can hold four rows
// each without id collisions (and cleanup's user_id catch takes all).
const POOL_SUFFIX: Record<string, string> = {
  red_dot: "r",
  flash_aiming: "f",
  moving_target: "m",
  skill_trajectory: "t",
};

// The scored records' grade letters, by percentile band — the band the
// engine's quantile-frozen curves actually grade (the authoritative
// number reads as percentile position), so the letter a real run
// reproducing this face would write follows the percentile, not the
// synthetic 0.9×pct+5 score.
const RECORD_GRADES = ["B", "A", "B", "B"];

export interface Fixture {
  page: string;
  preset: string | null;
  source: string;
  title: string;
  lines: string[];
  placeholders: string[];
}

export function fixtureOf(page: string, key: string): Fixture {
  return JSON.parse(
    readFileSync(new URL(`./fixtures/${page}__${key}.json`, import.meta.url), "utf8"),
  ) as Fixture;
}

export function sql(db: "system" | "esports", statements: string): void {
  execSync(`${PG} -d ${db}`, { input: statements });
}

/** Seeds one member's system rows — the account handle with its member
 * seat (the D-BG birth law: rows, not RPC) and, for the overview face,
 * the sms + wechat credentials the prototype's READY() member carries.
 * The digest's $…$ segments make stdin mandatory: a double-quoted -c
 * would eat them as shell variables. */
export function seedBareMember(
  tag: string,
  opts: { phone?: string; wechat?: string; passwordless?: boolean } = {},
): { account: string; userId: string } {
  const userId = `${tag}u`;
  const account = `${tag}a`;
  const credentials =
    opts.phone === undefined && opts.wechat === undefined
      ? ""
      : `INSERT INTO credential (id, created_at, updated_at, kind, identifier, user_id)
         VALUES ${[
           opts.phone !== undefined
             ? `('${tag}cs', now(), now(), 'sms', '${opts.phone}', '${userId}')`
             : null,
           opts.wechat !== undefined
             ? `('${tag}cw', now(), now(), 'wechat_mp', '${opts.wechat}', '${userId}')`
             : null,
         ]
           .filter((row): row is string => row !== null)
           .join(",")};`;
  // passwordless mints the merge flow's empty side: the mergeable bar
  // (system half) demands no password and no second credential — the
  // absorbed account's whole shape. The column is NOT NULL, and the
  // reader's own law (Deref(hash, "") != "") names the no-password
  // storage shape: the empty string, the wechat-born member's row.
  // Such a member cannot log in through this harness (the login leg is
  // account+password); only the preview ever reads it, never a session.
  const passwordHash = opts.passwordless === true ? "''" : `'${PASSWORD_HASH}'`;
  sql(
    "system",
    `INSERT INTO "user" (id, created_at, updated_at, account, password_hash, enabled)
     VALUES ('${userId}', now(), now(), '${account}', ${passwordHash}, true);
     INSERT INTO membership (id, created_at, updated_at, user_id, tenant_id, population)
     VALUES ('${tag}m', now(), now(), '${userId}', 'default', 'member');
     ${credentials}`,
  );
  return { account, userId };
}

/** Seeds the esports profile row on a bare member. The shapes the presets
 * froze: "ready" (the complete member), "profile-missing" (nickname and a
 * two-game preference, no birth — exactly the wizard's subject), and
 * "underage" (birth 2013-04 — the pay wall's own refusal subject at the
 * pinned 2026-09 clock, 13 by year+month). Sex rides the real storage
 * word, never the prototype's display glyph. */
export function seedProfile(
  tag: string,
  userId: string,
  shape: "ready" | "profile-missing" | "underage",
): void {
  const games =
    shape === "profile-missing"
      ? "('王者荣耀','CS2')"
      : shape === "underage"
        ? "('和平精英')"
        : "('王者荣耀','英雄联盟','CS2')";
  // Birth year 2004, not the prototype's 2008: the cohort key is birth
  // year + sex, and the shared dev pool carries a legacy 2008 male
  // calibration batch (another lane's residue, left in place) that a 2008
  // member would absorb — the frozen percentile needs a hermetic pool.
  // 2004-06 is 22 at the fixture's own moment, so the stored 18-22 band
  // stays honest, and no oracle line quotes the birth. The underage shape
  // carries the prototype's own 2013-04 — its cohort key meets no pool
  // (the wall face reads no percentile at all).
  const birth =
    shape === "ready"
      ? "2004, 6, 'male', '18-22'"
      : shape === "underage"
        ? "2013, 4, 'male', '12-14'"
        : "NULL, NULL, 'male', NULL";
  sql(
    "esports",
    `INSERT INTO profile (id, created_at, updated_at, tenant_id, user_id, nickname, birth_year, birth_month, sex, age_band, games)
     VALUES ('${tag}p', now(), now(), 'default', '${userId}', '龘好久', ${birth},
             -- games entries are catalog refs keyed by the game CODE — the
             -- namespace ListGames exposes and UpdateProfile's
             -- ExistingCodes gate writes (the ent row PK is a different
             -- namespace the wizard's pre-pick never matches). Resolved by
             -- name against the seeded catalog, never hardcoded.
             (SELECT jsonb_agg(jsonb_build_object('game_id', code)) FROM game WHERE name IN ${games}));`,
  );
}

/** The one-call member: system rows plus the profile shape. */
export function seedMember(
  tag: string,
  shape: "ready" | "profile-missing" | "underage" = "ready",
  creds: { phone?: string; wechat?: string } = {},
): { account: string; userId: string } {
  const { account, userId } = seedBareMember(tag, creds);
  seedProfile(tag, userId, shape);
  return { account, userId };
}

/** The resume preset's unfinished session: four frozen steps, two of them
 * carrying spent-but-invalid attempts — invalid on purpose, the only legal
 * row shape that keeps both truths the fixture froze: the steps read done
 * (any stored record marks its step done) while the session never counts
 * as a completed one (valid attempts are what the counter joins). */
export function seedResumeSession(tag: string, userId: string): void {
  const sessionId = `${tag}s`;
  sql(
    "esports",
    `INSERT INTO match_session (id, created_at, updated_at, tenant_id, user_id, genre, difficulty, seed, steps, retest_budget, status)
     VALUES ('${sessionId}', '2026-09-11 21:04:00+08', '2026-09-11 21:04:00+08', 'default', '${userId}', 'fps', 'experience', 1,
             '${JSON.stringify(FPS_STEPS)}'::jsonb, 3, 'active');
     INSERT INTO test_record (id, created_at, updated_at, tenant_id, test_code, user_id, attempt_no, session_id, started_at, ended_at, raw_detail, valid, invalid_reason)
     VALUES ('${tag}r1', '2026-09-11 21:09:30+08', '2026-09-11 21:09:30+08', 'default', 'red_dot', '${userId}', 1, '${sessionId}',
             '2026-09-11 21:08:00+08', '2026-09-11 21:09:30+08', '{}', false, 'BACKGROUND_OVER_BUDGET'),
            ('${tag}r2', '2026-09-11 21:14:10+08', '2026-09-11 21:14:10+08', 'default', 'flash_aiming', '${userId}', 1, '${sessionId}',
             '2026-09-11 21:12:40+08', '2026-09-11 21:14:10+08', '{}', false, 'BACKGROUND_OVER_BUDGET');`,
  );
}

/** The tested-fps preset's scored session: four authoritative attempts
 * (the prototype's rec() score semantics, each with its band letter), the
 * frozen snapshot that owns the verdict face (tier B, composite 75, the
 * four percentiles), and the four ten-row cohort pools that wake the
 * live percentiles to 78 / 86 / 71 / 64 — the digest's, the cognitive
 * dimensions', and the locked single face's read-time answers. Returns
 * the anchor record id the single report's route names. */
export function seedScoredSession(tag: string, userId: string): { anchorId: string } {
  const sessionId = `${tag}s`;
  const scores: Record<string, number> = {
    red_dot: 75,
    flash_aiming: 82,
    moving_target: 69,
    skill_trajectory: 63,
  };
  const times = ["09:12:10", "09:13:30", "09:15:00", "09:16:20"];
  const records = FPS_STEPS.map((step, i) => {
    const at = `2026-09-09 ${times[i]}+08`;
    return `('${tag}r${i + 1}', '${at}', '${at}', 'default', '${step.test_code}', '${userId}', 1, '${sessionId}',
            '${at}', '${at}', '{}', ${scores[step.test_code]}, ${scores[step.test_code]}, '${RECORD_GRADES[i]}', true, NULL)`;
  });
  const abilities = FPS_STEPS.map((step, i) => {
    const pct = [78, 86, 71, 64][i];
    const best = scores[step.test_code];
    return `{"Ability":"${step.ability}","GameCode":"${step.test_code}","Best":${best},"Attempts":[${best}],"BestGrade":null,"Percentile":${pct},"PrevSession":null,"HistoryBest":null}`;
  });
  // The anchor is the latest scored record; the snapshot freezes on it.
  const anchorId = `${tag}r4`;
  // The ten bench users are the red_dot pool's own rows (any pool's
  // length names them); the ?? only satisfies the index-access check.
  const benchUsers = COHORT_POOLS.red_dot ?? [];
  const cohortProfiles = benchUsers.map(
    (_, i) =>
      `('${tag}c${i}p', now(), now(), 'default', '${tag}c${i}', NULL, 2004, 6, 'male', '18-22', NULL)`,
  );
  const cohortRecords = Object.entries(COHORT_POOLS).flatMap(([code, pool]) =>
    pool.map(
      (score, i) =>
        `('${tag}c${i}${POOL_SUFFIX[code]}', now(), now(), 'default', '${code}', '${tag}c${i}', 1, NULL, now(), now(), '{}', ${score}, ${score}, true, NULL)`,
    ),
  );
  sql(
    "esports",
    `INSERT INTO match_session (id, created_at, updated_at, tenant_id, user_id, genre, difficulty, seed, steps, retest_budget, status)
     VALUES ('${sessionId}', '2026-09-09 09:12:00+08', '2026-09-09 09:12:00+08', 'default', '${userId}', 'fps', 'experience', 1,
             '${JSON.stringify(FPS_STEPS)}'::jsonb, 3, 'active');
     INSERT INTO test_record (id, created_at, updated_at, tenant_id, test_code, user_id, attempt_no, session_id, started_at, ended_at, raw_detail, raw_score, authoritative_score, grade, valid, invalid_reason)
     VALUES ${records.join(",")};
     INSERT INTO report_snapshot (id, created_at, updated_at, tenant_id, record_id, user_id, tier, composite_percentile, abilities, cohort_basis)
     VALUES ('${tag}sn', now(), now(), 'default', '${anchorId}', '${userId}', 'B', 75, '[${abilities.join(",")}]'::jsonb, 'same-age-same-sex');
     INSERT INTO profile (id, created_at, updated_at, tenant_id, user_id, nickname, birth_year, birth_month, sex, age_band, games)
     VALUES ${cohortProfiles.join(",")};
     INSERT INTO test_record (id, created_at, updated_at, tenant_id, test_code, user_id, attempt_no, session_id, started_at, ended_at, raw_detail, raw_score, authoritative_score, valid, invalid_reason)
     VALUES ${cohortRecords.join(",")};`,
  );
  return { anchorId };
}

/** Seeds one in-flight fps session with its attempt rows — the test
 * family's own substrate (prototype preset resume: the lineup's two
 * played games, their attempts, and the retake budget the view derives
 * as budget-minus-used). No snapshot and no cohort pool ride this seed:
 * every face it feeds (guide / intro / running / the review result)
 * renders no percentile line, so the live branch's null cohort answer is
 * invisible. The done face wants the frozen branch instead — pass the
 * snapshot and the session report reads its percentiles verbatim, again
 * pool-free. */
export function seedTestSession(
  tag: string,
  userId: string,
  o: {
    startedAt: string;
    records: {
      suffix: string;
      code: string;
      attempt: number;
      score: number;
      grade?: string;
      at: string;
    }[];
    snapshot?: { anchorSuffix: string; percentiles: number[] };
  },
): { sessionId: string } {
  const sessionId = `${tag}s`;
  const rows = o.records.map(
    (r) =>
      `('${tag}${r.suffix}', '${r.at}', '${r.at}', 'default', '${r.code}', '${userId}', ${r.attempt}, '${sessionId}',
              '${r.at}', '${r.at}', '{}', ${r.score}, ${r.score}, ${r.grade === undefined ? "NULL" : `'${r.grade}'`}, true, NULL)`,
  );
  // The IIFE narrows before the map: a callback cannot inherit the
  // ternary's own narrowing of o.snapshot (control flow stops at the
  // function boundary), while a const narrowed in scope keeps it inside
  // the closure.
  const snap = (() => {
    const spec = o.snapshot;
    if (spec === undefined) {
      return "";
    }
    return ` INSERT INTO report_snapshot (id, created_at, updated_at, tenant_id, record_id, user_id, tier, composite_percentile, abilities, cohort_basis)
         VALUES ('${tag}sn', now(), now(), 'default', '${tag}${spec.anchorSuffix}', '${userId}', 'B', 75, '[${FPS_STEPS.map(
           (step, i) =>
             `{"Ability":"${step.ability}","GameCode":"${step.test_code}","Best":${o.records[i]?.score ?? 0},"Attempts":[${o.records[i]?.score ?? 0}],"BestGrade":null,"Percentile":${spec.percentiles[i]},"PrevSession":null,"HistoryBest":null}`,
         ).join(",")}]'::jsonb, 'same-age-same-sex');`;
  })();
  sql(
    "esports",
    `INSERT INTO match_session (id, created_at, updated_at, tenant_id, user_id, genre, difficulty, seed, steps, retest_budget, status)
     VALUES ('${sessionId}', '${o.startedAt}', '${o.startedAt}', 'default', '${userId}', 'fps', 'experience', 1,
             '${JSON.stringify(FPS_STEPS)}'::jsonb, 3, 'active');
     INSERT INTO test_record (id, created_at, updated_at, tenant_id, test_code, user_id, attempt_no, session_id, started_at, ended_at, raw_detail, raw_score, authoritative_score, grade, valid, invalid_reason)
     VALUES ${rows.join(",")};${snap}`,
  );
  return { sessionId };
}

/** Seeds the member's standing membership window — the app-side shape of
 * the prototype presets' member.status: one 'membership' entitlement row
 * whose expires_at is the preset's endTime. An active window and an
 * expired one are the same row shape (status stays 'active' — expiry is
 * a read-time fact of the timestamp, never a status flip; only an
 * operator terminate writes that word), so the caller's expires_at picks
 * the face. */
export function seedMemberWindow(tag: string, userId: string, expiresAt: string): void {
  sql(
    "esports",
    `INSERT INTO entitlement (id, created_at, updated_at, tenant_id, user_id, scope_kind, scope_ref, expires_at, status, source)
     VALUES ('${tag}w', now(), now(), 'default', '${userId}', 'membership', NULL, '${expiresAt}', 'active', 'e2e-seed');`,
  );
}

/** Seeds one member-window trial unlock — the row the digest's
 * all-time unlock count reads (scope 'trial', source 'membership': the
 * marker the E8 settle hook writes when a member opens a report inside
 * the window). The member-overview preset's「已解锁 1 份报告」line rides
 * exactly one. */
export function seedMembershipUnlock(tag: string, userId: string, recordId: string): void {
  sql(
    "esports",
    `INSERT INTO entitlement (id, created_at, updated_at, tenant_id, user_id, scope_kind, scope_ref, status, source)
     VALUES ('${tag}k', now(), now(), 'default', '${userId}', 'trial', '${recordId}', 'active', 'membership');`,
  );
}

/** The member-overview preset's second session: the moba lineup scored
 * at the prototype's own pcts [66, 58, 74, 62] (RA, DE, RS, SRA), each
 * record's score the prototype's own 0.9×pct+5 semantics so the merged
 * cross-genre bests land on the oracle's literals (RA 64, DE 57 — RS and
 * SRA stay under the fps session's). The snapshot freezes the moba
 * composite 65 — the overview stats'「前 35%」reads the latest session's
 * own verdict. Created 30s after the fps session: the history lines show
 * the same minute, but the latest-scored tie-break stays deterministic. */
export function seedMobaSession(tag: string, userId: string): void {
  const sessionId = `${tag}s2`;
  // One active session per member is the model's own law (a partial
  // unique on status='active'): a second genre arrives through the
  // switch flow's word — the fps session turns 'switched', the moba
  // session takes the active seat. The digest reads both rows the same
  // (its sessions aggregate carries no status filter).
  const scores = [64, 57, 71, 60];
  const grades = ["B", "C", "B", "B"];
  const pcts = [66, 58, 74, 62];
  // Seconds apart, same as the fps session: the session-anchor pick is
  // latest-by-created_at, and a four-way tie leaves the pick to the
  // storage's whim — the last step must own the latest timestamp so the
  // anchor is the record the snapshot froze on. All four stay inside the
  // same minute the history line prints.
  const times = ["09:12:30", "09:13:50", "09:15:20", "09:16:40"];
  const records = MOBA_STEPS.map((step, i) => {
    const at = `2026-09-09 ${times[i]}+08`;
    return `('${tag}m${i + 1}', '${at}', '${at}', 'default', '${step.test_code}', '${userId}', 1, '${sessionId}',
            '${at}', '${at}', '{}', ${scores[i]}, ${scores[i]}, '${grades[i]}', true, NULL)`;
  });
  const abilities = MOBA_STEPS.map(
    (step, i) =>
      `{"Ability":"${step.ability}","GameCode":"${step.test_code}","Best":${scores[i]},"Attempts":[${scores[i]}],"BestGrade":null,"Percentile":${pcts[i]},"PrevSession":null,"HistoryBest":null}`,
  );
  sql(
    "esports",
    `UPDATE match_session SET status = 'switched' WHERE id = '${tag}s';
     INSERT INTO match_session (id, created_at, updated_at, tenant_id, user_id, genre, difficulty, seed, steps, retest_budget, status)
     VALUES ('${sessionId}', '2026-09-09 09:12:30+08', '2026-09-09 09:12:30+08', 'default', '${userId}', 'moba', 'experience', 1,
             '${JSON.stringify(MOBA_STEPS)}'::jsonb, 3, 'active');
     INSERT INTO test_record (id, created_at, updated_at, tenant_id, test_code, user_id, attempt_no, session_id, started_at, ended_at, raw_detail, raw_score, authoritative_score, grade, valid, invalid_reason)
     VALUES ${records.join(",")};
     INSERT INTO report_snapshot (id, created_at, updated_at, tenant_id, record_id, user_id, tier, composite_percentile, abilities, cohort_basis)
     VALUES ('${tag}sn2', now(), now(), 'default', '${tag}m4', '${userId}', 'B', 65, '[${abilities.join(",")}]'::jsonb, 'same-age-same-sex');`,
  );
}

/** Seeds one pay_order row — the receipt page's whole truth (the wall's
 * refusal faces need no order at all). state is the stored word the
 * verdict maps straight onto (paid+scope → the success face by scope,
 * closed → cancel, pending → the honest failure: nothing settled); the
 * sku_code pair keeps the row faithful to what each face bought. The
 * pinned clock's own moment (2026-09-15 21:30) rides paid_at so the
 * receipt's time line matches the oracle's literal. */
export function seedPayOrder(
  tag: string,
  userId: string,
  o: {
    no: string;
    state: "pending" | "paid" | "closed";
    scopeKind: "trial" | "period";
    scopeRef?: string;
    amountFen: number;
    paid?: boolean;
  },
): string {
  const orderNo = `PAY20260915${o.no}`;
  sql(
    "esports",
    `INSERT INTO pay_order (id, created_at, updated_at, tenant_id, order_no, user_id, sku_code, scope_kind, scope_ref, provider, channel, amount_fen, state, credential_kind, credential_value, expires_at, paid_at)
     VALUES ('${tag}o${o.no}', '2026-09-15 21:28:00+08', '2026-09-15 21:28:00+08', 'default', '${orderNo}', '${userId}',
             '${o.scopeKind === "period" ? "member_month" : "trial_report"}', '${o.scopeKind}',
             ${o.scopeRef === undefined ? "NULL" : `'${o.scopeRef}'`}, 'mock', 'wechat_jsapi', ${o.amountFen}, '${o.state}',
             'qr_code', 'e2e-seed', '2026-09-15 23:28:00+08', ${o.paid === true ? "'2026-09-15 21:30:00+08'" : "NULL"});`,
  );
  return orderNo;
}

/** Seeds one letterbox row — the message center's msg tab and the notice
 * detail read exactly this row (a letter has no other home: no delete
 * face, no single-letter read — the list is a letter's only address).
 * read stays the row's own column: the center's unread banner counts the
 * false rows, and opening the notice detail flips one server-side. */
export function seedLetter(
  tag: string,
  userId: string,
  o: {
    suffix: string;
    type: "announcement" | "report" | "membership" | "feedback_reply";
    title: string;
    content: string;
    refId?: string;
    read?: boolean;
    at: string;
  },
): string {
  const id = `${tag}L${o.suffix}`;
  sql(
    "esports",
    `INSERT INTO user_message (id, created_at, updated_at, user_id, type, title, content, ref_id, is_read)
     VALUES ('${id}', '${o.at}', '${o.at}', '${userId}', '${o.type}', '${o.title}', '${o.content}',
             ${o.refId === undefined ? "NULL" : `'${o.refId}'`}, ${o.read === true});`,
  );
  return id;
}

/** Seeds one feedback ticket — the root row plus its turns, the exact
 * shape Submit and the two append moves write: the root's content IS the
 * first turn's text (the denormalization the list rows read), a replied
 * root carries its admin turn, and a pending one stands on the opening
 * turn alone. createdAt feeds the detail's 提交于 line; updatedAt feeds
 * the queue's order and its meta stamp. */
export function seedTicket(
  tag: string,
  userId: string,
  o: {
    suffix: string;
    category: "feature_bug" | "assessment" | "report" | "payment" | "account" | "other";
    status: "pending" | "replied" | "closed";
    content: string;
    createdAt: string;
    updatedAt: string;
    turns: { sender: "user" | "admin"; content: string; at: string }[];
  },
): string {
  const id = `${tag}F${o.suffix}`;
  const turns = o.turns
    .map(
      (turn, i) =>
        `('${tag}T${o.suffix}${i}', '${turn.at}', '${turn.at}', '${id}', '${turn.sender}', '${turn.content}', ${turn.sender === "admin" ? "'e2e-seed'" : "NULL"})`,
    )
    .join(",");
  sql(
    "esports",
    `INSERT INTO user_feedback (id, created_at, updated_at, user_id, category, content, contact, status)
     VALUES ('${id}', '${o.createdAt}', '${o.updatedAt}', '${userId}', '${o.category}', '${o.content}', NULL, '${o.status}');${o.turns.length > 0 ? ` INSERT INTO feedback_message (id, created_at, updated_at, feedback_id, sender, content, admin_id) VALUES ${turns};` : ""}`,
  );
  return id;
}

/** Seeds one report-family unlock grant — the app-side shape of the
 * prototype's unlocked.cognitive flag (the unlocked-single preset): an
 * active report_type entitlement, which the kernel's verdict lets
 * outrank the membership policy. One row per scope per user — the
 * entitle read is an Only(), so a second row is an error, never a
 * merge. */
export function seedReportGrant(tag: string, ref = "cognitive"): void {
  sql(
    "esports",
    `INSERT INTO entitlement (id, created_at, updated_at, tenant_id, user_id, scope_kind, scope_ref, status, source)
     VALUES ('${tag}e', now(), now(), 'default', '${tag}u', 'report_type', '${ref}', 'active', 'e2e-seed');`,
  );
}

/** Reads one family's report_policy override row (null = no row, the
 * shipped default answers). The career faces' pin needs the standing
 * value to restore afterwards. */
export function reportPolicyOf(family: string): string | null {
  const out = execSync(
    `${PG} -d esports -t -A -c "SELECT policy FROM report_policy WHERE report_type = '${family}'"`,
  ).toString();
  const value = out.trim();
  return value === "" ? null : value;
}

/** Rewrites one family's report_policy override row; null removes it.
 * The career pages' spec pins the row to closed — the prototype
 * presumes the backend switch off, while the shipped default rules
 * career open (D-AB) — and restores the standing value after. */
export function setReportPolicy(family: string, policy: string | null): void {
  sql(
    "esports",
    `DELETE FROM report_policy WHERE report_type = '${family}';${policy === null ? "" : ` INSERT INTO report_policy (id, created_at, updated_at, report_type, policy, updated_by) VALUES ('e2e_pin_${family}', now(), now(), '${family}', '${policy}', 'e2e-seed');`}`,
  );
}

export function cleanup(tag: string): void {
  sql(
    "system",
    `DELETE FROM credential WHERE user_id = '${tag}u';
     DELETE FROM membership WHERE id = '${tag}m';
     DELETE FROM "user" WHERE id = '${tag}u';`,
  );
  sql(
    "esports",
    `DELETE FROM pay_order WHERE user_id = '${tag}u';
     DELETE FROM report_snapshot WHERE id LIKE '${tag}sn%';
     DELETE FROM test_record WHERE id LIKE '${tag}r%' OR id LIKE '${tag}m%' OR id LIKE '${tag}c%' OR user_id LIKE '${tag}c%';
     DELETE FROM match_session WHERE id LIKE '${tag}s%';
     DELETE FROM feedback_message WHERE id LIKE '${tag}T%' OR feedback_id LIKE '${tag}F%'
       OR feedback_id IN (SELECT id FROM user_feedback WHERE user_id = '${tag}u');
     DELETE FROM user_feedback WHERE id LIKE '${tag}F%' OR user_id = '${tag}u';
     DELETE FROM user_message WHERE id LIKE '${tag}L%' OR user_id = '${tag}u';
     DELETE FROM entitlement WHERE user_id = '${tag}u';
     DELETE FROM profile WHERE id LIKE '${tag}%' OR user_id LIKE '${tag}%';`,
  );
}

export async function loginAndOpen(
  page: Page,
  account: string,
  path = "/",
  clock?: string,
): Promise<void> {
  const res = await fetch(`${SYSTEM_HTTP}/api/v1/system/auth/login`, {
    method: "POST",
    headers: { "content-type": "application/json" },
    body: JSON.stringify({ account, password: "admin123456", population: "member" }),
  });
  const json = (await res.json().catch(() => ({}))) as { sessionToken?: string };
  expect(json.sessionToken, `member login for ${account}`).toBeTruthy();
  // The member world's namespaced session cookie (custody.server's own
  // spelling), planted directly — the UI login flow is not this spec's
  // subject.
  await page.context().addCookies([
    {
      name: "session_token_member",
      value: String(json.sessionToken),
      url: "http://localhost:5189",
    },
  ]);
  if (clock !== undefined) {
    // Faces that derive from the browser clock (the resume card's day
    // count, the age step's year enum) ride the page clock to the
    // oracle's own moment — a fixture frozen in 2026 must never meet a
    // 2027 calendar.
    await page.clock.install({ time: new Date(clock).getTime() });
  }
  await page.goto(path);
  await page.waitForLoadState("networkidle");
  // The settle marker each test names proves its member queries landed;
  // the rest of the face follows.
}

export async function collectFace(
  page: Page,
  settleFor: string,
  lingerMs = 400,
): Promise<{ raw: string; placeholders: string[] }> {
  await page.waitForFunction(
    // biome-ignore lint/performance/useDomNodeTextContent: innerText is the deliberate visibility oracle — the wait polls the same face the assert below reads; textContent would admit hidden nodes and settle early on markers a visitor cannot see.
    (marker) => document.body.innerText.includes(marker),
    settleFor,
    { timeout: 15_000 },
  );
  await page.waitForTimeout(lingerMs);
  return page.evaluate(() => {
    // Dev chrome never crosses to production faces, so it leaves the
    // oracle comparison too — the same strip the fixture side applies to
    // the prototype's own reviewer gear (REVIEW_CHROME): the breakpoint
    // pill declares itself (data-dev-chrome) and the devtools toggle
    // carries its own fixed label.
    document.querySelectorAll("[data-dev-chrome]").forEach((n) => {
      n.remove();
    });
    for (const b of document.querySelectorAll("button")) {
      if (b.textContent?.includes("TanStack Devtools")) {
        b.remove();
      }
    }
    return {
      // biome-ignore lint/performance/useDomNodeTextContent: innerText is the deliberate oracle — visibility-respecting text is what fidelity asserts; textContent would admit hidden nodes and script bodies.
      raw: document.body.innerText,
      placeholders: [
        ...document.querySelectorAll<HTMLInputElement | HTMLTextAreaElement>("input, textarea"),
      ]
        .filter((el) => el.getClientRects().length > 0)
        .map((el) => el.placeholder)
        .filter(Boolean),
    };
  });
}

export function assertFace(
  fixture: Fixture,
  face: { raw: string; placeholders: string[] },
  appTitle: string,
): void {
  const actual = normalizeLines(face.raw);
  const missing = multisetMinus(fixture.lines, actual);
  const extra = multisetMinus(actual, fixture.lines);
  const missingPh = multisetMinus(fixture.placeholders, face.placeholders);
  const extraPh = multisetMinus(face.placeholders, fixture.placeholders);
  const sameOrder = JSON.stringify(actual) === JSON.stringify(fixture.lines);
  expect(
    missing.length === 0 &&
      extra.length === 0 &&
      missingPh.length === 0 &&
      extraPh.length === 0 &&
      sameOrder &&
      appTitle === fixture.title,
    [
      `title: prototype=${JSON.stringify(fixture.title)} app=${JSON.stringify(appTitle)}`,
      `missing lines: ${JSON.stringify(missing)}`,
      `extra lines: ${JSON.stringify(extra)}`,
      `missing placeholders: ${JSON.stringify(missingPh)}`,
      `extra placeholders: ${JSON.stringify(extraPh)}`,
      `orderMatch: ${sameOrder}`,
    ].join("\n"),
  ).toBe(true);
}

export function multisetMinus(a: string[], b: string[]): string[] {
  const pool = new Map<string, number>();
  for (const line of b) pool.set(line, (pool.get(line) ?? 0) + 1);
  const out: string[] = [];
  for (const line of a) {
    const n = pool.get(line) ?? 0;
    if (n === 0) out.push(line);
    else pool.set(line, n - 1);
  }
  return out;
}
