import { timestampFromMs } from "@bufbuild/protobuf/wkt";
import { describe, expect, it } from "vitest";
import { resumeDescVars } from "./resume";

// The same local-stamp derivation the card renders (to the minute, the
// prototype's start-moment word), rebuilt here so the expectation stays
// true in any timezone.
const localStamp = (ms: number): string => {
  const d = new Date(ms);
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
};

const DAY = 86_400_000;

describe("resumeDescVars", () => {
  it("reads the RFC 3339 wire string — the shape the JSON transport delivers", () => {
    const created = Date.now() - 3 * DAY - 60_000;
    const vars = resumeDescVars(new Date(created).toISOString(), 1, 4);
    expect(vars).toEqual({ done: 1, total: 4, date: localStamp(created), days: 4 });
  });

  it("reads the declared WKT message through the wkt function", () => {
    const created = Date.now() - 3 * DAY - 60_000;
    const stamp = timestampFromMs(created);
    const vars = resumeDescVars(stamp, 2, 4);
    expect(vars.date).toBe(localStamp(created));
    expect(vars.days).toBe(4);
  });

  it("treats an absent stamp as started today", () => {
    const vars = resumeDescVars(undefined, 0, 4);
    expect(vars.days).toBe(1);
    expect(vars.date).toBe(localStamp(Date.now()));
  });

  it("degrades a malformed wire string to started today, never NaN", () => {
    const vars = resumeDescVars("not-a-date", 0, 4);
    expect(vars.days).toBe(1);
    expect(vars.date).toBe(localStamp(Date.now()));
  });
});
