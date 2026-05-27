import { ConnectError, createClient } from "@connectrpc/connect"
import { useTransport } from "@connectrpc/connect-query"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { useRouter } from "expo-router"
import { useState } from "react"
import { Alert, Pressable, Text, TextInput, View } from "react-native"
import { i18n } from "@/lib/i18n"
import { MobileArticleService } from "@/services/connect/genesis/api/v1/bff_mobile/mobile_article_pb"

export default function ArticleCreate() {
  const router = useRouter()
  const queryClient = useQueryClient()
  const transport = useTransport()
  const client = createClient(MobileArticleService, transport)
  const [title, setTitle] = useState("")
  const [content, setContent] = useState("")

  const mutation = useMutation({
    mutationFn: () =>
      client.createArticle({ title, content: content || undefined }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["articles"] })
      router.back()
    },
    onError: (error: unknown) => {
      const msg = error instanceof ConnectError ? error.message : String(error)
      Alert.alert(i18n._("article.error.title"), msg)
    },
  })

  const canPublish = title.trim().length > 0 && !mutation.isPending

  return (
    <View className="flex-1 bg-gray-50 dark:bg-gray-950">
      <View className="flex-row items-center justify-between px-4 pt-12 pb-3">
        <Pressable disabled={mutation.isPending} onPress={() => router.back()}>
          <Text className="font-medium text-indigo-500 text-sm">
            {i18n._("article.cancel")}
          </Text>
        </Pressable>
        <Text className="font-semibold text-base text-gray-900 dark:text-gray-100">
          {i18n._("article.create")}
        </Text>
        <Pressable
          className={`rounded-lg px-4 py-1.5 ${
            canPublish
              ? "bg-indigo-500 active:opacity-80"
              : "bg-gray-300 dark:bg-gray-700"
          }`}
          disabled={!canPublish}
          onPress={() => mutation.mutate()}
        >
          <Text
            className={`font-semibold text-sm ${
              canPublish ? "text-white" : "text-gray-500 dark:text-gray-400"
            }`}
          >
            {i18n._("article.publish")}
          </Text>
        </Pressable>
      </View>
      <View className="flex-1 px-4">
        <TextInput
          className="border-gray-200 border-b pb-2 font-semibold text-gray-950 text-xl dark:border-gray-800 dark:text-gray-50"
          editable={!mutation.isPending}
          onChangeText={setTitle}
          placeholder={i18n._("article.titlePlaceholder")}
          placeholderTextColor="#94a3b8"
          value={title}
        />
        <TextInput
          className="mt-3 min-h-40 text-base text-gray-700 dark:text-gray-300"
          editable={!mutation.isPending}
          multiline
          onChangeText={setContent}
          placeholder={i18n._("article.contentPlaceholder")}
          placeholderTextColor="#94a3b8"
          textAlignVertical="top"
          value={content}
        />
      </View>
    </View>
  )
}
