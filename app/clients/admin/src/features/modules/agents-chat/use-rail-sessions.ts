import { useInfiniteQuery } from "@connectrpc/connect-query";
import { listSessions } from "@cyber-ecosystem/gen-connect-ts/cyber/agent/v1/chat-ChatService_connectquery";
import { useCallback } from "react";
import { sessionsPageSize } from "./conversations";

// Rail query host — lives inside the rail region so a ListSessions fault
// throws regionally. Flattened pages pump up via onRailSessions; a refresh
// remounts this host through the region's key.
export function useRailSessions(baseUrl: string) {
  const query = useInfiniteQuery(
    listSessions,
    { page: { pageNo: 1, pageSize: sessionsPageSize } },
    {
      enabled: baseUrl !== "",
      getNextPageParam: (lastPage) =>
        lastPage.page?.more ? (lastPage.page.pageNo ?? 1) + 1 : undefined,
      pageParamKey: "page.pageNo",
    },
  );

  const fetchNextPage = query.fetchNextPage;
  const loadMore = useCallback(() => {
    if (!query.hasNextPage || query.isFetchingNextPage) return;
    void fetchNextPage();
  }, [query.hasNextPage, query.isFetchingNextPage, fetchNextPage]);

  return {
    hasNextPage: query.hasNextPage,
    loadMore,
    loadingMore: query.isFetchingNextPage,
    query,
  };
}
