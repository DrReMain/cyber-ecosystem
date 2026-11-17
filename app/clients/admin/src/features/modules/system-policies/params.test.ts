import { toJson } from "@bufbuild/protobuf";
import type { Policy } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/policy_pb";
import { PolicyParamsSchema } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/policy_pb";
import fc from "fast-check";
import { describe, expect, it } from "vitest";

import {
  dateKey,
  daysSummary,
  draftOf,
  duplicateDates,
  isSpanValid,
  paramsInputOf,
  passesNow,
  policyIssues,
} from "./params";

// The ABAC params face — the client-side mirror of the server plugins
// (passesNow), the span/date validation vocabulary, and the draft↔wire
// round-trip over the transport's flattened toJson shape.

const hhmm = (m: number): string =>
  `${String(Math.floor(m / 60)).padStart(2, "0")}:${String(m % 60).padStart(2, "0")}`;

describe("isSpanValid", () => {
  it("accepts exactly the same-day half-open spans, '24:00' as the end sentinel", () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 0, max: 1439 }),
        fc.integer({ min: 1, max: 1440 }),
        (start, len) => {
          const end = Math.min(start + len, 1440);
          const w = { start: hhmm(start), end: end === 1440 ? "24:00" : hhmm(end) };
          expect(isSpanValid(w)).toBe(true);
        },
      ),
    );
  });

  it("rejects inverted, empty, and malformed spans", () => {
    for (const w of [
      { start: "17:00", end: "09:00" },
      { start: "09:00", end: "09:00" },
      { start: "9:00", end: "17:00" },
      { start: "09:00", end: "25:00" },
    ]) {
      expect(isSpanValid(w)).toBe(false);
    }
  });
});

describe("duplicateDates", () => {
  it("collects exactly the dates seen more than once, once each", () => {
    fc.assert(
      fc.property(fc.array(fc.date({ noInvalidDate: true }), { maxLength: 20 }), (dates) => {
        const keys = dates.map(dateKey);
        const dups = duplicateDates(keys.map((date) => ({ date, allowed: true })));
        const expected = new Set(keys.filter((k, i) => keys.indexOf(k) !== i));
        expect(dups).toEqual(expected);
      }),
    );
  });
});

describe("policyIssues", () => {
  const draft = (over: Record<string, unknown>): Record<string, unknown> => ({
    name: "p",
    kind: "time_window",
    timeWindow: { windows: [{ start: "09:00", end: "17:00" }], days: [] },
    calendar: { base: "ALL", overrides: [] },
    enabled: true,
    ...over,
  });

  it("is clean for a valid time-window draft", () => {
    expect(policyIssues(draft({}) as never)).toEqual([]);
  });

  it("flags each defect class by its issue key", () => {
    expect(policyIssues(draft({ timeWindow: { windows: [], days: [] } }) as never)).toContain(
      "tw_required",
    );
    expect(
      policyIssues(
        draft({ timeWindow: { windows: [{ start: "18:00", end: "09:00" }], days: [] } }) as never,
      ),
    ).toContain("tw_invalid_span");
    expect(
      policyIssues(
        draft({
          kind: "calendar",
          calendar: {
            base: "ALL",
            overrides: [
              { date: "2026-01-01", allowed: false },
              { date: "2026-01-01", allowed: true },
            ],
          },
        }) as never,
      ),
    ).toContain("cal_duplicate_date");
    expect(
      policyIssues(
        draft({
          kind: "calendar",
          calendar: { base: "ALL", overrides: [{ date: "2026-1-1", allowed: false }] },
        }) as never,
      ),
    ).toContain("cal_invalid_date");
  });
});

describe("passesNow (the server plugins' local mirror)", () => {
  it("is exactly the half-open interval test, windows OR", () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 0, max: 700 }),
        fc.integer({ min: 1, max: 719 }),
        fc.integer({ min: 720, max: 1438 }),
        fc.integer({ min: 1, max: 719 }),
        fc.integer({ min: 0, max: 1439 }),
        (s1, l1, s2, l2, m) => {
          const e1 = s1 + l1;
          const e2 = Math.min(s2 + l2, 1440);
          const now = new Date(2026, 2, 10, Math.floor(m / 60), m % 60);
          const got = passesNow(
            {
              kind: "time_window",
              timeWindow: {
                days: [],
                windows: [
                  { start: hhmm(s1), end: hhmm(e1) },
                  { start: hhmm(s2), end: e2 === 1440 ? "24:00" : hhmm(e2) },
                ],
              },
            },
            now,
          );
          expect(got).toBe((m >= s1 && m < e1) || (m >= s2 && m < e2));
        },
      ),
    );
  });

  it("rides the JS weekday convention (0 = Sunday) for the days gate", () => {
    fc.assert(
      fc.property(fc.date({ noInvalidDate: true }), (now) => {
        const allDay = {
          kind: "time_window" as const,
          timeWindow: { windows: [{ start: "00:00", end: "24:00" }], days: [now.getDay()] },
        };
        expect(passesNow(allDay, now)).toBe(true);
        const shifted = {
          ...allDay,
          timeWindow: { ...allDay.timeWindow, days: [(now.getDay() + 1) % 7] },
        };
        expect(passesNow(shifted, now)).toBe(false);
      }),
    );
  });

  it("calendar: override wins, then the base speaks", () => {
    fc.assert(
      fc.property(
        fc.date({ noInvalidDate: true }),
        fc.boolean(),
        fc.constantFrom("ALL", "WEEKDAYS", "NONE"),
        (now, allowed, base) => {
          const withOverride = passesNow(
            { kind: "calendar", calendar: { base, overrides: [{ date: dateKey(now), allowed }] } },
            now,
          );
          expect(withOverride).toBe(allowed);
          const noHit = passesNow({ kind: "calendar", calendar: { base, overrides: [] } }, now);
          expect(noHit).toBe(
            base === "ALL" || (base === "WEEKDAYS" && now.getDay() >= 1 && now.getDay() <= 5),
          );
        },
      ),
    );
  });
});

describe("daysSummary", () => {
  it("classifies the four states with canonical ordering", () => {
    expect(daysSummary([])).toEqual({ kind: "every" });
    expect(daysSummary([5, 1, 4, 2, 3])).toEqual({ kind: "weekdays" });
    expect(daysSummary([6, 0])).toEqual({ kind: "weekend" });
    expect(daysSummary([1, 3])).toEqual({ kind: "list", days: [1, 3] });
  });
});

describe("draft ↔ wire round-trip", () => {
  const daysArb = fc.uniqueArray(fc.integer({ min: 0, max: 6 }), { maxLength: 6 });
  const twDraftArb = daysArb.chain((days) =>
    fc
      .array(fc.tuple(fc.integer({ min: 0, max: 1300 }), fc.integer({ min: 1, max: 139 })), {
        minLength: 1,
        maxLength: 4,
      })
      .map((spans) => ({
        kind: "time_window" as const,
        timeWindow: {
          days,
          windows: spans.map(([s, l]) => ({
            start: hhmm(s),
            end: hhmm(Math.min(s + l, 1439) + 1),
          })),
        },
      })),
  );
  const calDraftArb = fc
    .array(
      fc.tuple(
        fc.integer({ min: 1, max: 9999 }),
        fc.integer({ min: 1, max: 12 }),
        fc.integer({ min: 1, max: 28 }),
        fc.boolean(),
      ),
      { maxLength: 5 },
    )
    .map((rows) => ({
      kind: "calendar" as const,
      calendar: {
        base: "WEEKDAYS" as const,
        overrides: rows.map(([y, mo, d, allowed]) => ({
          date: `${y}-${String(mo).padStart(2, "0")}-${String(d).padStart(2, "0")}`,
          allowed,
        })),
      },
    }));

  it("time_window: paramsInputOf → toJson wire → draftOf is the identity (full week folds to [])", () => {
    fc.assert(
      fc.property(twDraftArb, (draft) => {
        const message = paramsInputOf(draft);
        const wire = { params: toJson(PolicyParamsSchema, message) } as unknown as Policy;
        const back = draftOf(wire);
        expect(back).toEqual({
          kind: "time_window",
          timeWindow: {
            days: [...new Set(draft.timeWindow.days)].sort((a, b) => a - b),
            windows: draft.timeWindow.windows,
          },
        });
      }),
    );
  });

  it("calendar: the same round-trip holds", () => {
    fc.assert(
      fc.property(calDraftArb, (draft) => {
        const message = paramsInputOf(draft);
        const wire = { params: toJson(PolicyParamsSchema, message) } as unknown as Policy;
        expect(draftOf(wire)).toEqual(draft);
      }),
    );
  });
});
