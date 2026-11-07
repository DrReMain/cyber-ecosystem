import { test } from "@playwright/test";
import {
  assertFace,
  cleanup,
  collectFace,
  fixtureOf,
  loginAndOpen,
  seedMember,
  seedResumeSession,
  seedScoredSession,
} from "./harness";

// The home face against the prototype oracle, one spec per review preset
// (ready / resume / tested-fps). The oracle corpus is zh-CN and the app
// negotiates language, so the comparison pins the locale. Home is a
// member-data face, not a static one: each preset is a freshly seeded
// member whose rows reproduce the prototype state the fixture froze —
// the peek session, the scored record with its frozen snapshot, and the
// cohort pool that wakes the reaction percentile. Runs only against the
// live dev stack (CTR_E2E=1): postgres + system + esports up. The seed
// machinery and the face laws live in ./harness.
//
// Standing, ledgered divergences keep these specs red on purpose (the
// login spec's precedent): the prototype's emoji glyphs (🔔 📄 › 🏠 📊 💬
// 👤) render as lucide SVG in the app — icon law, not copy; the resume
// line's "7 天内可续" tail names the retired seven-day window the match
// face no longer enforces; "看报告 ›" pairs the word with a chevron SVG.
test.skip(!process.env.CTR_E2E, "home fidelity needs the live stack (CTR_E2E=1)");
test.use({ locale: "zh-CN" });

test("home (preset ready) text matches the prototype snapshot", async ({ page }) => {
  const tag = `fh${Date.now().toString(36).slice(-7)}`;
  const { account } = seedMember(tag, "ready");
  try {
    await loginAndOpen(page, account);
    const face = await collectFace(page, "数据统计");
    assertFace(fixtureOf("home", "ready"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});

test("home (preset resume) text matches the prototype snapshot", async ({ page }) => {
  const tag = `fh${Date.now().toString(36).slice(-7)}z`;
  const { account, userId } = seedMember(tag, "ready");
  seedResumeSession(tag, userId);
  try {
    // Day three of a session started 2026-09-11 21:04 (the fixture's own
    // arithmetic: the inclusive day count reads 3 through 09-14 21:03).
    await loginAndOpen(page, account, "/", "2026-09-14T12:00:00+08:00");
    const face = await collectFace(page, "有一场未完成的测评");
    assertFace(fixtureOf("home", "resume"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});

test("home (preset tested-fps) text matches the prototype snapshot", async ({ page }) => {
  const tag = `fh${Date.now().toString(36).slice(-7)}t`;
  const { account, userId } = seedMember(tag, "ready");
  seedScoredSession(tag, userId);
  try {
    await loginAndOpen(page, account);
    const face = await collectFace(page, "最近一次测评");
    assertFace(fixtureOf("home", "tested-fps"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});
