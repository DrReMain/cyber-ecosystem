import { readFileSync } from "node:fs";
import { expect, test } from "@playwright/test";
import { normalizeLines } from "./text.mjs";

// The oracle corpus is zh-CN and the app negotiates language, so the
// comparison pins the locale instead of inheriting the runner default —
// an en-US runner would diff against the app's English face, not the
// product's target face.
test.use({ locale: "zh-CN" });

interface Fixture {
  page: string;
  preset: string | null;
  source: string;
  title: string;
  lines: string[];
  placeholders: string[];
}

const fixture = JSON.parse(
  readFileSync(new URL("./fixtures/login.json", import.meta.url), "utf8"),
) as Fixture;

// The oracle is the machine snapshot of the prototype page; expected copy
// comes only from that snapshot, never from the implementation, so every
// divergence below is a reported gap, not an adjusted expectation.
test("login text matches the prototype snapshot (line-for-line)", async ({ page }) => {
  await page.goto("/login");
  await page.waitForLoadState("load");
  await page.waitForTimeout(400);
  const { raw, placeholders } = await page.evaluate(() => {
    // Dev chrome never crosses to production faces, so it leaves the
    // oracle comparison too — the same strip the fixture side applies to
    // the prototype's own reviewer gear (REVIEW_CHROME): the breakpoint
    // pill declares itself (data-dev-chrome) and the devtools toggle
    // carries its own fixed label.
    document.querySelectorAll("[data-dev-chrome]").forEach((n) => {
      n.remove();
    });
    for (const b of document.querySelectorAll("button")) {
      if (b.textContent?.includes("TanStack Devtools")) {
        b.remove();
      }
    }
    return {
      // biome-ignore lint/performance/useDomNodeTextContent: innerText is the deliberate oracle — visibility-respecting text is what fidelity asserts; textContent would admit hidden nodes and script bodies.
      raw: document.body.innerText,
      placeholders: [
        ...document.querySelectorAll<HTMLInputElement | HTMLTextAreaElement>("input, textarea"),
      ]
        .filter((el) => el.getClientRects().length > 0)
        .map((el) => el.placeholder)
        .filter(Boolean),
    };
  });
  const actual = normalizeLines(raw);

  const missing = multisetMinus(fixture.lines, actual);
  const extra = multisetMinus(actual, fixture.lines);
  const missingPh = multisetMinus(fixture.placeholders, placeholders);
  const extraPh = multisetMinus(placeholders, fixture.placeholders);
  const sameOrder = JSON.stringify(actual) === JSON.stringify(fixture.lines);
  const appTitle = await page.title();

  expect(
    missing.length === 0 &&
      extra.length === 0 &&
      missingPh.length === 0 &&
      extraPh.length === 0 &&
      sameOrder &&
      appTitle === fixture.title,
    [
      `title: prototype=${JSON.stringify(fixture.title)} app=${JSON.stringify(appTitle)}`,
      `missing lines: ${JSON.stringify(missing)}`,
      `extra lines: ${JSON.stringify(extra)}`,
      `missing placeholders: ${JSON.stringify(missingPh)}`,
      `extra placeholders: ${JSON.stringify(extraPh)}`,
      `orderMatch: ${sameOrder}`,
    ].join("\n"),
  ).toBe(true);
});

function multisetMinus(a: string[], b: string[]): string[] {
  const pool = new Map<string, number>();
  for (const line of b) pool.set(line, (pool.get(line) ?? 0) + 1);
  const out: string[] = [];
  for (const line of a) {
    const n = pool.get(line) ?? 0;
    if (n === 0) out.push(line);
    else pool.set(line, n - 1);
  }
  return out;
}
