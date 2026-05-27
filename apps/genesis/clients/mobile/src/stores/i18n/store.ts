import { atom } from "jotai"
import { detectLocale, i18n, type Locale, SUPPORTED_LOCALES } from "@/lib/i18n"
import { storage } from "@/lib/mmkv"

function getStoredLocale(): Locale {
  const stored = storage.getString("locale")
  if (stored && (SUPPORTED_LOCALES as readonly string[]).includes(stored))
    return stored as Locale
  return detectLocale()
}

export function initI18n() {
  i18n.activate(getStoredLocale())
}

export const localeAtom = atom<Locale>(getStoredLocale())

export const setLocaleAtom = atom(null, (_get, set, locale: Locale) => {
  if (!(SUPPORTED_LOCALES as readonly string[]).includes(locale)) return
  storage.set("locale", locale)
  i18n.activate(locale)
  set(localeAtom, locale)
})
