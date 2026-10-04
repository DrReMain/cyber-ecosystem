import { createConnectQueryKey, useQuery } from "@connectrpc/connect-query";
import {
  getSessionMessages,
  listModels,
  listSessions,
} from "@cyber-ecosystem/gen-connect-ts/cyber/agent/v1/chat-ChatService_connectquery";
import { useQueryClient } from "@tanstack/react-query";
import { useNavigate, useSearch } from "@tanstack/react-router";
import { useCallback, useEffect, useRef, useState } from "react";
import { BaseErrorBoundary } from "#/features/app";
import { RailRegion } from "./rail-region";
import { TranscriptRegion } from "./transcript-region";
import { ChatView } from "./ui/chat-view";
import { useChatStream } from "./use-chat-stream";
import { useConversations } from "./use-conversations";

const MODEL_STORAGE_KEY = "cyber.agents-chat.model";
// Route id (index routes carry the trailing slash) vs. navigate path.
const CHAT_ROUTE_ID = "/dashboard/";
const CHAT_PATH = "/dashboard";
// Prefix keys (input-less) match every page of the region's infinite query.
const railQueryKey = createConnectQueryKey({ cardinality: undefined, schema: listSessions });
const messagesQueryKey = createConnectQueryKey({
  cardinality: undefined,
  schema: getSessionMessages,
});

export function ChatWorkspace({ baseUrl }: Readonly<{ baseUrl: string }>) {
  const navigate = useNavigate();
  const { session: initialSession } = useSearch({ from: CHAT_ROUTE_ID });
  const initialSessionRef = useRef(initialSession);

  const modelsQuery = useQuery(listModels, {}, { enabled: baseUrl !== "" });
  const models = modelsQuery.data?.models ?? [];

  const [storedModel, setStoredModel] = useState<string>(
    () =>
      (typeof window === "undefined" ? "" : globalThis.localStorage.getItem(MODEL_STORAGE_KEY)) ??
      "",
  );
  const model = storedModel !== "" ? storedModel : models[0];
  useEffect(() => {
    if (storedModel !== "") globalThis.localStorage.setItem(MODEL_STORAGE_KEY, storedModel);
  }, [storedModel]);

  const conversations = useConversations(initialSessionRef.current);
  const chat = useChatStream(model, conversations.store, { onSettled: conversations.onSettled });

  const activeId = conversations.activeId;
  const streaming = chat.streamingIds.includes(activeId);

  useEffect(() => {
    void navigate({
      to: CHAT_PATH,
      search: (prev) => ({ ...prev, session: conversations.activeSessionId }),
      replace: true,
    });
  }, [conversations.activeSessionId, navigate]);

  const [drafts, setDrafts] = useState<Record<string, string>>({});
  const draft = drafts[activeId] ?? "";

  // A region's retry must clear its cached query fault — a bare boundary
  // reset would re-render straight back into the same error.
  const queryClient = useQueryClient();
  const resetRailQuery = useCallback(
    () => void queryClient.resetQueries({ queryKey: railQueryKey }),
    [queryClient],
  );
  const resetMessagesQuery = useCallback(
    () => void queryClient.resetQueries({ queryKey: messagesQueryKey }),
    [queryClient],
  );

  const rail = (slot: { mode: "drawer" | "static"; open: boolean; onClose: () => void }) => (
    <BaseErrorBoundary onReset={resetRailQuery}>
      <RailRegion
        activeId={activeId}
        baseUrl={baseUrl}
        conversations={conversations.conversations}
        key={conversations.refreshNonce}
        mode={slot.mode}
        onClose={slot.onClose}
        onDelete={(id) => {
          chat.stop(id);
          conversations.remove(id);
        }}
        onNew={conversations.newChat}
        onRailSessions={conversations.onRailSessions}
        onRename={conversations.rename}
        onSelect={conversations.select}
        open={slot.open}
        streamingIds={chat.streamingIds}
      />
    </BaseErrorBoundary>
  );

  const content = (
    <BaseErrorBoundary onReset={resetMessagesQuery}>
      <TranscriptRegion
        baseUrl={baseUrl}
        error={chat.errors[activeId] ?? null}
        model={model}
        onMessagesPage={conversations.onMessagesPage}
        onRetry={() => chat.retry(activeId)}
        onSuggest={(text) => chat.send(activeId, text)}
        restoring={conversations.deepLinkPending}
        sessionId={conversations.activeSessionId}
        streaming={streaming}
        turns={conversations.activeTurns}
      />
    </BaseErrorBoundary>
  );

  return (
    <ChatView
      content={content}
      draft={draft}
      model={model}
      models={models}
      modelsLoading={modelsQuery.isPending}
      onDraftChange={(value) => setDrafts((prev) => ({ ...prev, [activeId]: value }))}
      onModelChange={setStoredModel}
      onSend={(text) => {
        chat.send(activeId, text);
        setDrafts((prev) => ({ ...prev, [activeId]: "" }));
      }}
      onStop={() => chat.stop(activeId)}
      rail={rail}
      streaming={streaming}
    />
  );
}
