import { useLocalSearchParams, useRouter } from "expo-router"
import { useAtom } from "jotai"
import { Pressable, Text, View } from "react-native"
import { i18n } from "@/lib/i18n"
import { formatDate } from "@/lib/relative-time"
import { localeAtom } from "@/stores/i18n/store"

interface ArticleDetailParams {
  id: string
  title: string
  content: string
  status: string
  createdAt: string
}

const STATUS_PILL: Record<
  string,
  { bg: string; text: string; i18nKey: string }
> = {
  draft: {
    bg: "bg-indigo-500/10",
    text: "text-indigo-500",
    i18nKey: "article.status.draft",
  },
  published: {
    bg: "bg-green-500/10",
    text: "text-green-500",
    i18nKey: "article.status.published",
  },
  archived: {
    bg: "bg-amber-500/10",
    text: "text-amber-500",
    i18nKey: "article.status.archived",
  },
}

export default function ArticleDetail() {
  const router = useRouter()
  const params = useLocalSearchParams() as unknown as ArticleDetailParams
  const pill = STATUS_PILL[params.status ?? ""] ?? STATUS_PILL.draft
  const [locale] = useAtom(localeAtom)
  const createdDate = formatDate(params.createdAt, locale)

  return (
    <View className="flex-1 bg-gray-50 dark:bg-gray-950">
      <View className="flex-row items-center gap-3 px-4 pt-12 pb-3">
        <Pressable onPress={() => router.back()}>
          <Text className="text-base text-indigo-500">←</Text>
        </Pressable>
        <Text className="font-semibold text-base text-gray-900 dark:text-gray-100">
          {i18n._("article.detail")}
        </Text>
      </View>
      <View className="px-4 pt-2">
        <View className={`${pill.bg} mb-3 self-start rounded-full px-3 py-1`}>
          <Text className={`${pill.text} font-semibold text-xs`}>
            {i18n._(pill.i18nKey)}
          </Text>
        </View>
        <Text className="font-bold text-gray-950 text-xl leading-tight dark:text-gray-50">
          {params.title ?? ""}
        </Text>
        {createdDate ? (
          <Text className="mt-2 text-gray-400 text-xs">
            {i18n._("article.createdAt", { date: createdDate })}
          </Text>
        ) : null}
        <View className="mt-4 h-px bg-gray-200 dark:bg-gray-800" />
        <Text className="mt-4 text-gray-700 text-sm leading-7 dark:text-gray-300">
          {params.content ?? ""}
        </Text>
      </View>
    </View>
  )
}
