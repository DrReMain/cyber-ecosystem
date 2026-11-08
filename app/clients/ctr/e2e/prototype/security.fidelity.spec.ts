import { test } from "@playwright/test";
import {
  assertFace,
  cleanup,
  collectFace,
  fixtureOf,
  loginAndOpen,
  seedBareMember,
  seedMember,
  seedReportGrant,
  seedTestSession,
} from "./harness";

// The security family against its oracles: the hub (/profile/security —
// three binding rows over the phone-unbound mirror, wechat bound and
// phone free), the bind-phone page's plain and occupied faces (the
// occupied card is the app's own preview verdict, reached by really
// sending the number at a member who holds it), the bind-wechat pair
// (the occupied face rides ?failed=true — the URL's own word form, the
// router's JSON-decoding search parser turns =1 into a number the
// validator drops), and the merge confirm page's two faces — the
// corpus's pair (empty absorbed into rich) and its blocked pane. The
// merge page's offer is an in-memory one-shot handoff (merge-flow.ts:
// the verdict can neither ride the URL nor be re-derived), so its two
// faces arrive by the real chain: bind-phone, number, send (the preview
// runs before any SMS is spent), the conflict card, its CTA. The rich
// caller that survives needs esports weight — one scored session and a
// report grant — while the absorbed side is the passwordless member
// seedBareMember mints (no password, one sms credential: the mergeable
// bar's system half). No clock is pinned: no face quotes a clock.
// Runs only against the live dev stack (CTR_E2E=1). Seeds and face
// laws live in ./harness.
//
// Shape law: every seed call sits inside the try, before the page work
// (the pay family's own law — cleanup is idempotent DELETEs, so a
// half-seeded session cleans up as safely as a whole one).
//
// Standing, ledgered divergences keep these specs red on purpose (the
// login spec's precedent):
// · icon law — the hub's 💬 📱 🔑 glyph tiles and the bind-wechat
//   pitch's 💬 box render as lucide SVG in the app, and the ‹ back fabs
//   likewise (the family account).
// · status chip vs action word — the prototype's unbound rows wear the
//   state word (未绑定 chip plus the › affordance); the app renders a
//   去绑定 action link where it can act (the row's own grammar).
// · the password row's model — every prototype preset leaves pwdSet
//   unset (the wechat-born member the corpus presumes), so the oracle
//   prints the unset row; the harness member logs in by password, so
//   the app reads the set row — set desc, 修改 link. The passwordless
//   member this diff would need cannot hold a session.
// · the wechat-bound desc — the oracle prefixes the openId word and
//   masks with dot bullets (openId wx_••••9a2f · 可用于一键登录); the
//   app renders the server's own mask (star block) and its fuller
//   可用于微信一键登录.
// · the unbind note — the prototype promises self-service unbind from
//   each detail page; the app's ruling is the service-desk word (those
//   detail pages do not exist by ruling).
// · the birth row — 2004-06 against the oracle's 2008-06: the cohort
//   key is birth year + sex, and the shared dev pool's 2008 calibration
//   batch makes that year un-ownable (the profile family's own law).
// · the phone placeholder — the oracle's bare 手机号 against the app's
//   手机号（11 位）.
// · the bind-phone conflict card — the app's card carries the merge
//   CTA the prototype's occupied card has not (检测到全新账号，去合并 /
//   为什么不能自动合并？查看说明), and folds the oracle's third line's
//   advice into its body paragraph (the oracle breaks it into its own
//   line).
// · the bind-wechat pitch — the oracle's <br> splits the promise into
//   two lines; the app carries it as one paragraph.
// · the bind-wechat occupied card — the oracle's 目标账号已被占用 lead
//   and its 建议 advice line have no app counterpart (the app's body
//   opens with the comma that follows the head), the primary button
//   reads 用该微信登录 where the oracle swaps in 重新授权, and the app
//   adds the conflict note line.
// · the merge seg — the prototype's demo switcher (空号并入富号 /
//   双方均有数据) is reviewer-facing state; the app derives the state
//   from the offer, so both seg lines are missing on every face.
// · the merge rich label — the oracle's demo pair names the rich side
//   by its phone mask (153****7568); the app's chain arrives from
//   bind-phone, whose caller is by definition phone-less, so the rich
//   tile names its wechat leg (微信授权). The keep-other direction that
//   would name a rich phone cannot hold: its caller-side mergeable bar
//   fails on the very password the session runs on.
// · the merge proof card — the oracle's completed-tense proofs (已通过
//   短信验证码校验 / openid 一致) describe the demo's double gate; the
//   app's sms leg states what it will do (已向空号手机发送验证码…),
//   drops the openid line (that proof belongs to the wechat leg), and
//   renders its proof inputs — the phone and code fields with their
//   placeholders and the send button — four extra rows plus the two
//   extra placeholders. 两项校验通过才允许合并 against 两项都过才允许合并.
// · the blocked pane's manual line — the oracle's own typo (如需合请走
//   人工) prints as written; the app carries the corrected 如需合并请走人工.
// · the hub's tabbar — the prototype renders the shell's five tab rows
//   on the hub (its own tab: 'me'); the app's security tree is a
//   headered subpage with no tabbar. A structural account, five rows.
// · the hub's sex row — a real app bug the face caught: the wire's sex
//   is the proto-JSON enum name (the age step normalizes it, its own
//   comment says so), but the hub compares === 1, so the row reads
//   未填 for every member. The oracle prints 男.
// · the hub's games row — the wizard resolves catalog codes to names
//   (listGames); the hub renders the stored codes raw (cs2、honor、lol
//   against 王者荣耀、英雄联盟、CS2). A display-shape account, bug-flagged
//   beside the sex row for a fix slice of their own.

test.skip(!process.env.CTR_E2E, "security fidelity needs the live stack (CTR_E2E=1)");
test.use({ locale: "zh-CN" });

// The hub's subject: the phone-unbound mirror — the wechat credential
// in, the sms credential out — over the ready profile. The password row
// reads the session's own truth either way (the ledger's model gap).
function seedWechatMember(tag: string): { account: string; userId: string } {
  return seedMember(tag, "ready", { wechat: "wx_openid9a2f" });
}

test("account security hub (wechat-bound, phone-free) matches the prototype snapshot", async ({
  page,
}) => {
  const tag = `fs${Date.now().toString(36).slice(-7)}a`;
  try {
    const { account } = seedWechatMember(tag);
    await loginAndOpen(page, account, "/profile/security");
    // The nickname row's value proves the profile query landed.
    const face = await collectFace(page, "龘好久");
    assertFace(fixtureOf("account-security", "security"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});

test("bind-phone (plain face) matches the prototype snapshot", async ({ page }) => {
  const tag = `fs${Date.now().toString(36).slice(-7)}b`;
  try {
    const { account } = seedWechatMember(tag);
    // from=guide names the entry word the oracle's own default prints
    // (完善资料); the app's bare-URL default reads the me entry instead —
    // the word is passed, not assumed.
    await loginAndOpen(page, account, "/profile/security/bind-phone?from=guide");
    // The entry line rides both member queries (the method half reads
    // credentials, the nickname half the profile).
    const face = await collectFace(page, "入口：");
    assertFace(fixtureOf("bind-phone", "phone"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});

test("bind-phone (occupied card) matches the prototype snapshot", async ({ page }) => {
  const tag = `fs${Date.now().toString(36).slice(-7)}c`;
  const holderTag = `${tag}h`;
  try {
    seedWechatMember(tag);
    // The member who holds the target number and carries esports weight:
    // both sides heavy, so the preview answers BLOCKED — the card the
    // oracle's err=conflict face freezes. The preview runs before any
    // SMS is spent, so the click leaves no side effects behind.
    const { userId: holderId } = seedMember(holderTag, "ready", { phone: "13900002214" });
    seedTestSession(holderTag, holderId, {
      startedAt: "2026-09-15 21:04:00+08",
      records: [
        {
          suffix: "r1",
          code: "red_dot",
          attempt: 1,
          score: 75,
          grade: "B",
          at: "2026-09-15 21:05:00+08",
        },
      ],
    });
    await loginAndOpen(page, `${tag}a`, "/profile/security/bind-phone?from=guide");
    await collectFace(page, "入口：");
    await page.getByLabel("手机号").fill("13900002214");
    await page.getByRole("button", { name: "发送验证码" }).click();
    const face = await collectFace(page, "该手机号已绑定其他账号");
    assertFace(fixtureOf("bind-phone", "phone-conflict"), face, await page.title());
  } finally {
    cleanup(tag);
    cleanup(holderTag);
  }
});

test("bind-wechat (plain face) matches the prototype snapshot", async ({ page }) => {
  const tag = `fs${Date.now().toString(36).slice(-7)}d`;
  try {
    // The phone-carrying member — the ready preset's own subject: the
    // entry line names the account by its phone mask (153****7568).
    seedMember(tag, "ready", { phone: "15300007568" });
    await loginAndOpen(page, `${tag}a`, "/profile/security/bind-wechat?from=guide");
    const face = await collectFace(page, "入口：");
    assertFace(fixtureOf("bind-wechat", "wechat"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});

test("bind-wechat (occupied card) matches the prototype snapshot", async ({ page }) => {
  const tag = `fs${Date.now().toString(36).slice(-7)}e`;
  try {
    seedMember(tag, "ready", { phone: "15300007568" });
    // failed=true is the URL's honest word form — the router's search
    // parser JSON-decodes =1 into a number and the validator drops it
    // (the review param's own law).
    await loginAndOpen(page, `${tag}a`, "/profile/security/bind-wechat?failed=true&from=guide");
    const face = await collectFace(page, "该微信已绑定其他账号");
    assertFace(fixtureOf("bind-wechat", "wechat-conflict"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});

test("merge confirm (empty into rich) matches the prototype snapshot", async ({ page }) => {
  const tag = `fs${Date.now().toString(36).slice(-7)}f`;
  const emptyTag = `${tag}e`;
  try {
    // The rich caller: one scored session plus a report grant — the
    // summary line prints 1 场记录 · 报告已解锁 exactly. The absorbed
    // side is the passwordless holder of 13900002214: no password, one
    // sms credential, no esports rows — the mergeable bar's empty side.
    const { userId } = seedWechatMember(tag);
    seedReportGrant(tag, "cognitive");
    seedTestSession(tag, userId, {
      startedAt: "2026-09-15 21:04:00+08",
      records: [
        {
          suffix: "r1",
          code: "red_dot",
          attempt: 1,
          score: 75,
          grade: "B",
          at: "2026-09-15 21:05:00+08",
        },
      ],
    });
    seedBareMember(emptyTag, { phone: "13900002214", passwordless: true });
    await loginAndOpen(page, `${tag}a`, "/profile/security/bind-phone?from=guide");
    await collectFace(page, "入口：");
    await page.getByLabel("手机号").fill("13900002214");
    await page.getByRole("button", { name: "发送验证码" }).click();
    await collectFace(page, "该手机号已绑定其他账号");
    // The card's CTA hands the offer to the merge page — the in-memory
    // one-shot handoff is the only road in.
    await page.getByRole("button", { name: "检测到全新账号，去合并" }).click();
    const face = await collectFace(page, "将被注销（空号）");
    assertFace(fixtureOf("account-merge-confirm", "merge"), face, await page.title());
  } finally {
    cleanup(tag);
    cleanup(emptyTag);
  }
});

test("merge blocked (both sides weighted) matches the prototype snapshot", async ({ page }) => {
  const tag = `fs${Date.now().toString(36).slice(-7)}g`;
  const otherTag = `${tag}o`;
  try {
    // Both sides carry weight: the rich caller above against a holder
    // with his own scored session — the preview answers BLOCKED, and
    // the CTA routes to the manual-path pane.
    const { userId } = seedWechatMember(tag);
    seedReportGrant(tag, "cognitive");
    seedTestSession(tag, userId, {
      startedAt: "2026-09-15 21:04:00+08",
      records: [
        {
          suffix: "r1",
          code: "red_dot",
          attempt: 1,
          score: 75,
          grade: "B",
          at: "2026-09-15 21:05:00+08",
        },
      ],
    });
    const { userId: otherId } = seedMember(otherTag, "ready", { phone: "13900002214" });
    seedTestSession(otherTag, otherId, {
      startedAt: "2026-09-15 20:04:00+08",
      records: [
        {
          suffix: "r1",
          code: "red_dot",
          attempt: 1,
          score: 71,
          grade: "B",
          at: "2026-09-15 20:05:00+08",
        },
      ],
    });
    await loginAndOpen(page, `${tag}a`, "/profile/security/bind-phone?from=guide");
    await collectFace(page, "入口：");
    await page.getByLabel("手机号").fill("13900002214");
    await page.getByRole("button", { name: "发送验证码" }).click();
    await collectFace(page, "该手机号已绑定其他账号");
    await page.getByRole("button", { name: "为什么不能自动合并？查看说明" }).click();
    const face = await collectFace(page, "不可自动合并");
    assertFace(fixtureOf("account-merge-confirm", "block"), face, await page.title());
  } finally {
    cleanup(tag);
    cleanup(otherTag);
  }
});
