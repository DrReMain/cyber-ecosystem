import { create } from "@bufbuild/protobuf";
import type {
  Policy,
  PolicyParams,
} from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/policy_pb";
import {
  CalendarOverrideSchema,
  CalendarParamsSchema,
  PolicyParamsSchema,
  TimeWindowParamsSchema,
  TimeWindowSpanSchema,
} from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/policy_pb";

export type PolicyKind = "time_window" | "calendar";
type CalendarBase = "ALL" | "WEEKDAYS" | "NONE";

export interface WindowDraft {
  start: string; // "HH:MM"
  end: string;
}

export interface TimeWindowDraft {
  windows: WindowDraft[];
  days: number[]; // Go time.Weekday: 0 = Sunday … 6 = Saturday; empty = every day
}

interface OverrideDraft {
  date: string; // "YYYY-MM-DD"
  allowed: boolean;
}

export interface CalendarDraft {
  base: CalendarBase;
  overrides: OverrideDraft[];
}

type PolicyParamsDraft =
  | { kind: "time_window"; timeWindow: TimeWindowDraft }
  | { kind: "calendar"; calendar: CalendarDraft };

export interface PolicyFormDraft {
  name: string;
  kind: PolicyKind;
  timeWindow: TimeWindowDraft;
  calendar: CalendarDraft;
  enabled: boolean;
}

// Sunday-first: the wire order is Go time.Weekday (0 = Sunday), and the
// picker must never present it as Monday-first.
export const DAY_ORDER = [0, 1, 2, 3, 4, 5, 6] as const;

const CALENDAR_BASES: readonly CalendarBase[] = ["ALL", "WEEKDAYS", "NONE"];

const HH_MM = /^([01]\d|2[0-3]):[0-5]\d$/;
const HH_MM_END = /^(([01]\d|2[0-3]):[0-5]\d|24:00)$/;
const DATE = /^\d{4}-\d{2}-\d{2}$/;

// Same-day half-open spans only - cross-midnight schedules are two windows.
// Canonical zero-padded strings compare chronologically, so ordering is a
// plain string check; "24:00" is the end-of-day sentinel and sorts last.
export const isSpanValid = (w: WindowDraft): boolean =>
  HH_MM.test(w.start) && HH_MM_END.test(w.end) && w.start < w.end;

export const isDateValid = (d: string): boolean => DATE.test(d);

export function duplicateDates(overrides: readonly OverrideDraft[]): Set<string> {
  const seen = new Set<string>();
  const dup = new Set<string>();
  for (const o of overrides) {
    if (seen.has(o.date)) dup.add(o.date);
    seen.add(o.date);
  }
  return dup;
}

export function policyIssues(draft: PolicyFormDraft): string[] {
  const issues: string[] = [];
  if (draft.kind === "time_window") {
    if (draft.timeWindow.windows.length === 0) issues.push("tw_required");
    if (draft.timeWindow.windows.some((w) => !isSpanValid(w))) issues.push("tw_invalid_span");
  } else {
    if (duplicateDates(draft.calendar.overrides).size > 0) issues.push("cal_duplicate_date");
    if (draft.calendar.overrides.some((o) => !isDateValid(o.date))) issues.push("cal_invalid_date");
  }
  return issues;
}

const minutes = (hhmm: string): number => Number(hhmm.slice(0, 2)) * 60 + Number(hhmm.slice(3));

function spanCovers(nowMin: number, w: WindowDraft): boolean {
  return nowMin >= minutes(w.start) && nowMin < minutes(w.end);
}

export function dateKey(d: Date): string {
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  return `${d.getFullYear()}-${mm}-${dd}`;
}

// A local mirror of the server plugins for instant form feedback only -
// enforcement always evaluates server-side at request time.
export function passesNow(draft: PolicyParamsDraft, now: Date): boolean {
  if (draft.kind === "time_window") {
    const { windows, days } = draft.timeWindow;
    if (days.length > 0 && !days.includes(now.getDay())) return false;
    const nowMin = now.getHours() * 60 + now.getMinutes();
    return windows.filter(isSpanValid).some((w) => spanCovers(nowMin, w));
  }
  const { base, overrides } = draft.calendar;
  const today = overrides.find((o) => o.date === dateKey(now));
  if (today) return today.allowed;
  if (base === "ALL") return true;
  if (base === "WEEKDAYS") return now.getDay() >= 1 && now.getDay() <= 5;
  return false;
}

export type DaysSummary =
  | { kind: "every" }
  | { kind: "weekdays" }
  | { kind: "weekend" }
  | { kind: "list"; days: number[] };

export function daysSummary(days: readonly number[]): DaysSummary {
  const sorted = [...days].sort((a, b) => a - b);
  const same = (other: readonly number[]): boolean =>
    sorted.length === other.length && sorted.every((v, i) => v === other[i]);
  if (sorted.length === 0) return { kind: "every" };
  if (same([1, 2, 3, 4, 5])) return { kind: "weekdays" };
  if (same([0, 6])) return { kind: "weekend" };
  return { kind: "list", days: sorted };
}

// The transport's flattened toJson outlet hands pages protojson shapes, not
// message instances: the params oneof arrives as { timeWindow: … } /
// { calendar: … } rather than { params: { case, value } }.
interface TimeWindowWire {
  windows?: { start?: string; end?: string }[];
  days?: number[];
}
interface CalendarWire {
  base?: string;
  overrides?: { date?: string; allowed?: boolean }[];
}
interface PolicyWire {
  params?: { timeWindow?: TimeWindowWire; calendar?: CalendarWire };
}

export function draftOf(policy: Policy): PolicyParamsDraft | null {
  const params = (policy as unknown as PolicyWire).params;
  if (params?.timeWindow !== undefined) {
    const tw = params.timeWindow;
    return {
      kind: "time_window",
      timeWindow: {
        windows: (tw.windows ?? []).map((w) => ({ start: w.start ?? "", end: w.end ?? "" })),
        days: [...(tw.days ?? [])],
      },
    };
  }
  if (params?.calendar !== undefined) {
    const cal = params.calendar;
    return {
      kind: "calendar",
      calendar: {
        base: CALENDAR_BASES.includes(cal.base as CalendarBase)
          ? (cal.base as CalendarBase)
          : "ALL",
        overrides: (cal.overrides ?? []).map((o) => ({
          date: o.date ?? "",
          allowed: o.allowed ?? false,
        })),
      },
    };
  }
  return null;
}

export function kindOf(policy: Policy): PolicyKind | null {
  return draftOf(policy)?.kind ?? null;
}

export function paramsDraftOf(draft: PolicyFormDraft): PolicyParamsDraft {
  if (draft.kind === "time_window") return { kind: "time_window", timeWindow: draft.timeWindow };
  return { kind: "calendar", calendar: draft.calendar };
}

// Canonical wire form: all-seven and [] evaluate identically ("every day"),
// so the full set collapses to [] at the save boundary; a subset is deduped
// and weekday-ordered.
const normalizeDays = (days: number[]): number[] => {
  const unique = [...new Set(days)].sort((a, b) => a - b);
  return unique.length === DAY_ORDER.length ? [] : unique;
};

export function paramsInputOf(draft: PolicyParamsDraft): PolicyParams {
  if (draft.kind === "time_window") {
    const params = create(TimeWindowParamsSchema, {
      windows: draft.timeWindow.windows.map((w) =>
        create(TimeWindowSpanSchema, { start: w.start, end: w.end }),
      ),
      days: normalizeDays(draft.timeWindow.days),
    });
    return create(PolicyParamsSchema, { params: { case: "timeWindow", value: params } });
  }
  const params = create(CalendarParamsSchema, {
    base: draft.calendar.base,
    overrides: draft.calendar.overrides.map((o) =>
      create(CalendarOverrideSchema, { date: o.date, allowed: o.allowed }),
    ),
  });
  return create(PolicyParamsSchema, { params: { case: "calendar", value: params } });
}
