import { mkdirSync, writeFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { test } from "@playwright/test";
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
// pre-login-context split). The report family rides the same member-data
// states: tested-fps (single/center/trend — one scored fps session), and
// cognitive at unlocked-single — tested-fps plus the purchased cognitive
// unlock, the preset whose app-side shape is a report_type grant. Career's
// page is state-less copy; ready logs it in without seeding records.
const PAGES = [
  "login",
  "home?preset=ready",
  "home?preset=resume",
  "home?preset=tested-fps",
  "profile-nickname?preset=profile-missing",
  "profile-age?preset=profile-missing",
  "profile-setup?preset=profile-missing",
  "profile-overview?preset=tested-fps",
  "password?preset=ready&m=set#set",
  "password?preset=ready&m=change#change",
  "password?preset=ready&m=forget#forget",
  "report-single?preset=tested-fps",
  "report-center?preset=tested-fps",
  "report-cognitive?preset=unlocked-single",
  "report-trend?preset=tested-fps",
  "report-career?preset=ready",
];

test.skip(!process.env.GEN_FIXTURES, "fixture regeneration is opt-in (GEN_FIXTURES=1)");

test("regenerate prototype text fixtures", async ({ browser }) => {
  const outDir = fileURLToPath(new URL("./fixtures/", import.meta.url));
  mkdirSync(outDir, { recursive: true });
  for (const spec of PAGES) {
    // "#key" names the fixture when the query alone cannot (password's
    // three m= faces share one preset).
    const [specNoKey = spec, key = ""] = spec.split("#");
    const [pageName, query = ""] = specNoKey.split("?");
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
    const title = await page.title();
    // innerText drops input placeholders, so they get their own channel;
    // hidden inputs (e.g. the inactive login tab's field) are excluded to
    // match what innerText does for text nodes.
    const { raw, placeholders } = await page.evaluate(() => ({
      // biome-ignore lint/performance/useDomNodeTextContent: innerText is the deliberate oracle — visibility-respecting text is what fidelity asserts; textContent would admit hidden nodes and script bodies.
      raw: document.body.innerText,
      placeholders: [
        ...document.querySelectorAll<HTMLInputElement | HTMLTextAreaElement>("input, textarea"),
      ]
        .filter((el) => el.getClientRects().length > 0)
        .map((el) => el.placeholder)
        .filter(Boolean),
    }));
    const preset = new URLSearchParams(query).get("preset");
    const suffix = key !== "" ? key : (preset ?? "");
    const file = `${pageName}${suffix ? `__${suffix}` : ""}.json`;
    writeFileSync(
      `${outDir}${file}`,
      `${JSON.stringify(
        // source stays the relative page+query, never the file:// path —
        // the corpus is machine-independent (two dev machines generate it).
        {
          page: pageName,
          preset,
          source: `${pageName}.html${query ? `?${query}` : ""}`,
          title,
          lines: normalizeLines(raw),
          placeholders,
        },
        null,
        2,
      )}\n`,
    );
    await context.close();
  }
});
