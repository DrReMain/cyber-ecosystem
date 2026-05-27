import { i18n } from "@/lib/i18n"

const hasRTF = typeof Intl !== "undefined" && "RelativeTimeFormat" in Intl

function fallbackRelativeTime(seconds: number): string {
  if (seconds < 60) return i18n._("article.time.justNow")
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return i18n._("article.time.minutesAgo", { n: minutes })
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return i18n._("article.time.hoursAgo", { n: hours })
  const days = Math.floor(hours / 24)
  return i18n._("article.time.daysAgo", { n: days })
}

export function formatRelativeTime(
  input: string | { seconds: bigint; nanos: number } | undefined,
  locale?: string,
): string {
  if (!input) return ""
  let date: Date
  if (typeof input === "string") {
    date = new Date(input)
  } else {
    date = new Date(Number(input.seconds) * 1000 + input.nanos / 1e6)
  }
  if (Number.isNaN(date.getTime())) return ""

  const seconds = Math.floor((Date.now() - date.getTime()) / 1000)

  if (hasRTF) {
    const rtf = new Intl.RelativeTimeFormat(locale, { numeric: "auto" })
    if (seconds < 60) return rtf.format(-seconds, "second")
    const minutes = Math.floor(seconds / 60)
    if (minutes < 60) return rtf.format(-minutes, "minute")
    const hours = Math.floor(minutes / 60)
    if (hours < 24) return rtf.format(-hours, "hour")
    const days = Math.floor(hours / 24)
    return rtf.format(-days, "day")
  }

  return fallbackRelativeTime(seconds)
}

export function formatDate(
  input: string | { seconds: bigint; nanos: number } | undefined,
  locale?: string,
): string {
  if (!input) return ""
  let date: Date
  if (typeof input === "string") {
    date = new Date(input)
  } else {
    date = new Date(Number(input.seconds) * 1000 + input.nanos / 1e6)
  }
  if (Number.isNaN(date.getTime())) return ""
  return date.toLocaleDateString(locale, {
    year: "numeric",
    month: "long",
    day: "numeric",
  })
}
