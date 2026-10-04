import { useEffect } from "react";
import type { Conversation, SessionSummary } from "./conversations";
import { ConversationRail } from "./ui/conversation-rail";
import { useRailSessions } from "./use-rail-sessions";

interface RailRegionProps {
  baseUrl: string;
  conversations: Conversation[];
  activeId: string;
  streamingIds: string[];
  mode: "drawer" | "static";
  open: boolean;
  onRailSessions: (summaries: readonly SessionSummary[]) => void;
  onClose: () => void;
  onSelect: (id: string) => void;
  onNew: () => void;
  onDelete: (id: string) => void;
  onRename: (id: string, title: string) => void;
}

export function RailRegion({
  baseUrl,
  conversations,
  activeId,
  streamingIds,
  mode,
  open,
  onRailSessions,
  onClose,
  onSelect,
  onNew,
  onDelete,
  onRename,
}: Readonly<RailRegionProps>) {
  const { hasNextPage, loadMore, loadingMore, query } = useRailSessions(baseUrl);

  useEffect(() => {
    const pages = query.data?.pages;
    if (!pages) return;
    onRailSessions(
      pages.flatMap((page) =>
        (page.list ?? []).flatMap((s) =>
          s.id === undefined ? [] : [{ id: s.id, title: s.title }],
        ),
      ),
    );
  }, [query.data, onRailSessions]);

  return (
    <ConversationRail
      activeId={activeId}
      conversations={conversations}
      hasMore={hasNextPage}
      loadingMore={loadingMore}
      mode={mode}
      onClose={onClose}
      onDelete={onDelete}
      onLoadMore={loadMore}
      onNew={onNew}
      onRename={onRename}
      onSelect={onSelect}
      open={open}
      streamingIds={streamingIds}
    />
  );
}
