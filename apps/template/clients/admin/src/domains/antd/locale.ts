import dayjs from "dayjs"
import "dayjs/locale/ar-sa"
import "dayjs/locale/ja"
import "dayjs/locale/ko"
import "dayjs/locale/zh-cn"
import arEG from "antd/locale/ar_EG"
import enUS from "antd/locale/en_US"
import jaJP from "antd/locale/ja_JP"
import koKR from "antd/locale/ko_KR"
import zhCN from "antd/locale/zh_CN"
import { useEffect, useState } from "react"
import { getLocale, getTextDirection } from "#/paraglide/runtime"

interface LocaleEntry {
  antd: typeof enUS
  dayjs: string
}

// ---------------------------------------------------------------------------------------------------------------------

const defaultEntry: LocaleEntry = { antd: zhCN, dayjs: "zh-cn" }
const registry: Record<string, LocaleEntry> = {
  "en-US": { antd: enUS, dayjs: "en" },
  "zh-CN": { antd: zhCN, dayjs: "zh-cn" },
  "ar-SA": { antd: arEG, dayjs: "ar-sa" },
  "ja-JP": { antd: jaJP, dayjs: "ja" },
  "ko-KR": { antd: koKR, dayjs: "ko" },
}

// ---------------------------------------------------------------------------------------------------------------------

export function useAntdLocale() {
  const locale = getLocale()
  const entry = registry[locale] ?? defaultEntry
  const direction = getTextDirection(locale)

  const [dayjsReady, setDayjsReady] = useState(false)
  useEffect(() => {
    dayjs.locale(entry.dayjs)
    setDayjsReady(true)
  }, [entry.dayjs])

  return { locale: entry.antd, direction, dayjsReady }
}
