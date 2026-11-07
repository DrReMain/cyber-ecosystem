import { test } from "@playwright/test";
import {
  assertFace,
  cleanup,
  collectFace,
  fixtureOf,
  loginAndOpen,
  reportPolicyOf,
  seedMember,
  seedReportGrant,
  seedScoredSession,
  setReportPolicy,
} from "./harness";

// The report family against the prototype oracle: the single report and
// the trend at tested-fps (one scored fps session — the single page's
// locked face, the free seat's trio the only open detail), the center
// at the same state, cognitive at unlocked-single (tested-fps plus the
// purchased cognitive unlock — seeded as the report_type grant the
// kernel's verdict lets outrank the membership policy), and career's
// closed face (the prototype presumes the backend switch off; the spec
// pins the report_policy row to closed and restores the standing value
// — the shipped default rules career open, D-AB). Runs only against
// the live dev stack (CTR_E2E=1). Seeds and face laws live in
// ./harness.
//
// Standing, ledgered divergences keep these specs red on purpose (the
// login spec's precedent):
// · icon law — the prototype's emoji glyphs (‹ back fabs, 👑 🔒 📊 🧠 🎯
//   📈, the ✓ refresh mark) render as lucide SVG in the app.
// · sub-ability family — the prototype's radar fifth axis (决策效率,
//   the is_primary=0 aggregation) is a self-declared 示意值 with the
//   data layer 待修正; the contract carries no sub face, so the app
//   renders the honest four-axis shape and its own no-subs note (the
//   sub's locked row and its unlock CTA ride the same family).
// · parked faces — the single page's 分享测试结果 button (MSG-2) and
//   the cognitive page's three config modules 学科优势匹配 / 职业发展
//   潜能 / 提升计划 (P4/ljr: contract-less, 后台可配置) have no app
//   counterpart yet.
// · audit-passed app shapes — the single page carries its verdict once
//   (the composite card's chip), not twice as the prototype's header
//   badge + card tag pair; the E5 audit accepted the one-chip shape.
test.skip(!process.env.CTR_E2E, "report fidelity needs the live stack (CTR_E2E=1)");
test.use({ locale: "zh-CN" });

test("report single (tested-fps, locked face) matches the prototype snapshot", async ({ page }) => {
  const tag = `fr${Date.now().toString(36).slice(-7)}s`;
  const { account, userId } = seedMember(tag, "ready");
  const { anchorId } = seedScoredSession(tag, userId);
  try {
    // The composite card's 前 25% is the live judgement's proof: the
    // locked face benches the board read-time (pools 78/86/71/64 mean
    // 74.75 → 25 / 75), never the frozen snapshot.
    await loginAndOpen(page, account, `/report/${anchorId}`);
    const face = await collectFace(page, "前 25%");
    assertFace(fixtureOf("report-single", "tested-fps"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});

test("report center (tested-fps, career pinned closed) matches the prototype snapshot", async ({
  page,
}) => {
  const tag = `fr${Date.now().toString(36).slice(-7)}c`;
  const { account, userId } = seedMember(tag, "ready");
  seedScoredSession(tag, userId);
  const standing = reportPolicyOf("career");
  setReportPolicy("career", "closed");
  try {
    // The coverage banner's 4 / 6 proves the digest answered; the four
    // cards' tags ride the same read (cognitive locked for a bare
    // member, trend's free base tier for a measured one).
    await loginAndOpen(page, account, "/report/center");
    const face = await collectFace(page, "已覆盖 4 / 6");
    assertFace(fixtureOf("report-center", "tested-fps"), face, await page.title());
  } finally {
    setReportPolicy("career", standing);
    cleanup(tag);
  }
});

test("report cognitive (unlocked-single) matches the prototype snapshot", async ({ page }) => {
  const tag = `fr${Date.now().toString(36).slice(-7)}g`;
  const { account, userId } = seedMember(tag, "ready");
  seedScoredSession(tag, userId);
  seedReportGrant(tag, "cognitive");
  try {
    // 神经反应 is the weighted dimension's own word — its presence
    // proves the grant opened the gate and both source percentiles
    // (78 / 86) fed the matrix.
    await loginAndOpen(page, account, "/report/cognitive");
    const face = await collectFace(page, "神经反应");
    assertFace(fixtureOf("report-cognitive", "unlocked-single"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});

test("report trend (tested-fps) matches the prototype snapshot", async ({ page }) => {
  const tag = `fr${Date.now().toString(36).slice(-7)}t`;
  const { account, userId } = seedMember(tag, "ready");
  seedScoredSession(tag, userId);
  try {
    // The banner's counts prove the chronology answered: one session,
    // four series — each a single point read from the frozen snapshot's
    // percentiles.
    await loginAndOpen(page, account, "/report/trend");
    const face = await collectFace(page, "共 1 次测评");
    assertFace(fixtureOf("report-trend", "tested-fps"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});

test("report career (policy pinned closed) matches the prototype snapshot", async ({ page }) => {
  const tag = `fr${Date.now().toString(36).slice(-7)}x`;
  const { account } = seedMember(tag, "ready");
  const standing = reportPolicyOf("career");
  setReportPolicy("career", "closed");
  try {
    // The closed face is pure copy — a logged member with no records
    // keeps the page honest (the open default's library faces are the
    // e2e report tests' own subject, not the oracle's).
    await loginAndOpen(page, account, "/report/career");
    const face = await collectFace(page, "暂未开放");
    assertFace(fixtureOf("report-career", "ready"), face, await page.title());
  } finally {
    setReportPolicy("career", standing);
    cleanup(tag);
  }
});
