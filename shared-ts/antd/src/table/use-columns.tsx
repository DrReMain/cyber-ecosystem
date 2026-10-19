import type { Timestamp } from "@bufbuild/protobuf/wkt";
import { timestampMs } from "@bufbuild/protobuf/wkt";
import type { TableColumnType } from "antd";
import { Grid, Typography } from "antd";
import type { CSSProperties } from "react";
import { createContext, useContext, useSyncExternalStore } from "react";

export type DateLike = Timestamp | number | string | null | undefined;

// Display axes, composed rather than enumerated: the calendar part toggles,
// the clock part picks its finest grain. Defaults render the common
// datetime-to-the-minute shape.
export interface TimeFormat {
  date?: boolean;
  time?: false | "minute" | "second";
}

interface ResolvedFormat {
  date: boolean;
  time: false | "minute" | "second";
}

const resolveFormat = (format?: TimeFormat): ResolvedFormat => {
  const date = format?.date ?? true;
  const time = format?.time ?? "minute";
  // Both parts off carries no shape to render; the full default keeps the
  // formatter total instead of returning an empty string.
  return date || time !== false ? { date, time } : { date: true, time: "minute" };
};

// Host-injected BCP-47 tag, formatted through Intl/ECMA-402: engine-level
// CLDR (native calendars and digits, e.g. ar-SA) with no locale data shipped.
// Undefined keeps the locale-neutral skeleton below, so hosts without
// localization see no change.
export const TimeLocaleContext = createContext<string | undefined>(undefined);

const intlOptions = (r: ResolvedFormat): Intl.DateTimeFormatOptions => {
  const options: Intl.DateTimeFormatOptions = {};
  if (r.date) options.dateStyle = "medium";
  if (r.time !== false) options.timeStyle = r.time === "second" ? "medium" : "short";
  return options;
};

// Formatters are pure and reusable — cache per locale×shape so table rows
// don't each rebuild one. Plain module cache: no request-scoped state.
const formatterCache = new Map<string, Intl.DateTimeFormat>();

function intlFormat(ms: number, r: ResolvedFormat, locale: string): string {
  const key = `${locale}|${r.date}|${r.time}`;
  let fmt = formatterCache.get(key);
  if (fmt === undefined) {
    try {
      fmt = new Intl.DateTimeFormat(locale, intlOptions(r));
    } catch {
      // Malformed tag — degrade to the neutral skeleton rather than blowing
      // up the table, and don't cache the failure.
      return formatStamp(ms, r);
    }
    formatterCache.set(key, fmt);
  }
  return fmt.format(ms);
}

const MONO_NOWRAP: CSSProperties = {
  fontFamily: "var(--ant-font-family-code, monospace)",
  whiteSpace: "nowrap",
};

const dataIndexTitle = (dataIndex: unknown): string => {
  if (dataIndex == null) return "";
  return Array.isArray(dataIndex) ? dataIndex.map(String).join(".") : String(dataIndex);
};

export function dateLikeMs(v: DateLike): number | null {
  if (v == null) return null;
  if (typeof v === "number") return v;
  if (typeof v === "string") {
    const parsed = new Date(v).getTime();
    // NaN check, not falsiness: the epoch itself parses to 0.
    return Number.isNaN(parsed) ? null : parsed;
  }
  return timestampMs(v);
}

export function useColumns() {
  const { xs } = Grid.useBreakpoint();
  const client = useSyncExternalStore(
    () => () => {},
    () => true,
    () => false,
  );

  const fixed = (side: "left" | "right"): "left" | "right" | false => (client && xs ? false : side);

  const timeLocale = useContext(TimeLocaleContext);
  const formatTime = (v: DateLike, format?: TimeFormat): string => {
    const ms = dateLikeMs(v);
    if (ms == null) return "";
    const r = resolveFormat(format);
    if (timeLocale === undefined) return formatStamp(ms, r);
    return intlFormat(ms, r, timeLocale);
  };

  const fieldTimestamp = <T extends Record<string, unknown>>(
    dataIndex: NonNullable<TableColumnType<T>["dataIndex"]>,
    options?: { title?: string; format?: TimeFormat; fixed?: "left" | "right" },
  ) =>
    ({
      title: options?.title ?? dataIndexTitle(dataIndex),
      dataIndex,
      onCell: () => ({ style: MONO_NOWRAP }),
      fixed: options?.fixed ? fixed(options.fixed) : undefined,
      render: (v: DateLike) => formatTime(v, options?.format) || "-",
    }) as TableColumnType<T>;

  const fieldAction = <T extends Record<string, unknown>>(
    render: TableColumnType<T>["render"],
    options: { title: string },
  ) =>
    ({
      title: options.title,
      dataIndex: "_ACTION",
      onCell: () => ({ style: { whiteSpace: "nowrap" } }),
      fixed: fixed("right"),
      render,
    }) as TableColumnType<T>;

  const fieldCopy = <T extends Record<string, unknown>>(
    dataIndex: NonNullable<TableColumnType<T>["dataIndex"]>,
    options?: { title?: string; fixed?: "left" | "right" },
  ) =>
    ({
      title: options?.title ?? dataIndexTitle(dataIndex),
      dataIndex,
      fixed: options?.fixed ? fixed(options.fixed) : undefined,
      render: (v: string) => (
        <Typography.Text copyable={{ tooltips: false }} style={MONO_NOWRAP}>
          {v}
        </Typography.Text>
      ),
    }) as TableColumnType<T>;

  return { fixed, formatTime, fieldTimestamp, fieldAction, fieldCopy };
}

function formatStamp(ms: number, r: ResolvedFormat): string {
  const d = new Date(ms);
  const pad = (n: number) => String(n).padStart(2, "0");
  const parts: string[] = [];
  if (r.date) {
    parts.push(`${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`);
  }
  if (r.time !== false) {
    const clock = `${pad(d.getHours())}:${pad(d.getMinutes())}`;
    parts.push(r.time === "second" ? `${clock}:${pad(d.getSeconds())}` : clock);
  }
  return parts.join(" ");
}
