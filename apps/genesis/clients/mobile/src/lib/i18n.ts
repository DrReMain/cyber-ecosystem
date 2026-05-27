import { i18n } from "@lingui/core"
import { getLocales } from "expo-localization"
import { messages as arSA } from "../../locale/ar-SA/messages"
import { messages as enUS } from "../../locale/en-US/messages"
import { messages as jaJP } from "../../locale/ja-JP/messages"
import { messages as koKR } from "../../locale/ko-KR/messages"
import { messages as zhCN } from "../../locale/zh-CN/messages"

export const LOCALES = {
  "en-US": { label: "English", rtl: false },
  "zh-CN": { label: "简体中文", rtl: false },
  "ar-SA": { label: "العربية", rtl: true },
  "ja-JP": { label: "日本語", rtl: false },
  "ko-KR": { label: "한국어", rtl: false },
} as const

export type Locale = keyof typeof LOCALES
export const SUPPORTED_LOCALES = Object.keys(LOCALES) as Locale[]

i18n.load({
  "en-US": enUS,
  "zh-CN": zhCN,
  "ar-SA": arSA,
  "ja-JP": jaJP,
  "ko-KR": koKR,
})

export function detectLocale(): Locale {
  const deviceLocale = getLocales()[0]
  if (!deviceLocale) return "en-US"

  const tag = `${deviceLocale.languageCode}-${deviceLocale.regionCode ?? ""}`

  for (const locale of SUPPORTED_LOCALES) {
    if (tag === locale) return locale
  }

  for (const locale of SUPPORTED_LOCALES) {
    if (locale.startsWith(deviceLocale.languageCode ?? "")) return locale
  }

  return "en-US"
}

export function isRTL(locale: Locale): boolean {
  return LOCALES[locale].rtl
}

export { i18n }
