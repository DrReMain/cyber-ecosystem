import type { SessionMessage } from "@cyber-ecosystem/gen-connect-ts/cyber/agent/v1/chat_pb";
import { useEffect } from "react";
import { ErrorFallback } from "#/features/app";
import type { ChatTurn } from "./conversations";
import { ChatTranscript } from "./ui/chat-transcript";
import { EmptyConversation } from "./ui/empty-conversation";
import { useSessionMessages } from "./use-session-messages";

interface TranscriptRegionProps {
  baseUrl: string;
  sessionId: string | undefined;
  turns: ReadonlyArray<ChatTurn>;
  streaming: boolean;
  restoring: boolean;
  error: unknown;
  model: string | undefined;
  onMessagesPage: (sessionId: string, pages: readonly (readonly SessionMessage[])[]) => void;
  onRetry: () => void;
  onSuggest: (text: string) => void;
}

export function TranscriptRegion({
  baseUrl,
  sessionId,
  turns,
  streaming,
  restoring,
  error,
  model,
  onMessagesPage,
  onRetry,
  onSuggest,
}: Readonly<TranscriptRegionProps>) {
  const { hasNextPage, loadEarlier, query } = useSessionMessages(baseUrl, sessionId);

  useEffect(() => {
    const pages = query.data?.pages;
    if (sessionId === undefined || !pages) return;
    onMessagesPage(
      sessionId,
      pages.map((page) => page.list ?? []),
    );
  }, [query.data, sessionId, onMessagesPage]);

  const pending = restoring || (sessionId !== undefined && query.isPending && turns.length === 0);

  return (
    <div className="flex min-h-0 grow flex-col">
      {turns.length === 0 ? (
        <EmptyConversation messagesPending={pending} model={model} onSuggest={onSuggest} />
      ) : (
        <ChatTranscript
          canLoadEarlier={hasNextPage}
          onLoadEarlier={loadEarlier}
          streaming={streaming}
          turns={turns}
        />
      )}
      {error != null && (
        <div className="h-24 shrink-0">
          <ErrorFallback error={error} onRetry={onRetry} />
        </div>
      )}
    </div>
  );
}
