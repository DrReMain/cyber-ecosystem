import { readdirSync, statSync } from "node:fs";
import { join } from "node:path";
import { describe, expect, it } from "vitest";

// The web face of the Go spec_guard: every framework-free logic file in a
// formal-page module (search schemas, row types/builders, domain models —
// the flat .ts files) carries its executable spec beside it. Hooks
// (use-*.ts), server fns (*.fn.ts), and barrels (index.ts) live outside the
// logic face; UI components' proof is the browser loop, not a unit table.
const MODULES_ROOT = join(import.meta.dirname, "features", "modules");

function logicFaces(dir: string): string[] {
  const out: string[] = [];
  for (const entry of readdirSync(dir)) {
    const full = join(dir, entry);
    if (statSync(full).isDirectory()) {
      out.push(...logicFaces(full));
      continue;
    }
    if (!entry.endsWith(".ts") || entry.endsWith(".test.ts") || entry.endsWith(".fn.ts")) continue;
    if (entry === "index.ts" || entry.startsWith("use-")) continue;
    out.push(full);
  }
  return out;
}

describe("spec presence guard (logic faces)", () => {
  it("every logic face carries its table beside it", () => {
    const missing = logicFaces(MODULES_ROOT).filter(
      (f) => !statSync(`${f.slice(0, -3)}.test.ts`, { throwIfNoEntry: false }),
    );
    expect(missing.map((f) => f.replace(MODULES_ROOT, ""))).toEqual([]);
  });
});
