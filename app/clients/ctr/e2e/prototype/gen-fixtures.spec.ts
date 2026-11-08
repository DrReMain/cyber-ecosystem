import { mkdirSync, writeFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { type Page, test } from "@playwright/test";
import { normalizeLines } from "./text.mjs";

// Regenerates the prototype text fixtures — the literal oracle the fidelity
// specs assert against. Opt-in via GEN_FIXTURES=1 (nx target `fixtures`);
// a normal e2e run skips this file, so the oracle only moves on purpose.
// Reviewer-only chrome (version badge, logic-check gear + drawer, env
// switch) is stripped; the bottom tabbar is product UI and stays.
const PROTO =
  process.env.ESPORTS_PROTO_DIR ??
  "/Users/shijiao/Desktop/lab/esports/manage/design-systems/ctr/prototype";
const REVIEW_CHROME = ".vchip, .ckbtn, #__ck, .envsw";
// The wizard pages freeze at profile-missing — the only logged preset whose
// partial profile is exactly the wizard's own subject (guard:'login' bounces
// a blank storage to login.html, and READY-complete members are not this
// face). The overview rides tested-fps (the richest member-data face).
// password's three faces share one shell: the m= param picks the copy, and
// an explicit #key suffix keeps three fixtures apart where the preset alone
// would collide; the forget face's rebuild lives at /login/forgot (G2's
// pre-login-context split). The message family rides static demo data the
// pages own — the faces are URL params alone: message-center's two tabs
// (?t=, the seg's own state) and feedback-detail's two shapes (?kind=
// notice | fb, one page two subjects); feedback-submit is one stateless
// form. All three pages guard login, so preset=ready opens the guard
// without touching the faces (none of them reads preset state).
// The report family rides the same member-data
// states: tested-fps (single/center/trend — one scored fps session), and
// cognitive at unlocked-single — tested-fps plus the purchased cognitive
// unlock, the preset whose app-side shape is a report_type grant. Career's
// page is state-less copy; ready logs it in without seeding records.
// The pay family: member-detail's two window states (member / member-expired
// — the app-side shape one membership entitlement row each), the wall's
// refusal face on pay-confirm at underage, and pay-result's four verdicts —
// r picks the copy, #key keeps four fixtures apart, and the preset is ready
// because a fresh context without one bounces off the login guard (the
// shell applies ?preset before the guard runs). The wall's own three states
// live in the in-repo P7 draft, not the prototype tree — captured by the
// P7 loop below.
const PAGES = [
  "login",
  "home?preset=ready",
  "home?preset=resume",
  "home?preset=tested-fps",
  "profile-nickname?preset=profile-missing",
  "profile-age?preset=profile-missing",
  "profile-setup?preset=profile-missing",
  "profile-overview?preset=tested-fps",
  "profile-overview?preset=member",
  "password?preset=ready&m=set#set",
  "password?preset=ready&m=change#change",
  "password?preset=ready&m=forget#forget",
  "report-single?preset=tested-fps",
  "report-center?preset=tested-fps",
  "report-cognitive?preset=unlocked-single",
  "report-trend?preset=tested-fps",
  "report-career?preset=ready",
  "member-detail?preset=member",
  "member-detail?preset=member-expired",
  "pay-confirm?preset=underage",
  "pay-result?preset=ready&r=ok#ok",
  "pay-result?preset=ready&r=member#member",
  "pay-result?preset=ready&r=fail#fail",
  "pay-result?preset=ready&r=cancel#cancel",
  "message-center?preset=ready&t=msg#msg",
  "message-center?preset=ready&t=fb#fb",
  "feedback-submit?preset=ready#submit",
  "feedback-detail?preset=ready&kind=notice#notice",
  "feedback-detail?preset=ready&kind=fb#fb",
  // The test family rides preset resume — the only preset with an
  // in-flight session (attempts on two of the four lineup games). The
  // running face is captured in its t=0 window: the page's own demo
  // timers first tick at 1000ms, and the loop's 300ms settle lands
  // safely before it, so cd reads its total and the sample counter 0.
  "test-guide?preset=resume#guide",
  "test-intro?preset=resume&ab=reactionSpeed#intro",
  "test-running?preset=resume&ab=reactionSpeed#running",
  "test-result?preset=resume&ab=reactionSpeed#result",
];

// pay-result draws its receipt card off flags.lastOrder — the settle flow's
// own write, which a fresh context never made. The two success faces inject
// the order the prototype's cashier would have written (a fixed number and
// the face's amount; the fail/cancel exits never reach that write and stay
// blank — their fixtures carry no receipt card on purpose).
const LAST_ORDER = new Map([
  ["ok", { no: "PAY20260915123", amount: "19.90", channel: "wx_jsapi" }],
  ["member", { no: "PAY20260915123", amount: "29.90", channel: "wx_jsapi" }],
]);

// The pay wall's oracle is the in-repo P7 draft (matrix 24 rules it the
// authority over the prototype tree's own pay-confirm; PAY3 built to it):
// three states on one page, tab-switched by show(i). The tab strip is the
// draft's own reviewer chrome and is stripped before capture; the phone
// frame is the draft's presentation, its inner text is the face.
const P7_DRAFT = fileURLToPath(
  new URL("../../../../../docs/roadmap/esports/p7-member-paywall.html", import.meta.url),
);
const P7_STATES = [
  { key: "p7-nonmember", show: 0 },
  { key: "p7-renewing", show: 1 },
  { key: "p7-expired", show: 2 },
];

test.skip(!process.env.GEN_FIXTURES, "fixture regeneration is opt-in (GEN_FIXTURES=1)");

// innerText drops input placeholders, so they get their own channel; hidden
// inputs (e.g. the inactive login tab's field) are excluded to match what
// innerText does for text nodes. One capture shape serves every source —
// prototype pages and the P7 draft alike.
async function collect(page: Page): Promise<{ raw: string; placeholders: string[] }> {
  return page.evaluate(() => ({
    // biome-ignore lint/performance/useDomNodeTextContent: innerText is the deliberate oracle — visibility-respecting text is what fidelity asserts; textContent would admit hidden nodes and script bodies.
    raw: document.body.innerText,
    placeholders: [
      ...document.querySelectorAll<HTMLInputElement | HTMLTextAreaElement>("input, textarea"),
    ]
      .filter((el) => el.getClientRects().length > 0)
      .map((el) => el.placeholder)
      .filter(Boolean),
  }));
}

function writeFixture(
  outDir: string,
  face: {
    page: string;
    // the filename suffix: the caller's disambiguator when the query alone
    // cannot name the face (password m=, pay-result r=, the P7 states),
    // else the URL preset — the corpus's own naming, unchanged.
    key: string;
    preset: string | null;
    source: string;
    title: string;
  },
  captured: { raw: string; placeholders: string[] },
): void {
  const suffix = face.key !== "" ? face.key : (face.preset ?? "");
  writeFileSync(
    `${outDir}${face.page}${suffix ? `__${suffix}` : ""}.json`,
    `${JSON.stringify(
      // source stays the relative page+query, never the file:// path —
      // the corpus is machine-independent (two dev machines generate it).
      {
        page: face.page,
        preset: face.preset,
        source: face.source,
        title: face.title,
        lines: normalizeLines(captured.raw),
        placeholders: captured.placeholders,
      },
      null,
      2,
    )}\n`,
  );
}

test("regenerate prototype text fixtures", async ({ browser }) => {
  const outDir = fileURLToPath(new URL("./fixtures/", import.meta.url));
  mkdirSync(outDir, { recursive: true });
  for (const spec of PAGES) {
    // "#key" names the fixture when the query alone cannot (password's
    // three m= faces, pay-result's four r= verdicts).
    const [specNoKey = spec, key = ""] = spec.split("#");
    const [pageName = "", query = ""] = specNoKey.split("?");
    const url = `file://${PROTO}/${pageName}.html${query ? `?${query}` : ""}`;
    const context = await browser.newContext(); // fresh storage per page
    const page = await context.newPage();
    await page.goto(url, { waitUntil: "load" });
    await page.waitForTimeout(300);
    await page.evaluate((sel) => {
      document.querySelectorAll(sel).forEach((n) => {
        n.remove();
      });
    }, REVIEW_CHROME);
    const order = LAST_ORDER.get(key);
    if (order !== undefined) {
      // The receipt card renders at load off flags.lastOrder — a fact the
      // settle flow writes, which a fresh context never made. CTR.set
      // persists it, but the page already painted; the reload re-renders
      // off the stored state. The preset param goes first: the shell's
      // reset() would re-apply the preset over it, while a plain load
      // restores the stored state whole — logged flag included, so the
      // login guard still clears.
      await page.evaluate((o) => {
        // CTR is the prototype's own page global — the shell's state pen;
        // typed through the window because tsc never sees the page's
        // script scope.
        (window as unknown as { CTR: { set: (key: string, value: unknown) => void } }).CTR.set(
          "flags.lastOrder",
          o,
        );
        const u = new URL(location.href);
        u.searchParams.delete("preset");
        history.replaceState(null, "", u);
      }, order);
      await page.reload({ waitUntil: "load" });
      await page.waitForTimeout(300);
      await page.evaluate((sel) => {
        document.querySelectorAll(sel).forEach((n) => {
          n.remove();
        });
      }, REVIEW_CHROME);
    }
    const title = await page.title();
    const captured = await collect(page);
    writeFixture(
      outDir,
      {
        page: pageName,
        key,
        preset: new URLSearchParams(query).get("preset"),
        source: `${pageName}.html${query ? `?${query}` : ""}`,
        title,
      },
      captured,
    );
    await context.close();
  }

  for (const state of P7_STATES) {
    const context = await browser.newContext();
    const page = await context.newPage();
    await page.goto(`file://${P7_DRAFT}`, { waitUntil: "load" });
    await page.waitForTimeout(300);
    // The tab strip is reviewer chrome (the same law as REVIEW_CHROME);
    // then show(i) picks the state before the capture.
    await page.evaluate((idx) => {
      document.querySelector(".tabs")?.remove();
      // show(i) is the draft's own tab switcher — same page-global law
      // as CTR above.
      (window as unknown as { show: (i: number) => void }).show(idx);
    }, state.show);
    const title = await page.title();
    writeFixture(
      outDir,
      {
        page: "pay",
        key: state.key,
        preset: null,
        source: `p7-member-paywall.html#state${state.show}`,
        title,
      },
      await collect(page),
    );
    await context.close();
  }

  // The test-done face (prototype test-done.html): its own script
  // CONSUMES the storage session (finishSession folds it into a record
  // the moment the page runs), so the first load with ?preset=resume
  // would fold the preset's own half-played session. The face the
  // corpus wants is the fully-played one — a session whose every
  // lineup game carries an attempt — injected the same way the P7
  // states inject theirs: write it, drop the preset param (reset()
  // would re-apply the preset over it on the next load), reload, and
  // the page folds the injected truth instead. The attempt values are
  // the percentiles the face prints (the prototype's own demo
  // semantics); its score lines derive them as round(0.9·pct+5).
  const doneContext = await browser.newContext();
  const donePage = await doneContext.newPage();
  await donePage.goto(`file://${PROTO}/test-done.html?preset=resume`, {
    waitUntil: "load",
  });
  await donePage.waitForTimeout(300);
  await donePage.evaluate(() => {
    (window as unknown as { CTR: { set: (key: string, value: unknown) => void } }).CTR.set(
      "session",
      {
        genre: "fps",
        steps: [
          { ab: "reactionSpeed", code: "red_dot" },
          { ab: "aimingAccuracy", code: "flash_aiming" },
          { ab: "trackingStability", code: "moving_target" },
          { ab: "skillReleaseAccuracy", code: "skill_trajectory" },
        ],
        done: [
          "reactionSpeed",
          "aimingAccuracy",
          "trackingStability",
          "skillReleaseAccuracy",
        ],
        retestLeft: { red_dot: 1, flash_aiming: 2, moving_target: 3, skill_trajectory: 3 },
        attempts: { red_dot: [78], flash_aiming: [86], moving_target: [71], skill_trajectory: [64] },
        startedAt: "2026-09-15 21:04",
        resumable: true,
      },
    );
    const u = new URL(location.href);
    u.searchParams.delete("preset");
    history.replaceState(null, "", u);
  });
  await donePage.reload({ waitUntil: "load" });
  await donePage.waitForTimeout(300);
  await donePage.evaluate((sel) => {
    document.querySelectorAll(sel).forEach((n) => {
      n.remove();
    });
  }, REVIEW_CHROME);
  writeFixture(
    outDir,
    {
      page: "test-done",
      key: "done",
      preset: null,
      source: "test-done.html#injected",
      title: await donePage.title(),
    },
    await collect(donePage),
  );
  await doneContext.close();
});
