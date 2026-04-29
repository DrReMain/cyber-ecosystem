import type { Timestamp } from "@bufbuild/protobuf/wkt"
import { timestampMs } from "@bufbuild/protobuf/wkt"
import type { TableColumnType } from "antd"
import { Grid, Typography } from "antd"
import { useSyncExternalStore } from "react"
import { m } from "#/paraglide/messages"
import { getLocale } from "#/paraglide/runtime"

type DateLike = Timestamp | number | string | null | undefined

function toMs(v: DateLike): number | null {
  if (v == null) return null
  if (typeof v === "number") return v
  if (typeof v === "string") return new Date(v).getTime() || null
  return timestampMs(v)
}

type TimeFormat = "datetime" | "date" | "time"

interface TimeFieldOptions {
  title?: string
  format?: TimeFormat
}

const emptySubscribe = () => () => {}

function pad(n: number): string {
  return String(n).padStart(2, "0")
}

function formatTimeSSR(date: Date, format: TimeFormat): string {
  const fy = date.getFullYear()
  const fM = pad(date.getMonth() + 1)
  const fd = pad(date.getDate())
  const fh = pad(date.getHours())
  const fm = pad(date.getMinutes())
  const fs = pad(date.getSeconds())
  switch (format) {
    case "datetime":
      return `${fy}/${fM}/${fd} ${fh}:${fm}:${fs}`
    case "date":
      return `${fy}/${fM}/${fd}`
    case "time":
      return `${fh}:${fm}:${fs}`
  }
}

function formatTimeLocale(date: Date, format: TimeFormat): string {
  const locale = getLocale()
  switch (format) {
    case "datetime":
      return date.toLocaleString(locale)
    case "date":
      return date.toLocaleDateString(locale)
    case "time":
      return date.toLocaleTimeString(locale)
  }
}

export function useUtils() {
  const { xs } = Grid.useBreakpoint()
  const client = useSyncExternalStore(
    emptySubscribe,
    () => true,
    () => false,
  )

  const fixed = (side: "left" | "right"): "left" | "right" | false =>
    xs ? false : side

  const formatTime = (v: DateLike, format: TimeFormat = "datetime"): string => {
    const ms = toMs(v)
    if (ms == null) return ""
    const date = new Date(ms)
    return client ? formatTimeLocale(date, format) : formatTimeSSR(date, format)
  }

  const fieldTimestamp = <T extends Record<string, unknown>>(
    dataIndex: string,
    options?: TimeFieldOptions & { fixed?: "left" | "right" },
  ) =>
    ({
      title: options?.title ?? dataIndex,
      dataIndex,
      className: "font-mono whitespace-nowrap",
      fixed: options?.fixed ? fixed(options.fixed) : undefined,
      render: (v: DateLike) => formatTime(v, options?.format) || "-",
    }) as TableColumnType<T>

  const fieldAction = <T extends Record<string, unknown>>(
    renderFn: TableColumnType<T>["render"],
    options?: { title?: string },
  ) =>
    ({
      title: options?.title ?? m.column_action(),
      dataIndex: "_ACTION",
      className: "whitespace-nowrap",
      fixed: fixed("right"),
      render: renderFn,
    }) as TableColumnType<T>

  const fieldCopy = <T extends Record<string, unknown>>(
    dataIndex: string,
    options?: { title?: string; fixed?: "left" | "right" },
  ) =>
    ({
      title: options?.title ?? dataIndex,
      dataIndex,
      className: "",
      fixed: options?.fixed ? fixed(options.fixed) : undefined,
      render: (v: string) => (
        <Typography.Text
          className="whitespace-nowrap font-mono"
          copyable={{ tooltips: false }}
        >
          {v}
        </Typography.Text>
      ),
    }) as TableColumnType<T>

  return { fixed, formatTime, fieldTimestamp, fieldAction, fieldCopy }
}
