import { createClient } from "@connectrpc/connect"
import { useTransport } from "@connectrpc/connect-query"
import { useInfiniteQuery } from "@tanstack/react-query"
import { useRouter } from "expo-router"
import { useCallback, useRef } from "react"
import { FlatList, Pressable, RefreshControl, Text, View } from "react-native"
import type { ArticleCardProps } from "@/components/ArticleCard"
import { ArticleCard } from "@/components/ArticleCard"
import { EmptyState } from "@/components/EmptyState"
import { LoadingFooter } from "@/components/LoadingFooter"
import { i18n } from "@/lib/i18n"
import { MobileArticleService } from "@/services/connect/genesis/api/v1/bff_mobile/mobile_article_pb"

const PAGE_SIZE = 10

export function ArticleList() {
  const router = useRouter()
  const transport = useTransport()
  const client = createClient(MobileArticleService, transport)

  const {
    data,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    isLoading,
    isError,
    isRefetching,
    refetch,
  } = useInfiniteQuery({
    queryKey: ["articles"],
    queryFn: ({ pageParam }) =>
      client.queryArticle({ page: { pageNo: pageParam, pageSize: PAGE_SIZE } }),
    initialPageParam: 1,
    getNextPageParam: (lastPage) => {
      if (!lastPage.page?.more) return undefined
      return lastPage.page.pageNo + 1
    },
  })

  const articles = data?.pages.flatMap((page) => page.list) ?? []

  const handlePress = useCallback(
    (props: ArticleCardProps) => {
      router.push({
        pathname: "/article/[id]",
        params: {
          id: props.id,
          title: props.title ?? "",
          content: props.content ?? "",
          status: props.status ?? "",
          createdAt: props.createdAt ?? "",
        },
      })
    },
    [router],
  )

  const fetchingRef = useRef(false)

  const handleEndReached = useCallback(() => {
    if (hasNextPage && !fetchingRef.current) {
      fetchingRef.current = true
      fetchNextPage().finally(() => {
        fetchingRef.current = false
      })
    }
  }, [hasNextPage, fetchNextPage])

  if (isLoading) {
    return <LoadingFooter />
  }

  if (isError) {
    return (
      <View className="flex-1 items-center justify-center px-4 py-12">
        <Text className="mb-3 text-base text-gray-400">
          {i18n._("article.error.load")}
        </Text>
        <Pressable
          className="rounded-lg bg-indigo-500 px-4 py-2 active:opacity-80"
          onPress={() => refetch()}
        >
          <Text className="font-semibold text-sm text-white">
            {i18n._("article.error.retry")}
          </Text>
        </Pressable>
      </View>
    )
  }

  return (
    <FlatList
      contentContainerClassName="p-4 gap-2"
      data={articles}
      ItemSeparatorComponent={() => <View className="h-2" />}
      keyExtractor={(item) => item.id ?? ""}
      ListEmptyComponent={<EmptyState />}
      ListFooterComponent={isFetchingNextPage ? <LoadingFooter /> : null}
      onEndReached={handleEndReached}
      onEndReachedThreshold={0.5}
      refreshControl={
        <RefreshControl
          colors={["#6366f1"]}
          onRefresh={() => refetch()}
          refreshing={isRefetching}
        />
      }
      renderItem={({ item }) => (
        <ArticleCard
          content={item.content}
          createdAt={
            item.createdAt
              ? new Date(Number(item.createdAt.seconds) * 1000).toISOString()
              : undefined
          }
          id={item.id ?? ""}
          onPress={handlePress}
          status={item.status}
          title={item.title ?? ""}
        />
      )}
    />
  )
}
