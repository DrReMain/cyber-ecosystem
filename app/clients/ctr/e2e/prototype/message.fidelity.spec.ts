import { test } from "@playwright/test";
import {
  assertFace,
  cleanup,
  collectFace,
  fixtureOf,
  loginAndOpen,
  seedLetter,
  seedMember,
  seedTicket,
} from "./harness";

// The message family against its oracles: the center's two tabs (the
// letterbox and the member's own feedback queue, ?t= the seg's own
// state), the submit form (stateless — one face), and the detail page's
// two shapes (?kind= in the prototype; two routes in the app — the
// notice letter at /messages/$messageId, the conversation at
// /feedback/$feedbackId). Every face rides seeded rows as its own
// truth: the letterbox's three letters (two unread — the banner's own
// count), the queue's two tickets (a replied conversation with its
// admin turn, a pending one standing on its opening turn alone). No
// clock is pinned in this family — every stamp a face quotes is a row's
// own literal, and no line derives from the client or server clock.
// Runs only against the live dev stack (CTR_E2E=1). Seeds and face
// laws live in ./harness.
//
// Shape law: every seed call sits inside the try, before the page work —
// the pay family's own law (a seed that throws mid-suite must still
// reach the finally's cleanup, or its rows outlive the run; cleanup is
// idempotent DELETEs, so a half-seeded letter cleans up as safely as a
// whole one).
//
// Standing, ledgered divergences keep these specs red on purpose (the
// login spec's precedent):
// · icon law — the prototype's emoji glyphs (the 👤 header fab, the
//   📢 📄 👑 letter glyphs, the ‹ back fabs, the › queue chevrons, the
//   tabbar's 🏠 📊 💬 👤) render as lucide SVG in the app.
// · the msg tab's footer and the notice detail's ruling card — the
//   prototype's own reviewer annotations (both literally point at the ⚙
//   gear: the user_message contract note, the 本期裁决 card); the app
//   ships product copy only, so neither line has a home.
// · the fb tab's state-machine footer — the app carries the lifecycle
//   note on the submit page (that page's own oracle line, verbatim) but
//   not on the center's queue tab.
// · the announcement letter carries no CTA — the oracle's announcement
//   deep-links 去看看我的报告, while the app ties a letter's CTA to its
//   kind (report → the report, feedback_reply → the conversation; an
//   announcement has no ref home). The seeded letter follows the
//   oracle's own type word (系统公告), so the CTA line is the gap.
// · the subject-less ticket — PM-15's registered narrowing (W6 ljr
//   window): the prototype's tickets carry short subjects (匹配总是崩);
//   the product's ticket root IS its first message, so the queue row
//   and the detail's root card read the opening turn as their title
//   line, and the detail echoes that text twice (root card + the
//   thread's first turn) by the same denormalization.
// · the detail pages carry no tabbar — the prototype's feedback-detail
//   is its own outlier (every other detail page declares tab:''; it
//   declares tab:'msg'), and the app's family law mounts the bar on the
//   three root tabs only; both detail kinds ride this account.
// · the conversation's document title — the prototype's static
//   ctr · 详情 against the app's kind-named ctr · 反馈详情 (the notice
//   kind shares 详情 verbatim); honest naming, ledgered here.
test.skip(!process.env.CTR_E2E, "message fidelity needs the live stack (CTR_E2E=1)");
test.use({ locale: "zh-CN" });

// The replied conversation's two turns — the queue face and the detail
// face share them (the ticket's content is its first turn, the
// denormalization the queue rows read).
const CRASH_TURN = "点「开始测评」有的时候会白屏，重进又好了。我用的 iPhone 15 / 微信里打开。";
const REPLY_TURN =
  "已定位：微信内 H5 支付唤起前有概率被系统回收，导致页面白屏。我们已在新版本修复（v1.3.x）。请更新后重试，若仍复现请回复本反馈。";

test("message center (msg tab) matches the prototype snapshot", async ({ page }) => {
  const tag = `fm${Date.now().toString(36).slice(-7)}a`;
  try {
    const { account, userId } = seedMember(tag, "ready");
    // The letterbox's three letters, newest first (the list's own
    // order): two unread — the banner's 2 条未读 is the unread column's
    // own count — and one read membership note.
    seedLetter(tag, userId, {
      suffix: "1",
      type: "announcement",
      title: "赛季因子调整通知",
      content: "新赛季起「反应速度」权重由 1.0 调整为 1.15，历史成绩已按新口径重算。",
      at: "2026-09-13 10:00:00+08",
    });
    seedLetter(tag, userId, {
      suffix: "2",
      type: "report",
      title: "你的认知与潜力评估报告已生成",
      content: "本次覆盖 4 / 6 核心能力，报告含 3 个可渲染维度。",
      at: "2026-09-12 10:00:00+08",
    });
    seedLetter(tag, userId, {
      suffix: "3",
      type: "membership",
      title: "会员权益说明更新",
      content: "「期内解锁的报告永久保留」已在协议 v3 中明确。",
      read: true,
      at: "2026-09-10 10:00:00+08",
    });
    await loginAndOpen(page, account, "/messages");
    const face = await collectFace(page, "2 条未读");
    assertFace(fixtureOf("message-center", "msg"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});

test("message center (fb tab) matches the prototype snapshot", async ({ page }) => {
  const tag = `fm${Date.now().toString(36).slice(-7)}b`;
  try {
    const { account, userId } = seedMember(tag, "ready");
    // The queue's two tickets, most recently active first (updated_at
    // drives both the order and the row's meta stamp): the replied
    // conversation with its admin turn, and a pending one whose whole
    // message is its own title line.
    seedTicket(tag, userId, {
      suffix: "1",
      category: "feature_bug",
      status: "replied",
      content: CRASH_TURN,
      createdAt: "2026-09-08 14:20:00+08",
      updatedAt: "2026-09-08 17:05:00+08",
      turns: [
        { sender: "user", content: CRASH_TURN, at: "2026-09-08 14:20:00+08" },
        { sender: "admin", content: REPLY_TURN, at: "2026-09-08 17:05:00+08" },
      ],
    });
    seedTicket(tag, userId, {
      suffix: "2",
      category: "assessment",
      status: "pending",
      content: "希望增加难度选项",
      createdAt: "2026-09-05 09:00:00+08",
      updatedAt: "2026-09-05 09:00:00+08",
      turns: [{ sender: "user", content: "希望增加难度选项", at: "2026-09-05 09:00:00+08" }],
    });
    await loginAndOpen(page, account, "/messages?t=fb");
    const face = await collectFace(page, "管理员已回复");
    assertFace(fixtureOf("message-center", "fb"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});

test("feedback submit matches the prototype snapshot", async ({ page }) => {
  const tag = `fm${Date.now().toString(36).slice(-7)}c`;
  try {
    const { account } = seedMember(tag, "ready");
    // One stateless form — the member seat is the guard's own need, the
    // face reads nothing else.
    await loginAndOpen(page, account, "/feedback");
    const face = await collectFace(page, "提交反馈");
    assertFace(fixtureOf("feedback-submit", "submit"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});

test("notice detail (announcement letter) matches the prototype snapshot", async ({ page }) => {
  const tag = `fm${Date.now().toString(36).slice(-7)}d`;
  try {
    const { account, userId } = seedMember(tag, "ready");
    // The oracle's own announcement — its meta line names the kind
    // (系统公告), its body is the seeded letter's content whole: the
    // detail renders it whitespace-pre-line, one line per paragraph
    // break. Unread, as delivered: opening the letter is the visit that
    // flips the row server-side.
    const letterId = seedLetter(tag, userId, {
      suffix: "1",
      type: "announcement",
      title: "赛季因子调整通知",
      content:
        "新赛季起，反应速度在综合分中的权重由 1.0 调整为 1.15；决策效率由 1.0 调整为 0.95。\n受影响的调整：\n· 历史测评成绩已按新口径重算综合分（原始分不变）\n· 已解锁的报告保留解锁时的档位快照，不随本次调整变化\n· 未解锁报告与历史最佳（PR）按新口径重新计算\n因子只影响评分层，不改客观数据 —— 你的原始操作记录未做任何修改。",
      at: "2026-09-13 10:00:00+08",
    });
    await loginAndOpen(page, account, `/messages/${letterId}`);
    const face = await collectFace(page, "系统公告 · 2026-09-13");
    assertFace(fixtureOf("feedback-detail", "notice"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});

test("feedback detail (replied conversation) matches the prototype snapshot", async ({ page }) => {
  const tag = `fm${Date.now().toString(36).slice(-7)}e`;
  try {
    const { account, userId } = seedMember(tag, "ready");
    // The replied conversation whole: the root (its content the opening
    // turn — the queue row and this card read the same
    // denormalization), the user turn, and the admin answer whose
    // arrival walked the root to replied.
    const ticketId = seedTicket(tag, userId, {
      suffix: "1",
      category: "feature_bug",
      status: "replied",
      content: CRASH_TURN,
      createdAt: "2026-09-08 14:20:00+08",
      updatedAt: "2026-09-08 17:05:00+08",
      turns: [
        { sender: "user", content: CRASH_TURN, at: "2026-09-08 14:20:00+08" },
        { sender: "admin", content: REPLY_TURN, at: "2026-09-08 17:05:00+08" },
      ],
    });
    await loginAndOpen(page, account, `/feedback/${ticketId}`);
    const face = await collectFace(page, "类型：功能异常");
    assertFace(fixtureOf("feedback-detail", "fb"), face, await page.title());
  } finally {
    cleanup(tag);
  }
});
