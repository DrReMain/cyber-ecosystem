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

// The cohort pool behind the tested preset's reaction percentile: the
// member's red_dot 75 benched against these ten same-age-same-sex scores
// lands at raw percentile 78.01 → 78 (round), a hair off both bucket
// edges so float noise cannot flip it.
const COHORT_SCORES = [40, 62, 63, 68, 69, 71, 72, 73, 73, 74];

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
  opts: { phone?: string; wechat?: string } = {},
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
  sql(
    "system",
    `INSERT INTO "user" (id, created_at, updated_at, account, password_hash, enabled)
     VALUES ('${userId}', now(), now(), '${account}', '${PASSWORD_HASH}', true);
     INSERT INTO membership (id, created_at, updated_at, user_id, tenant_id, population)
     VALUES ('${tag}m', now(), now(), '${userId}', 'default', 'member');
     ${credentials}`,
  );
  return { account, userId };
}

/** Seeds the esports profile row on a bare member. The two shapes the
 * presets froze: "ready" (the complete member) and "profile-missing"
 * (nickname and a two-game preference, no birth — exactly the wizard's
 * subject). Sex rides the real storage word, never the prototype's
 * display glyph. */
export function seedProfile(tag: string, userId: string, shape: "ready" | "profile-missing"): void {
  const games = shape === "ready" ? "('王者荣耀','英雄联盟','CS2')" : "('王者荣耀','CS2')";
  // Birth year 2004, not the prototype's 2008: the cohort key is birth
  // year + sex, and the shared dev pool carries a legacy 2008 male
  // calibration batch (another lane's residue, left in place) that a 2008
  // member would absorb — the frozen percentile needs a hermetic pool.
  // 2004-06 is 22 at the fixture's own moment, so the stored 18-22 band
  // stays honest, and no oracle line quotes the birth.
  const birth = shape === "ready" ? "2004, 6, 'male', '18-22'" : "NULL, NULL, 'male', NULL";
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
  shape: "ready" | "profile-missing" = "ready",
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
 * four percentiles), and the ten-row cohort that wakes the reaction
 * best's live percentile to 78. */
export function seedScoredSession(tag: string, userId: string): void {
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
  const cohortProfiles = COHORT_SCORES.map(
    (_, i) =>
      `('${tag}c${i}p', now(), now(), 'default', '${tag}c${i}', NULL, 2004, 6, 'male', '18-22', NULL)`,
  );
  const cohortRecords = COHORT_SCORES.map(
    (score, i) =>
      `('${tag}c${i}r', now(), now(), 'default', 'red_dot', '${tag}c${i}', 1, NULL, now(), now(), '{}', ${score}, ${score}, true, NULL)`,
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
    `DELETE FROM report_snapshot WHERE id = '${tag}sn';
     DELETE FROM test_record WHERE id LIKE '${tag}r%' OR id LIKE '${tag}c%r' OR user_id LIKE '${tag}c%';
     DELETE FROM match_session WHERE id = '${tag}s';
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
): Promise<{ raw: string; placeholders: string[] }> {
  await page.waitForFunction(
    // biome-ignore lint/performance/useDomNodeTextContent: innerText is the deliberate visibility oracle — the wait polls the same face the assert below reads; textContent would admit hidden nodes and settle early on markers a visitor cannot see.
    (marker) => document.body.innerText.includes(marker),
    settleFor,
    { timeout: 15_000 },
  );
  await page.waitForTimeout(400);
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
