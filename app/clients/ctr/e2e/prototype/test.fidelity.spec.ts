import { test } from "@playwright/test";
import {
  assertFace,
  cleanup,
  collectFace,
  fixtureOf,
  loginAndOpen,
  seedMember,
  seedTestSession,
} from "./harness";

// The test family against its oracles: the guide (/match/guide — the
// name mismatch is the matrix's own note, the须知 lives in the match
// domain), the single-test page's intro / running / result faces (one
// route, three phases — running is the matrix's only landscape page and
// its only interactive oracle), and the closing ceremony on the session
// report (/report/session/$sessionId). Four faces ride one seeded
// shape — the prototype's resume preset, an in-flight fps session with
// attempts on two of the four lineup games — and the fifth seeds its
// own fully-played session with a frozen snapshot, the shape the
// prototype's finishSession folds. No clock is pinned: no face quotes
// the client or server clock (the running face's live numbers are the
// ledgered coordinate gap, not a pin candidate). Runs only against the
// live dev stack (CTR_E2E=1). Seeds and face laws live in ./harness.
//
// Shape law: every seed call sits inside the try, before the page work
// (the pay family's own law — cleanup is idempotent DELETEs, so a
// half-seeded session cleans up as safely as a whole one).
//
// Standing, ledgered divergences keep these specs red on purpose (the
// login spec's precedent):
// · icon law — the guide's ⚠️ 🔆 ⏱ 🔁 🏆 💾 glyph rows render as lucide
//   SVG in the app, and the ‹ back fabs likewise (the family account).
// · document titles — the prototype's own title/h1 pairs disagree
//   (guide: 测试说明 vs 测评前须知; the single-test page names each
//   phase 单项测试介绍 / 测试中 / 单项测试结果); the app titles every
//   face by its h1 word (测评前须知 / 测试进行中) — honest naming, the
//   four-face account.
// · the resume window — the prototype's localStorage semantics quote
//   7-day / 24-hour cache expiry (rule 5's tail, the guide footer's
//   （7 天内可续测）); the app's session lives server-side, locked in
//   its row, so its words are 进度会一直保留 / 随时可续测.
// · the intro's unit line — the oracle prints FPS · 10 分/命中; the
//   catalog's unit_score_label rides empty and the category word is
//   the app's own hidden vocabulary (Q1-⑦), so the line never renders.
// · the history-best semantics — the app's session report floors
//   historyBest at the session's own best (never null), while the
//   prototype reads cross-session records only: the intro's prev card
//   and the result's best tag print 78 where the oracle prints —.
// · the running face's demo apparatus — the prototype's placeholder
//   game (cd 5, 采样中… ticks, the skip button, the 剩余时间 note) has
//   no app counterpart: the real game runs 30 live seconds, and the
//   HUD's two-decimal countdown is a live number — the coordinate gap.
// · the run header's wording — the oracle's 反应速度 · 1 / 4 against
//   the app's fuller 反应速度 · 第 1 题 / 共 4 题; and the × exit fab
//   against the pause control (the app pauses first, exits from the
//   overlay).
// · the result face's detail rows — the oracle's three rows are demo
//   formulas off the best score; a read-back visit has no local facts
//   to derive, so the app renders no fact rows rather than inventing
//   them (the end face's own law). The receipt's wording differs with
//   it: 多次尝试取最高分 generalizes the oracle's 3 次取最高 (first
//   sessions carry three chances, later ones one), the （后台可看全貌）
//   aside has no app home, and the retake chip spaces its numbers
//   (1 / 3 against 1/3).
// · the score-currency line's seam — the oracle's 成绩口径 row folds a
//   source-code newline into a space after 展示； the app's message
//   carries the semicolon alone; the receipt's —— dash wears spaces on
//   the oracle side too.
// · the done grid's icon tiles — the prototype's tiles are empty
//   color chips; the app stamps each with the ability word's first two
//   characters (瞄准 / 反应 / 追踪 / 技能), four extra lines.
// · innerText's occlusion blind spot — the app's single-test route
//   keeps the run HUD mounted under the intro/result overlays (the
//   overlays cover it visually, but innerText admits covered text), so
//   those two faces carry the HUD's three lines as extra rows. A
//   collection-mechanism account, not a product gap.
test.skip(!process.env.CTR_E2E, "test fidelity needs the live stack (CTR_E2E=1)");
test.use({ locale: "zh-CN" });

// The resume shape the four in-flight faces share: the session's own
// timeline (two games played the evening it opened), the second red_dot
// attempt carrying the grade the result face prints (the best attempt's
// own row), the flash_aiming attempt riding gradeless — no face renders
// it. retestLeft needs no seeding: the view derives budget-minus-used
// (3−2=1, 3−1=2), exactly the preset's own numbers.
const RESUME_RECORDS = [
  { suffix: "r1", code: "red_dot", attempt: 1, score: 71, at: "2026-09-11 21:09:30+08" },
  {
    suffix: "r2",
    code: "red_dot",
    attempt: 2,
    score: 78,
    grade: "B",
    at: "2026-09-11 21:11:00+08",
  },
  { suffix: "r3", code: "flash_aiming", attempt: 1, score: 86, at: "2026-09-11 21:14:10+08" },
];

function seedResume(tag: string, userId: string): void {
  seedTestSession(tag, userId, {
    startedAt: "2026-09-11 21:04:00+08",
    records: RESUME_RECORDS,
  });
}

test("match guide (测评前须知) matches the prototype snapshot", async ({ page }) => {
  const tag = `ft${Date.now().toString(36).slice(-7)}a`;
  try {
    const { account, userId } = seedMember(tag, "ready");
    // The in-flight session the guide reads: four steps, two played.
    // firstSession rides the row's own budget (3 = first), and the
    // rule rows quote the lineup count the steps carry.
    seedResume(tag, userId);
    await loginAndOpen(page, account, "/match/guide");
    const face = await collectFace(page, "我准备好了");
    assertFace(fixtureOf("test-guide", "guide"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});

test("test intro (red_dot, mid-session) matches the prototype snapshot", async ({ page }) => {
  const tag = `ft${Date.now().toString(36).slice(-7)}b`;
  try {
    const { account, userId } = seedMember(tag, "ready");
    // The first lineup game, mid-session: the intro reads the session
    // context (2 / 4 played, one retake left) and the report line the
    // played receipt prints (session best 78 — its history floor the
    // ledger accounts).
    seedResume(tag, userId);
    await loginAndOpen(page, account, "/test/red_dot");
    const face = await collectFace(page, "继续 · 开始测试");
    assertFace(fixtureOf("test-intro", "intro"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});

test("test running (red_dot, live game) matches the prototype snapshot", async ({ page }) => {
  const tag = `ft${Date.now().toString(36).slice(-7)}c`;
  try {
    const { account, userId } = seedMember(tag, "ready");
    // The only interactive oracle: the app reaches its running face by
    // starting the run for real — the CTA, then the app's own 3-2-1
    // countdown (an addition the prototype jumps straight past), then
    // the live game. The capture lands inside the 30-second window the
    // game runs; the HUD's own seconds line is the live number the
    // ledger carries, never a pinned one.
    seedResume(tag, userId);
    await loginAndOpen(page, account, "/test/red_dot");
    await collectFace(page, "继续 · 开始测试");
    await page.getByRole("button", { name: "继续 · 开始测试" }).click();
    // The countdown overlay owns the next three seconds; settle past it
    // before the face read.
    await page.waitForTimeout(3400);
    const face = await collectFace(page, "测试中 · 随机红点触摸测试");
    assertFace(fixtureOf("test-running", "running"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});

test("test result (red_dot, read-back) matches the prototype snapshot", async ({ page }) => {
  const tag = `ft${Date.now().toString(36).slice(-7)}d`;
  try {
    const { account, userId } = seedMember(tag, "ready");
    // review=true lands on the stored verdict — the done step's
    // read-back (the prototype's own ro face): the report line's best
    // with its grade, the two-attempt receipt, and the continue CTA
    // naming the next undone game. The settle marker waits for the
    // verdict chip, not the header — the header lives on every phase of
    // this route. The word form is the URL's honest shape: the router's
    // search parser JSON-decodes 1 into a number, and the route's
    // validator only grants the boolean and the string — a numeric 1
    // would be rewritten to false in the address bar.
    seedResume(tag, userId);
    await loginAndOpen(page, account, "/test/red_dot?review=true");
    const face = await collectFace(page, "B · 良好");
    assertFace(fixtureOf("test-result", "result"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});

test("session done (closing ceremony) matches the prototype snapshot", async ({ page }) => {
  const tag = `ft${Date.now().toString(36).slice(-7)}e`;
  try {
    const { account, userId } = seedMember(tag, "ready");
    // The fully-played session the oracle's finishSession folds: every
    // lineup game with one attempt, the scores the score rows print
    // (75 / 82 / 69 / 63 — the prototype's round(0.9·pct+5) currency),
    // the grades the chips print, and the frozen snapshot whose
    // percentiles (78 / 86 / 71 / 64) the beat rows read verbatim —
    // pool-free by design, the snapshot is the whole cohort answer.
    // The four stamps span 220 seconds first-to-last: the ceremony's
    // own 用时约 3 分 40 秒 is the play span the app derives live.
    const { sessionId } = seedTestSession(tag, userId, {
      startedAt: "2026-09-15 21:00:00+08",
      records: [
        {
          suffix: "r1",
          code: "red_dot",
          attempt: 1,
          score: 75,
          grade: "B",
          at: "2026-09-15 21:04:30+08",
        },
        {
          suffix: "r2",
          code: "flash_aiming",
          attempt: 1,
          score: 82,
          grade: "A",
          at: "2026-09-15 21:06:20+08",
        },
        {
          suffix: "r3",
          code: "moving_target",
          attempt: 1,
          score: 69,
          grade: "B",
          at: "2026-09-15 21:07:40+08",
        },
        {
          suffix: "r4",
          code: "skill_trajectory",
          attempt: 1,
          score: 63,
          grade: "B",
          at: "2026-09-15 21:08:10+08",
        },
      ],
      snapshot: { anchorSuffix: "r4", percentiles: [78, 86, 71, 64] },
    });
    await loginAndOpen(page, account, `/report/session/${sessionId}`);
    const face = await collectFace(page, "4/4 完成");
    assertFace(fixtureOf("test-done", "done"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});
