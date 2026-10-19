import { useInfiniteQuery } from "@connectrpc/connect-query";
import { getSessionMessages } from "@cyber-ecosystem/gen-connect-ts/cyber/agent/v1/chat-ChatService_connectquery";
import { useCallback } from "react";
import { messagesPageSize } from "./conversations";

// Transcript query host — lives inside the content region so a
// GetSessionMessages fault throws regionally. Pages pump up via
// onMessagesPage; undefined session = a draft, the query stays disabled.
export function useSessionMessages(baseUrl: string, sessionId: string | undefined) {
  const query = useInfiniteQuery(
    getSessionMessages,
    { sessionId: sessionId ?? "", page: { pageNo: 1, pageSize: messagesPageSize } },
    {
      enabled: sessionId !== undefined && baseUrl !== "",
      getNextPageParam: (lastPage) =>
        lastPage.page?.more ? (lastPage.page.pageNo ?? 1) + 1 : undefined,
      pageParamKey: "page.pageNo",
    },
  );

  const fetchNextPage = query.fetchNextPage;
  const loadEarlier = useCallback(() => {
    if (!query.hasNextPage || query.isFetchingNextPage) return;
    void fetchNextPage();
  }, [query.hasNextPage, query.isFetchingNextPage, fetchNextPage]);

  return { hasNextPage: query.hasNextPage, loadEarlier, query };
}
