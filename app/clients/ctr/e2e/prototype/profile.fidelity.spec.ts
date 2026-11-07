import { test } from "@playwright/test";
import {
  assertFace,
  cleanup,
  collectFace,
  fixtureOf,
  loginAndOpen,
  seedBareMember,
  seedMember,
  seedScoredSession,
} from "./harness";

// The profile family against the prototype oracle: the three wizard
// steps frozen at profile-missing (the wizard's own subject — a logged,
// partial member), the overview at tested-fps (the richest member-data
// face, seeded like home's: scored session + snapshot + cohort + the
// credentials the prototype's READY() member carries), and password's
// set / change faces plus forget's rebuild at /login/forgot (G2's
// pre-login-context split). Runs only against the live dev stack
// (CTR_E2E=1). Seeds and face laws live in ./harness.
//
// Standing, ledgered divergences keep these specs red on purpose (the
// login spec's precedent):
// · icon law — the prototype's emoji glyphs (‹ back fabs, 💬 📱 🔒 👤 👑
//   📈 🏠 📊 💬 👤) render as lucide SVG in the app; the 💬-prefixed
//   wechat pill on the nickname step is the same family.
// · ruled distillations — the nickname rules card and placeholder
//   (A2: prototype design annotations are not user copy), the age sex
//   hint's wechat-autofill tail (D-AN ruled the autofill out), and the
//   ICP footer's 备 (A1: the app restores the real filing format the
//   prototype dropped).
// · pending product calls — the age step's eight-band fallback card is
//   P11 (contract is single-track birth; ljr to rule).
// · contract-less faces — the overview meta's city tail (no location
//   source) and the improvement stat's dash (no composite-sequence read
//   face; the prototype shows 0).
// · ruled app additions — the overview's 我的资料 edit card and the
//   language row (D-AT) have no prototype counterpart; password's field
//   labels, bound-phone notice, set-note and forgot cross-link are the
//   audited G2 shape; the prototype's demo footer (演示：…) is a design
//   annotation that never migrates (G2), as is forget's （type=reset）
//   parenthetical.
test.skip(!process.env.CTR_E2E, "profile fidelity needs the live stack (CTR_E2E=1)");
test.use({ locale: "zh-CN" });

// The nickname step under WeChat's UA: the pill is the prototype's
// standing wechatEnv=true assumption made real — the app gates it on the
// actual UA, so the honest comparison rides the same environment the
// prototype presumes.
test.describe("profile wizard · nickname (WeChat UA)", () => {
  test.use({ userAgent: "Mozilla/5.0 (Linux; Android 14) MicroMessenger/8.0.49" });

  test("profile setup (nickname) matches the prototype snapshot", async ({ page }) => {
    const tag = `fp${Date.now().toString(36).slice(-7)}n`;
    const { account } = seedMember(tag, "profile-missing");
    try {
      await loginAndOpen(page, account, "/profile/setup/nickname");
      // The counter's 3 / 12 is the prefill's proof: the stored nickname
      // rode the profile read into the input — counted in the app's
      // ruled character cap (A2), not the prototype's weighted 24.
      const face = await collectFace(page, "3 / 12");
      assertFace(fixtureOf("profile-nickname", "profile-missing"), face, await page.title());
    } finally {
      cleanup(tag);
    }
  });
});

test("profile setup (age) matches the prototype snapshot", async ({ page }) => {
  const tag = `fp${Date.now().toString(36).slice(-7)}a`;
  const { account } = seedMember(tag, "profile-missing");
  try {
    // The year enum is clock-fed (1966–2020 at the prototype's own
    // 2026); the frozen fixture must never meet a later calendar.
    await loginAndOpen(page, account, "/profile/setup/age", "2026-09-14T12:00:00+08:00");
    const face = await collectFace(page, "出生年");
    assertFace(fixtureOf("profile-age", "profile-missing"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});

test("profile setup (games) matches the prototype snapshot", async ({ page }) => {
  const tag = `fp${Date.now().toString(36).slice(-7)}g`;
  const { account } = seedMember(tag, "profile-missing");
  try {
    await loginAndOpen(page, account, "/profile/setup/games");
    // The vote note proves both reads landed: the catalog painted the
    // chips and the profile row pre-picked the preset's two games.
    const face = await collectFace(page, "已选 2 款 · 类型投票 FPS 1 / MOBA 1");
    assertFace(fixtureOf("profile-setup", "profile-missing"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});

test("profile overview (preset tested-fps) matches the prototype snapshot", async ({ page }) => {
  const tag = `fp${Date.now().toString(36).slice(-7)}o`;
  const { account, userId } = seedMember(tag, "ready", {
    phone: "15312347568",
    wechat: "wx_9a2f",
  });
  seedScoredSession(tag, userId);
  try {
    await loginAndOpen(page, account, "/profile");
    // The composite stat's 前 25% is the digest → session-report chain's
    // proof: every member query on the page has answered by the time it
    // renders.
    const face = await collectFace(page, "前 25%");
    assertFace(fixtureOf("profile-overview", "tested-fps"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});

test("password (mode set) matches the prototype snapshot", async ({ page }) => {
  const tag = `fp${Date.now().toString(36).slice(-7)}s`;
  const { account } = seedBareMember(tag, { phone: "15312347568" });
  try {
    await loginAndOpen(page, account, "/profile/security/password");
    // The bound-phone notice carries the masked number — the security
    // read's proof.
    const face = await collectFace(page, "153****7568");
    assertFace(fixtureOf("password", "set"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});

test("password (mode change) matches the prototype snapshot", async ({ page }) => {
  const tag = `fp${Date.now().toString(36).slice(-7)}c`;
  const { account } = seedBareMember(tag);
  try {
    await loginAndOpen(page, account, "/profile/security/password?mode=change");
    const face = await collectFace(page, "修改登录密码");
    assertFace(fixtureOf("password", "change"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});

test("password (mode forget, rebuilt at /login/forgot) matches the prototype snapshot", async ({
  page,
}) => {
  // The forget face is pre-login by design (G2): a bare visit, no seed,
  // no cookie — the public reset leg's own context.
  await page.goto("/login/forgot");
  await page.waitForLoadState("networkidle");
  const face = await collectFace(page, "重设密码");
  assertFace(fixtureOf("password", "forget"), face, await page.title());
});
