import { create } from "@bufbuild/protobuf";
import { TimestampSchema } from "@bufbuild/protobuf/wkt";
import { renderHook } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { TimeLocaleContext, useColumns } from "./use-columns";

const ts = (seconds: number) => create(TimestampSchema, { seconds: BigInt(seconds), nanos: 0 });
const pad = (n: number) => String(n).padStart(2, "0");
// Stamps render in the viewer's local timezone; build expectations from the
// same local getters so the suite is timezone-portable.
const localStamp = (ms: number) => {
  const d = new Date(ms);
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
};

describe("useColumns factories", () => {
  it("fieldTimestamp renders fixed-shape stamps from proto timestamps", () => {
    const { result } = renderHook(() => useColumns());
    const col = result.current.fieldTimestamp<Record<string, unknown>>("createdAt", {
      title: "创建",
    });
    expect(col.title).toBe("创建");
    expect(col.dataIndex).toBe("createdAt");
    expect(col.render?.(ts(0), {}, 0)).toBe(localStamp(0));
  });

  it("fieldTimestamp keeps array dataIndex for nested paths", () => {
    const { result } = renderHook(() => useColumns());
    const col = result.current.fieldTimestamp<Record<string, unknown>>(["user", "createdAt"], {
      title: "创建",
    });
    expect(col.dataIndex).toEqual(["user", "createdAt"]);
    const fallback = result.current.fieldTimestamp<Record<string, unknown>>(["user", "createdAt"]);
    expect(fallback.title).toBe("user.createdAt");
  });

  it("fieldTimestamp falls back to '-' for empty values", () => {
    const { result } = renderHook(() => useColumns());
    const col = result.current.fieldTimestamp<Record<string, unknown>>("x");
    expect(col.render?.(undefined, {}, 0)).toBe("-");
    expect(col.render?.(null, {}, 0)).toBe("-");
  });

  it("fieldTimestamp accepts epoch numbers and date strings", () => {
    const { result } = renderHook(() => useColumns());
    const col = result.current.fieldTimestamp<Record<string, unknown>>("x");
    expect(col.render?.(86_400_000, {}, 0)).toBe(localStamp(86_400_000));
    expect(typeof col.render?.("1970-01-01T00:00:00Z", {}, 0)).toBe("string");
  });

  it("fieldAction requires an explicit title and pins right", () => {
    const { result } = renderHook(() => useColumns());
    const col = result.current.fieldAction<Record<string, unknown>>(() => null, { title: "操作" });
    expect(col.title).toBe("操作");
    expect(col.fixed).toBe("right");
  });

  it("fieldCopy renders copyable text", () => {
    const { result } = renderHook(() => useColumns());
    const col = result.current.fieldCopy<Record<string, unknown>>("id", { title: "ID" });
    expect(col.title).toBe("ID");
  });

  it("formatTime composes the date and time axes", () => {
    const { result } = renderHook(() => useColumns());
    const ms = Date.UTC(2026, 8, 24, 5, 6);
    const d = new Date(ms);
    expect(result.current.formatTime(ms, { time: false })).toBe(
      `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`,
    );
    expect(result.current.formatTime(ms, { date: false })).toBe(
      `${pad(d.getHours())}:${pad(d.getMinutes())}`,
    );
    expect(result.current.formatTime(ms, { date: false, time: "second" })).toBe(
      `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`,
    );
    expect(result.current.formatTime(null)).toBe("");
  });

  it("time: 'second' adds the seconds field to the default datetime shape", () => {
    const { result } = renderHook(() => useColumns());
    const ms = Date.UTC(2026, 8, 24, 5, 6, 7);
    const d = new Date(ms);
    expect(result.current.formatTime(ms, { time: "second" })).toBe(
      `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`,
    );
  });

  it("both parts off falls back to the full default shape", () => {
    const { result } = renderHook(() => useColumns());
    const ms = Date.UTC(2026, 8, 24, 5, 6);
    expect(result.current.formatTime(ms, { date: false, time: false })).toBe(
      result.current.formatTime(ms),
    );
  });
});

describe("time locale injection", () => {
  it("formatTime localizes through the injected BCP-47 tag", () => {
    const { result } = renderHook(() => useColumns(), {
      wrapper: ({ children }) => (
        <TimeLocaleContext.Provider value="zh-CN">{children}</TimeLocaleContext.Provider>
      ),
    });
    // Hard-coded expectations: deriving them from the same Intl call would
    // assert nothing. Node ships full ICU, so zh-CN medium/short are stable.
    const ms = new Date(2026, 0, 1, 5, 4, 7).getTime();
    expect(result.current.formatTime(ms, { time: false })).toBe("2026年1月1日");
    expect(result.current.formatTime(ms)).toBe("2026年1月1日 05:04");
    expect(result.current.formatTime(ms, { time: "second" })).toBe("2026年1月1日 05:04:07");
    expect(result.current.formatTime(ms, { date: false, time: "second" })).toBe("05:04:07");
  });

  it("degrades to the neutral skeleton on a malformed tag", () => {
    const { result } = renderHook(() => useColumns(), {
      wrapper: ({ children }) => (
        <TimeLocaleContext.Provider value="not-a-tag!">{children}</TimeLocaleContext.Provider>
      ),
    });
    const ms = new Date(2026, 0, 1, 5, 4).getTime();
    expect(result.current.formatTime(ms, { time: false })).toBe(localStamp(ms).split(" ")[0]);
  });

  it("keeps the locale-neutral skeleton without a provider", () => {
    const { result } = renderHook(() => useColumns());
    expect(result.current.formatTime(0, { time: false })).toBe(localStamp(0).split(" ")[0]);
  });
});
