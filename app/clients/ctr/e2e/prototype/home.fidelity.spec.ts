import { execSync } from "node:child_process";
import { readFileSync } from "node:fs";
import { expect, type Page, test } from "@playwright/test";
import { normalizeLines } from "./text.mjs";

// The home face against the prototype oracle, one spec per review preset
// (ready / resume / tested-fps). The oracle corpus is zh-CN and the app
// negotiates language, so the comparison pins the locale. Home is a
// member-data face, not a static one: each preset is a freshly seeded
// member whose rows reproduce the prototype state the fixture froze —
// the peek session, the scored record with its frozen snapshot, and the
// cohort pool that wakes the reaction percentile. Runs only against the
// live dev stack (CTR_E2E=1): postgres + system + esports up.
//
// Standing, ledgered divergences keep these specs red on purpose (the
// login spec's precedent): the prototype's emoji glyphs (🔔 📄 › 🏠 📊 💬
// 👤) render as lucide SVG in the app — icon law, not copy; the resume
// line's "7 天内可续" tail names the retired seven-day window the match
// face no longer enforces; "看报告 ›" pairs the word with a chevron SVG.
test.skip(!process.env.CTR_E2E, "home fidelity needs the live stack (CTR_E2E=1)");
test.use({ locale: "zh-CN" });

const SYSTEM_HTTP = "http://localhost:11001";
const PG = "docker exec -i cyber-ecosystem-postgresql-1 psql -U postgres -v ON_ERROR_STOP=1";
// The admin row's argon2id digest (password admin123456) — copied, not
// minted: a static literal keeps the seed deterministic with zero crypto
// in the runner; each member still owns its own credential row set.
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

interface Fixture {
  page: string;
  preset: string | null;
  source: string;
  title: string;
  lines: string[];
  placeholders: string[];
}

function fixtureOf(preset: string): Fixture {
  return JSON.parse(
    readFileSync(new URL(`./fixtures/home__${preset}.json`, import.meta.url), "utf8"),
  ) as Fixture;
}

function sql(db: "system" | "esports", statements: string): void {
  execSync(`${PG} -d ${db}`, { input: statements });
}

/** Seeds one member: the system rows (user handle + member seat) and the
 * esports profile the prototype's READY() base state froze — nickname,
 * sex, birth, and the three-game preference. Returns the ids the cleanup
 * knows. */
function seedMember(tag: string): { account: string; userId: string } {
  const userId = `${tag}u`;
  const account = `${tag}a`;
  sql(
    "system",
    `INSERT INTO "user" (id, created_at, updated_at, account, password_hash, enabled)
     VALUES ('${userId}', now(), now(), '${account}', '${PASSWORD_HASH}', true);
     INSERT INTO membership (id, created_at, updated_at, user_id, tenant_id, population)
     VALUES ('${tag}m', now(), now(), '${userId}', 'default', 'member');`,
  );
  sql(
    "esports",
    `INSERT INTO profile (id, created_at, updated_at, tenant_id, user_id, nickname, birth_year, birth_month, sex, age_band, games)
     VALUES ('${tag}p', now(), now(), 'default', '${userId}', '龘好久', 2008, 6, '男', '18-22',
             -- games entries are catalog refs (GameEntry{game_id}), so the
             -- preset's three picks resolve by name against the seeded
             -- catalog — never hardcoded ids that drift with re-seeds.
             (SELECT jsonb_agg(jsonb_build_object('game_id', id)) FROM game WHERE name IN ('王者荣耀','英雄联盟','CS2')));`,
  );
  return { account, userId };
}

/** The resume preset's unfinished session: four frozen steps, two of them
 * carrying spent-but-invalid attempts — invalid on purpose, the only legal
 * row shape that keeps both truths the fixture froze: the steps read done
 * (any stored record marks its step done) while the session never counts
 * as a completed one (valid attempts are what the counter joins). */
function seedResumeSession(tag: string, userId: string): void {
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
 * (the prototype's rec() score semantics), the frozen snapshot that owns
 * the verdict face (tier B, composite 75, the four percentiles), and the
 * ten-row cohort that wakes the reaction best's live percentile to 78. */
function seedScoredSession(tag: string, userId: string): void {
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
            '${at}', '${at}', '{}', ${scores[step.test_code]}, ${scores[step.test_code]}, true, NULL)`;
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
      `('${tag}c${i}p', now(), now(), 'default', '${tag}c${i}', NULL, 2008, 6, '男', '18-22', NULL)`,
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
     INSERT INTO test_record (id, created_at, updated_at, tenant_id, test_code, user_id, attempt_no, session_id, started_at, ended_at, raw_detail, raw_score, authoritative_score, valid, invalid_reason)
     VALUES ${records.join(",")};
     INSERT INTO report_snapshot (id, created_at, updated_at, tenant_id, record_id, user_id, tier, composite_percentile, abilities, cohort_basis)
     VALUES ('${tag}sn', now(), now(), 'default', '${anchorId}', '${userId}', 'B', 75, '[${abilities.join(",")}]'::jsonb, 'same-age-same-sex');
     INSERT INTO profile (id, created_at, updated_at, tenant_id, user_id, nickname, birth_year, birth_month, sex, age_band, games)
     VALUES ${cohortProfiles.join(",")};
     INSERT INTO test_record (id, created_at, updated_at, tenant_id, test_code, user_id, attempt_no, session_id, started_at, ended_at, raw_detail, raw_score, authoritative_score, valid, invalid_reason)
     VALUES ${cohortRecords.join(",")};`,
  );
}

function cleanup(tag: string): void {
  sql(
    "system",
    `DELETE FROM membership WHERE id = '${tag}m';
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

async function loginAndOpen(page: Page, account: string, clock?: string): Promise<void> {
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
    // The resume card's age face derives from the browser clock; the
    // fixture froze "已过 3 天" against its own start stamp, so the page
    // clock rides along to the oracle's day three.
    await page.clock.install({ time: new Date(clock).getTime() });
  }
  await page.goto("/");
  await page.waitForLoadState("networkidle");
  // Let the member queries land: the stats card's title marks the digest
  // answered; the preset's own headline marks the rest of the face.
}

async function collectFace(
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

function assertFace(
  fixture: Fixture,
  face: { raw: string; placeholders: string[] },
  appTitle: string,
) {
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

test("home (preset ready) text matches the prototype snapshot", async ({ page }) => {
  const tag = `fh${Date.now().toString(36).slice(-7)}`;
  const { account } = seedMember(tag);
  try {
    await loginAndOpen(page, account);
    const face = await collectFace(page, "数据统计");
    assertFace(fixtureOf("ready"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});

test("home (preset resume) text matches the prototype snapshot", async ({ page }) => {
  const tag = `fh${Date.now().toString(36).slice(-7)}z`;
  const { account, userId } = seedMember(tag);
  seedResumeSession(tag, userId);
  try {
    // Day three of a session started 2026-09-11 21:04 (the fixture's own
    // arithmetic: the inclusive day count reads 3 through 09-14 21:03).
    await loginAndOpen(page, account, "2026-09-14T12:00:00+08:00");
    const face = await collectFace(page, "有一场未完成的测评");
    assertFace(fixtureOf("resume"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});

test("home (preset tested-fps) text matches the prototype snapshot", async ({ page }) => {
  const tag = `fh${Date.now().toString(36).slice(-7)}t`;
  const { account, userId } = seedMember(tag);
  seedScoredSession(tag, userId);
  try {
    await loginAndOpen(page, account);
    const face = await collectFace(page, "最近一次测评");
    assertFace(fixtureOf("tested-fps"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});

function multisetMinus(a: string[], b: string[]): string[] {
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
