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
const PAGES = ["login", "home?preset=ready", "home?preset=resume", "home?preset=tested-fps"];

test.skip(!process.env.GEN_FIXTURES, "fixture regeneration is opt-in (GEN_FIXTURES=1)");

test("regenerate prototype text fixtures", async ({ browser }) => {
  const outDir = fileURLToPath(new URL("./fixtures/", import.meta.url));
  mkdirSync(outDir, { recursive: true });
  for (const spec of PAGES) {
    const [pageName, query = ""] = spec.split("?");
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
    const file = `${pageName}${preset ? `__${preset}` : ""}.json`;
    writeFileSync(
      `${outDir}${file}`,
      `${JSON.stringify(
        { page: pageName, preset, source: url, title, lines: normalizeLines(raw), placeholders },
        null,
        2,
      )}\n`,
    );
    await context.close();
  }
});
