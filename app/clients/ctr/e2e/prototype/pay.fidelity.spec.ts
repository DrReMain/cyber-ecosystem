import { test } from "@playwright/test";
import {
  assertFace,
  cleanup,
  collectFace,
  fixtureOf,
  loginAndOpen,
  seedMember,
  seedMembershipUnlock,
  seedMemberWindow,
  seedMobaSession,
  seedPayOrder,
  seedScoredSession,
} from "./harness";

// The pay family against its oracles: the wall's three faces against the
// in-repo P7 draft (the matrix rules it the authority over the prototype
// tree's own pay-confirm — PAY3 built to it; state ① is entered with a
// record ref so the single-unlock option shows, ② is the active-window
// renewal the draft calls its own subject, ③ the lapsed window back on
// dual options), the refusal face against the prototype's underage
// preset, member-detail's two window states against its member /
// member-expired presets, and the receipt's four verdicts against
// pay-result's r= faces — each seeded as the order row its own truth
// reads (paid by scope, closed, a still-pending order as the honest
// failure). The whole family rides one pinned clock, the draft's own
// moment (2026-09-15 21:30): every date the faces quote — the windows'
// end days, the receipt's time line — is a literal at that coordinate.
// The expired-wall face alone moves the clock past its window (the draft
// shows the same 2026-12-15 end as both a future and a past date — only
// the clock says which). Runs only against the live dev stack
// (CTR_E2E=1). Seeds and face laws live in ./harness.
//
// Shape law: every seed call sits inside the try, before the page work —
// a seed that throws mid-suite must still reach the finally's cleanup,
// or its cohort rows outlive the run and poison every later percentile
// face (a leaked pool moved 78 to 79 across the suite once; the rows
// were hunted down by cohort count). Cleanup is idempotent DELETEs, so
// a half-seeded member cleans up as safely as a whole one.
//
// Standing, ledgered divergences keep these specs red on purpose (the
// login spec's precedent):
// · icon law — the draft's and the prototype's emoji glyphs (‹ back fabs,
//   👑 ✅ ⚠️ 🚫 verdict marks, the → and › arrow glyphs) render as lucide
//   SVG in the app.
// · the member receipt's date line — the oracle's 2026-10-16 window is a
//   literal under its own 2026-09 clock but the receipt reads the
//   server-side liveness truth (the real machine clock), where that
//   window has already lapsed: the app answers the bare sentence (its
//   own documented downgrade), never a wrong date. A clock-coordinate
//   gap, not a product one; the overview member face prints the same
//   coordinate gap on its window line by the same law (its window is
//   seeded machine-future for the same reason).
// · the fail face's second sentence — B-8's ruling already removed the
//   「我的 → 订单」 dead pointer from the app copy; the prototype's
//   two-line shape rides the same ledger.
// · the cancel face's note — the app's cancel path closes a real order
//   row (create-then-close), so 「未产生扣款」 is its honest word where
//   the prototype's 「未产生订单」 describes its own no-row cancel.
// · the underage refusal's structure — the app refuses bare (the
//   pay_gate product face: no purchase pitched under the age line, no
//   SKU card, the filing footer stays the wall's own), where the
//   prototype's pay-confirm keeps pitching the single-report card above
//   the refusal; its refund-note footer rides the same structural
//   ledger.
// · the overview member face also carries the profile family's standing
//   gaps (city tail, the progress counter's dash, the D-AT extras card,
//   the ICP 备 audit form) and the report family's sub-ability ledger
//   (the 副能力 fifth axis) — the shared page's own accounts.
test.skip(!process.env.CTR_E2E, "pay fidelity needs the live stack (CTR_E2E=1)");
test.use({ locale: "zh-CN" });

// The family's shared clock: the draft's own moment. The order rows'
// paid_at rides it, and every window end the faces quote is a literal
// under it.
const CLOCK = "2026-09-15T21:30:00+08:00";

test("member detail (active window) matches the prototype snapshot", async ({ page }) => {
  const tag = `fp${Date.now().toString(36).slice(-7)}a`;
  try {
    const { account, userId } = seedMember(tag, "ready");
    seedMemberWindow(tag, userId, "2026-10-15 12:00:00+08");
    // The purple strip's end date is the window row's own literal — the
    // active face's proof, and the CTA flips to 续费 with it.
    await loginAndOpen(page, account, "/member", CLOCK);
    const face = await collectFace(page, "有效期至 2026-10-15");
    assertFace(fixtureOf("member-detail", "member"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});

test("member detail (lapsed window) matches the prototype snapshot", async ({ page }) => {
  const tag = `fp${Date.now().toString(36).slice(-7)}b`;
  try {
    const { account, userId } = seedMember(tag, "ready");
    seedMemberWindow(tag, userId, "2026-09-10 12:00:00+08");
    // The gray strip quotes the end date it lapsed at; the CTA returns
    // to 开通 — expiry is the read-time fact, never a status flip.
    await loginAndOpen(page, account, "/member", CLOCK);
    const face = await collectFace(page, "已过期");
    assertFace(fixtureOf("member-detail", "member-expired"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});

test("pay wall (P7 state 1, dual options with a record ref) matches the draft snapshot", async ({
  page,
}) => {
  const tag = `fp${Date.now().toString(36).slice(-7)}c`;
  try {
    const { account, userId } = seedMember(tag, "ready");
    const { anchorId } = seedScoredSession(tag, userId);
    // The single-unlock option renders only with a record in scope —
    // the ref this arrival carries is the report wall's own subject.
    await loginAndOpen(page, account, `/pay?ref=${anchorId}`, CLOCK);
    const face = await collectFace(page, "选择开通方式");
    assertFace(fixtureOf("pay", "p7-nonmember"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});

test("pay wall (P7 state 2, renewal) matches the draft snapshot", async ({ page }) => {
  const tag = `fp${Date.now().toString(36).slice(-7)}d`;
  try {
    const { account, userId } = seedMember(tag, "ready");
    seedMemberWindow(tag, userId, "2026-12-15 12:00:00+08");
    // The draft's own subject: the standing window hides the single
    // unlock (ruling ④) and leaves renewal only.
    await loginAndOpen(page, account, "/pay", CLOCK);
    const face = await collectFace(page, "确认续费");
    assertFace(fixtureOf("pay", "p7-renewing"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});

test("pay wall (P7 state 3, lapsed back to dual) matches the draft snapshot", async ({ page }) => {
  const tag = `fp${Date.now().toString(36).slice(-7)}e`;
  try {
    const { account, userId } = seedMember(tag, "ready");
    const { anchorId } = seedScoredSession(tag, userId);
    seedMemberWindow(tag, userId, "2026-12-15 12:00:00+08");
    // The same end date the renewal face calls future, read from the
    // other side: the clock moves past the window, the neutral banner
    // announces the lapse, dual options return. The page clock sits past
    // the window while the SSR pass sat before it (the server's own
    // clock), so the hydrating tree rebuilds — a one-off unknown-error
    // toast (the window capture's default sink) fires and lives its
    // normal few seconds; the capture lingers past it, reading the
    // settled face the oracle freezes (transient chrome never enters
    // either side's snapshot).
    await loginAndOpen(page, account, `/pay?ref=${anchorId}`, "2026-12-16T12:00:00+08:00");
    const face = await collectFace(page, "已于 2026-12-15 到期", 5000);
    assertFace(fixtureOf("pay", "p7-expired"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});

test("pay wall (underage refusal) matches the prototype snapshot", async ({ page }) => {
  const tag = `fp${Date.now().toString(36).slice(-7)}f`;
  try {
    const { account } = seedMember(tag, "underage");
    // The gate's own refusal: 13 by year+month at the pinned clock, no
    // purchase pitched under it — the shipped default age line (18)
    // answers from the empty compliance row.
    await loginAndOpen(page, account, "/pay", CLOCK);
    const face = await collectFace(page, "未满 18");
    assertFace(fixtureOf("pay-confirm", "underage"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});

test("pay result (single paid) matches the prototype snapshot", async ({ page }) => {
  const tag = `fp${Date.now().toString(36).slice(-7)}g`;
  try {
    const { account, userId } = seedMember(tag, "ready");
    const orderNo = seedPayOrder(tag, userId, {
      no: "123",
      state: "paid",
      scopeKind: "trial",
      scopeRef: `${tag}r1`,
      amountFen: 1990,
      paid: true,
    });
    // The order row's own truth: paid + trial scope draws the single
    // success, the receipt quoting the row's number, fen, channel, and
    // the pinned clock's moment.
    await loginAndOpen(page, account, `/pay/result?order=${orderNo}`, CLOCK);
    const face = await collectFace(page, "支付成功");
    assertFace(fixtureOf("pay-result", "ok"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});

test("pay result (member paid) matches the prototype snapshot", async ({ page }) => {
  const tag = `fp${Date.now().toString(36).slice(-7)}h`;
  try {
    const { account, userId } = seedMember(tag, "ready");
    seedMemberWindow(tag, userId, "2026-10-16 12:00:00+08");
    const orderNo = seedPayOrder(tag, userId, {
      no: "123",
      state: "paid",
      scopeKind: "period",
      amountFen: 2990,
      paid: true,
    });
    // The member success quotes two facts beyond the row: the sku's own
    // day count and the standing window the purchase opened — the
    // sub line's date and the 查看会员权益 exit ride them.
    await loginAndOpen(page, account, `/pay/result?order=${orderNo}&face=open`, CLOCK);
    const face = await collectFace(page, "会员开通成功");
    assertFace(fixtureOf("pay-result", "member"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});

test("pay result (still pending, the honest failure) matches the prototype snapshot", async ({
  page,
}) => {
  const tag = `fp${Date.now().toString(36).slice(-7)}i`;
  try {
    const { account, userId } = seedMember(tag, "ready");
    const orderNo = seedPayOrder(tag, userId, {
      no: "124",
      state: "pending",
      scopeKind: "trial",
      amountFen: 1990,
    });
    // Nothing settled: a still-pending row is the failure face (the
    // retry exits into the wall with the order's own scope ref).
    await loginAndOpen(page, account, `/pay/result?order=${orderNo}`, CLOCK);
    const face = await collectFace(page, "支付失败");
    assertFace(fixtureOf("pay-result", "fail"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});

test("pay result (closed) matches the prototype snapshot", async ({ page }) => {
  const tag = `fp${Date.now().toString(36).slice(-7)}j`;
  try {
    const { account, userId } = seedMember(tag, "ready");
    const orderNo = seedPayOrder(tag, userId, {
      no: "125",
      state: "closed",
      scopeKind: "trial",
      amountFen: 1990,
    });
    // The cancel leg's own row: closed by the wall's abandon path or
    // expired on its own — the face reads the state, never the path.
    await loginAndOpen(page, account, `/pay/result?order=${orderNo}`, CLOCK);
    const face = await collectFace(page, "已取消支付");
    assertFace(fixtureOf("pay-result", "cancel"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});

test("profile overview (member preset) matches the prototype snapshot", async ({ page }) => {
  const tag = `fp${Date.now().toString(36).slice(-7)}k`;
  try {
    const { account, userId } = seedMember(tag, "ready", {
      phone: "15312347568",
      wechat: "wx_9a2f",
    });
    seedScoredSession(tag, userId);
    seedMobaSession(tag, userId);
    // The window must clear the machine's real clock, not the pinned
    // one: the member card's active form and the moba report's unlock
    // verdict are server-side liveness reads (member_until drops a
    // lapsed window to null), and the oracle's own 2026-10-15 literal
    // is already past the real today — a future end keeps both faces
    // honest, the date line it prints stays a ledgered clock-coordinate
    // gap.
    seedMemberWindow(tag, userId, "2027-06-01 12:00:00+08");
    // The unlock rides the moba anchor: the report verdict's freeze
    // reads it (the cross-genre composite answers the stats' 前 35% off
    // the frozen snapshot) and the digest's all-time count reads the
    // same row —「已解锁 1 份报告」.
    seedMembershipUnlock(tag, userId, `${tag}m4`);
    // The member-data face at its richest: both genres scored (the
    // moba session's RA and DE rows join the cross-genre bests), the
    // standing window turns the member card active, and the one
    // member-window unlock counts.
    await loginAndOpen(page, account, "/profile", CLOCK);
    const face = await collectFace(page, "已解锁 1 份报告");
    assertFace(fixtureOf("profile-overview", "member"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});
