import { useRouter } from "expo-router"
import { useAtom } from "jotai"
import { useState } from "react"
import { Pressable, Text, View } from "react-native"
import { ArticleList } from "@/components/ArticleList"
import { LocalePicker } from "@/components/LocalePicker"
import { i18n, LOCALES } from "@/lib/i18n"
import { localeAtom } from "@/stores/i18n/store"

export default function Index() {
  const router = useRouter()
  const [locale] = useAtom(localeAtom)
  const [localePickerVisible, setLocalePickerVisible] = useState(false)

  return (
    <View className="flex-1 bg-gray-50 dark:bg-gray-950">
      <View className="flex-row items-center justify-between px-4 pt-12 pb-3">
        <Text className="font-bold text-2xl text-gray-950 dark:text-gray-50">
          {i18n._("article.title")}
        </Text>
        <View className="flex-row items-center gap-2">
          <Pressable
            className="rounded-full bg-gray-200 px-2.5 py-1 active:opacity-80 dark:bg-gray-800"
            onPress={() => setLocalePickerVisible(true)}
          >
            <Text className="font-medium text-gray-600 text-xs dark:text-gray-300">
              {LOCALES[locale].label}
            </Text>
          </Pressable>
          <Pressable
            className="h-9 w-9 items-center justify-center rounded-full bg-indigo-500 active:opacity-80"
            onPress={() => router.push("/article/create")}
          >
            <Text className="font-light text-white text-xl">+</Text>
          </Pressable>
        </View>
      </View>
      <ArticleList />
      <LocalePicker
        onClose={() => setLocalePickerVisible(false)}
        visible={localePickerVisible}
      />
    </View>
  )
}
