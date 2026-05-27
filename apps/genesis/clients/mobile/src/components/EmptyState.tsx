import { Text, View } from "react-native"
import { i18n } from "@/lib/i18n"

export function EmptyState() {
  return (
    <View className="flex-1 items-center justify-center px-4 py-12">
      <Text className="mb-3 text-4xl">📝</Text>
      <Text className="text-base text-gray-400">{i18n._("article.empty")}</Text>
    </View>
  )
}
