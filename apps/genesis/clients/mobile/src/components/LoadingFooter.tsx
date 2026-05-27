import { ActivityIndicator, Text, View } from "react-native"
import { i18n } from "@/lib/i18n"

export function LoadingFooter() {
  return (
    <View className="items-center py-4">
      <ActivityIndicator color="#6366f1" size="small" />
      <Text className="mt-2 text-gray-400 text-xs">
        {i18n._("article.loading")}
      </Text>
    </View>
  )
}
