import type { MessageInitShape, MessageShape } from "@bufbuild/protobuf";
import { useTransport } from "@connectrpc/connect-query";
import { ChatService } from "@cyber-ecosystem/gen-connect-ts/cyber/agent/v1/chat_pb";
import { useCallback, useEffect, useRef, useState } from "react";
import { view } from "#/domains/error";
import { runStream } from "#/services/connect/run-stream";
import type { ChatTurn, ConversationStore } from "./conversations";

export type ChatOutcome = "finished" | "aborted" | "failed";

export interface ChatStreamHandlers {
  // Fires once per exchange after the controller is released — the only
  // window where renaming a draft to its server session id is safe.
  onSettled?(conversationId: string, outcome: ChatOutcome, sawDelta: boolean): void;
}

export function useChatStream(
  model: string | undefined,
  store: ConversationStore,
  handlers?: Readonly<ChatStreamHandlers>,
) {
  const transport = useTransport();
  const onSettled = handlers?.onSettled;
  const [streamingIds, setStreamingIds] = useState<string[]>([]);
  const [errors, setErrors] = useState<Record<string, unknown>>({});
  const controllers = useRef(new Map<string, AbortController>());
  const nextKey = useRef(0);

  useEffect(
    () => () => {
      for (const controller of controllers.current.values()) controller.abort();
    },
    [],
  );

  const markStreaming = useCallback((id: string, streaming: boolean) => {
    setStreamingIds((prev) =>
      streaming ? (prev.includes(id) ? prev : [...prev, id]) : prev.filter((x) => x !== id),
    );
  }, []);

  const runExchange = useCallback(
    (conversationId: string, content: string) => {
      nextKey.current += 1;
      const assistantKey = `assistant-${nextKey.current}`;
      store.appendTurn(conversationId, {
        key: assistantKey,
        role: "assistant",
        content: "",
        reasoning: "",
      });
      markStreaming(conversationId, true);
      setErrors((prev) => ({ ...prev, [conversationId]: null }));

      const controller = new AbortController();
      controllers.current.set(conversationId, controller);
      let sawDelta = false;
      // The generator body runs when the transport consumes it — bind time,
      // not call time: a brand-new draft has no binding yet and reads "".
      const input = (async function* () {
        yield {
          model,
          sessionId: store.sessionIdOf(conversationId),
          content,
        } satisfies MessageInitShape<typeof ChatService.method.chat.input>;
      })();

      let caught: ChatOutcome | undefined;
      runStream(transport, ChatService.method.chat, controller.signal, input, (frame) => {
        // Unset scalars normalize to "" on message instances — truthiness.
        if (frame.sessionId) store.bindSession(conversationId, frame.sessionId);
        appendDelta(frame);
      })
        .catch((err: unknown) => {
          if (view(err).silent) {
            caught = "aborted";
            return;
          }
          caught = "failed";
          setErrors((prev) => ({ ...prev, [conversationId]: err }));
          // A failure before the first delta leaves an empty bubble behind.
          store.updateTurns(conversationId, (turns) => {
            const last = turns[turns.length - 1];
            if (last?.key === assistantKey && last.content === "" && last.reasoning === "") {
              return turns.slice(0, -1);
            }
            return turns;
          });
        })
        .finally(() => {
          controllers.current.delete(conversationId);
          markStreaming(conversationId, false);
          // A clean for-await exit under an aborted signal is still an abort
          // (runStream returns on abort, it does not throw). Settling only
          // happens here: the controllers map no longer holds this id, so
          // the settle handler may safely re-key the conversation.
          const outcome = caught ?? (controller.signal.aborted ? "aborted" : "finished");
          onSettled?.(conversationId, outcome, sawDelta);
        });

      function appendDelta(frame: MessageShape<(typeof ChatService.method.chat)["output"]>) {
        // Frames arrive protojson-shaped: an unset delta is omitted, not "".
        const contentDelta = frame.contentDelta ?? "";
        const reasoningDelta = frame.reasoningDelta ?? "";
        if (contentDelta === "" && reasoningDelta === "") return;
        sawDelta = true;
        store.updateTurns(conversationId, (turns) => {
          const last = turns[turns.length - 1];
          if (last?.key !== assistantKey) return turns;
          const next = turns.slice();
          next[next.length - 1] = {
            ...last,
            content: last.content + contentDelta,
            reasoning: last.reasoning + reasoningDelta,
          };
          return next;
        });
      }
    },
    [transport, model, store, markStreaming, onSettled],
  );

  const send = useCallback(
    (conversationId: string, text: string) => {
      const content = text.trim();
      if (content === "" || model === undefined) return;
      // The controller map is a ref, so this guard cannot go stale mid-render.
      if (controllers.current.has(conversationId)) return;
      nextKey.current += 1;
      const userTurn: ChatTurn = {
        key: `user-${nextKey.current}`,
        role: "user",
        content,
        reasoning: "",
      };
      store.appendTurn(conversationId, userTurn);
      store.touch(conversationId);
      // History is server-owned: only the new content travels.
      runExchange(conversationId, content);
    },
    [model, store, runExchange],
  );

  const retry = useCallback(
    (conversationId: string) => {
      setErrors((prev) => ({ ...prev, [conversationId]: null }));
      const turns = store.getTurns(conversationId);
      for (let i = turns.length - 1; i >= 0; i -= 1) {
        const turn = turns[i];
        if (turn?.role === "user") {
          store.updateTurns(conversationId, (prev) => prev.slice(0, i + 1));
          store.touch(conversationId);
          // Same truncation, content-only resend — the binding (or its
          // absence after an unbind) decides append vs rebuild server-side.
          runExchange(conversationId, turn.content);
          return;
        }
      }
    },
    [store, runExchange],
  );

  const stop = useCallback((conversationId: string) => {
    controllers.current.get(conversationId)?.abort();
  }, []);

  return { streamingIds, errors, send, retry, stop };
}
