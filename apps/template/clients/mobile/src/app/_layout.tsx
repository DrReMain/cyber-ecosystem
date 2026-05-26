import "@/global.css"

import { I18nProvider } from "@lingui/react"
import { Stack } from "expo-router"
import { useAtom } from "jotai"
import { View } from "react-native"
import { i18n, LOCALES } from "@/lib/i18n"
import { QueryProvider } from "@/lib/query-provider"
import { ConnectProvider } from "@/services/connect-provider"
import { AppProvider } from "@/stores/_core/provider"
import { initI18n, localeAtom } from "@/stores/i18n/store"
import { initTheme } from "@/stores/theme/store"

initTheme()
initI18n()

export default function RootLayout() {
  return (
    <AppProvider>
      <I18nProvider i18n={i18n}>
        <ConnectProvider>
          <QueryProvider>
            <AppShell />
          </QueryProvider>
        </ConnectProvider>
      </I18nProvider>
    </AppProvider>
  )
}

function AppShell() {
  const [locale] = useAtom(localeAtom)
  const direction: "ltr" | "rtl" = LOCALES[locale].rtl ? "rtl" : "ltr"

  return (
    <View style={{ direction, flex: 1 }}>
      <Stack screenOptions={{ headerShown: false }} />
    </View>
  )
}
