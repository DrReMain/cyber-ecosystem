import { useAtom } from "jotai"
import { Pressable, Text, View } from "react-native"
import { i18n } from "@/lib/i18n"
import { formatRelativeTime } from "@/lib/relative-time"
import { localeAtom } from "@/stores/i18n/store"

export interface ArticleCardProps {
  id: string
  title: string
  content?: string
  status?: string
  createdAt?: string
  onPress: (params: ArticleCardProps) => void
}

const STATUS_STYLES: Record<
  string,
  {
    bg: string
    text: string
    pillBg: string
    pillText: string
    i18nKey: string
  }
> = {
  draft: {
    bg: "bg-indigo-50 dark:bg-indigo-950",
    text: "text-indigo-900 dark:text-indigo-100",
    pillBg: "bg-indigo-500/10",
    pillText: "text-indigo-500",
    i18nKey: "article.status.draft",
  },
  published: {
    bg: "bg-green-50 dark:bg-green-950",
    text: "text-green-900 dark:text-green-100",
    pillBg: "bg-green-500/10",
    pillText: "text-green-500",
    i18nKey: "article.status.published",
  },
  archived: {
    bg: "bg-amber-50 dark:bg-amber-950",
    text: "text-amber-900 dark:text-amber-100",
    pillBg: "bg-amber-500/10",
    pillText: "text-amber-500",
    i18nKey: "article.status.archived",
  },
}

const DEFAULT_STYLE = STATUS_STYLES.draft

export function ArticleCard({
  id,
  title,
  content,
  status,
  createdAt,
  onPress,
}: ArticleCardProps) {
  const style = (status && STATUS_STYLES[status]) || DEFAULT_STYLE
  const [locale] = useAtom(localeAtom)

  return (
    <Pressable
      className={`${style.bg} rounded-xl p-3 active:opacity-80`}
      onPress={() =>
        onPress({ id, title, content, status, createdAt, onPress })
      }
    >
      <View className="flex-row items-start justify-between gap-2">
        <Text
          className={`${style.text} flex-1 font-semibold text-sm`}
          numberOfLines={1}
        >
          {title}
        </Text>
        <View className={`${style.pillBg} rounded-full px-2 py-0.5`}>
          <Text className={`${style.pillText} font-semibold text-[10px]`}>
            {i18n._(style.i18nKey)}
          </Text>
        </View>
      </View>
      {content ? (
        <Text
          className={`${style.text} mt-1 text-xs opacity-60`}
          numberOfLines={1}
        >
          {content}
        </Text>
      ) : null}
      {createdAt ? (
        <Text className="mt-1.5 text-[11px] text-gray-400">
          {formatRelativeTime(createdAt, locale)}
        </Text>
      ) : null}
    </Pressable>
  )
}
