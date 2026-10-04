import { useMutation } from "@connectrpc/connect-query";
import type { SessionMessage } from "@cyber-ecosystem/gen-connect-ts/cyber/agent/v1/chat_pb";
import {
  deleteSession,
  updateSession,
} from "@cyber-ecosystem/gen-connect-ts/cyber/agent/v1/chat-ChatService_connectquery";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  type ChatTurn,
  type Conversation,
  conversationModel,
  deepLinkStillEligible,
  initialDraftId,
  railConversations,
  type SessionBinding,
  type SessionSummary,
  type WorkspaceState,
} from "./conversations";

// Shared conversation state machine — the single owner of rails, drafts,
// bindings, and hydration. Region query hosts pump their data pages in.
export function useConversations(initialSessionId?: string) {
  const [state, setState] = useState<WorkspaceState>(() => ({
    conversations: [{ id: initialDraftId, turns: [] }],
    activeId: initialDraftId,
  }));
  const nextId = useRef(1);
  const stateRef = useRef(state);
  stateRef.current = state;

  // Draft → server-session bindings. A ref: the first frame binds mid-stream
  // without a re-render; the events that depend on it re-render on their own.
  const bindingsRef = useRef<Record<string, SessionBinding>>({});
  // Messages pages already absorbed per session — the prepend guard.
  const absorbedPagesRef = useRef<Record<string, number>>({});

  // --- rail data: pages pumped up by the rail region's query host ---
  const [sessions, setSessions] = useState<SessionSummary[]>([]);
  const sessionsRef = useRef(sessions);
  sessionsRef.current = sessions;
  // True once rail page 1 has landed — even empty. The pending deep link
  // resolves against this, so a stale id can never spin forever.
  const [railReady, setRailReady] = useState(false);
  const [refreshNonce, setRefreshNonce] = useState(0);

  const onRailSessions = useCallback((summaries: readonly SessionSummary[]) => {
    setSessions((prev) => conversationModel.mergeRailSessions(prev, summaries));
    setRailReady(true);
  }, []);

  const refresh = useCallback(() => {
    // Remounts the rail host: order and scrolled accumulation re-truth from
    // the server.
    setRefreshNonce((n) => n + 1);
  }, []);

  // --- messages data: pages pumped up by the transcript region ---
  const onMessagesPage = useCallback(
    (sessionId: string, pages: readonly (readonly SessionMessage[])[]) => {
      const conversation = stateRef.current.conversations.find((c) => c.id === sessionId);
      if (conversation === undefined) return;
      let absorbed = absorbedPagesRef.current[sessionId] ?? 0;
      // Live turns already carry page 1's content (the in-flight exchange
      // hydrated on the fly): page 1 counts as absorbed without applying.
      if (absorbed === 0 && conversation.turns.length > 0) absorbed = 1;
      if (pages.length <= absorbed) {
        absorbedPagesRef.current[sessionId] = absorbed;
        return;
      }
      // Pages arrive newest-first; earlier pages prepend, oldest first.
      const prefix: ChatTurn[] = [];
      for (let i = pages.length - 1; i >= absorbed; i -= 1) {
        prefix.push(...conversationModel.hydrateTurns(pages[i] ?? [], sessionId, i + 1));
      }
      absorbedPagesRef.current[sessionId] = pages.length;
      setState((prev) => conversationModel.withTurnsPrepended(prev, sessionId, prefix));
    },
    [],
  );

  const activeServerId = (() => {
    const binding = bindingsRef.current[state.activeId];
    if (binding?.persisted) return state.activeId;
    return sessions.some((s) => s.id === state.activeId) ? state.activeId : undefined;
  })();

  const deleteMut = useMutation(deleteSession);
  const renameMut = useMutation(updateSession);

  // --- conversation state operations ---
  const getTurns = useCallback(
    (id: string): Conversation["turns"] =>
      stateRef.current.conversations.find((c) => c.id === id)?.turns ?? [],
    [],
  );
  const updateTurns = useCallback(
    (id: string, updater: (turns: Conversation["turns"]) => Conversation["turns"]) => {
      setState((prev) => ({
        ...prev,
        conversations: prev.conversations.map((c) =>
          c.id === id ? { ...c, turns: updater(c.turns) } : c,
        ),
      }));
    },
    [],
  );
  const appendTurn = useCallback((id: string, turn: Conversation["turns"][number]) => {
    setState((prev) => ({
      ...prev,
      conversations: prev.conversations.map((c) =>
        c.id === id ? { ...c, turns: [...c.turns, turn] } : c,
      ),
    }));
  }, []);
  const sessionIdOf = useCallback((id: string) => {
    const binding = bindingsRef.current[id];
    if (binding) return binding.sessionId;
    return sessionsRef.current.some((s) => s.id === id) ? id : "";
  }, []);
  const bindSession = useCallback((id: string, sessionId: string) => {
    if (bindingsRef.current[id]?.sessionId === sessionId) return;
    bindingsRef.current[id] = { sessionId, persisted: false };
  }, []);
  // Optimistic mirror of the server's send-time updated_at bump; refresh
  // re-truths the order.
  const touch = useCallback((id: string) => {
    setSessions((prev) => {
      const idx = prev.findIndex((s) => s.id === id);
      const row = prev[idx];
      if (idx <= 0 || row === undefined) return prev;
      return [row, ...prev.slice(0, idx), ...prev.slice(idx + 1)];
    });
  }, []);

  const store = useMemo(
    () => ({ getTurns, updateTurns, appendTurn, sessionIdOf, bindSession, touch }),
    [getTurns, updateTurns, appendTurn, sessionIdOf, bindSession, touch],
  );

  const newChat = useCallback(() => {
    nextId.current += 1;
    const id = `c${nextId.current}`;
    setState((prev) => ({
      // Drop other empty drafts: they were never listed and carry nothing.
      conversations: [...prev.conversations.filter((c) => c.turns.length > 0), { id, turns: [] }],
      activeId: id,
    }));
  }, []);

  const select = useCallback((id: string) => {
    setState((prev) => {
      if (prev.conversations.some((c) => c.id === id)) return { ...prev, activeId: id };
      const summary = sessionsRef.current.find((s) => s.id === id);
      return {
        conversations: [...prev.conversations, { id, title: summary?.title, turns: [] }],
        activeId: id,
      };
    });
    if (bindingsRef.current[id] === undefined) {
      bindingsRef.current[id] = { sessionId: id, persisted: true };
    }
  }, []);

  // Deep link: adopt the URL session once the rail's first page lands; a
  // stale id is ignored and the URL mirror clears it. While pending, the
  // content pane holds the restore spinner instead of flashing welcome.
  const [deepLink, setDeepLink] = useState(initialSessionId);
  useEffect(() => {
    if (deepLink === undefined || !railReady) return;
    if (deepLinkStillEligible(stateRef.current) && sessions.some((s) => s.id === deepLink)) {
      select(deepLink);
    }
    setDeepLink(undefined);
  }, [deepLink, railReady, sessions, select]);

  const remove = useCallback(
    (id: string) => {
      const isServer =
        sessionsRef.current.some((s) => s.id === id) || bindingsRef.current[id]?.persisted === true;
      const dropLocal = () => {
        delete bindingsRef.current[id];
        delete absorbedPagesRef.current[id];
        nextId.current += 1;
        setState((prev) => conversationModel.withoutConversation(prev, id, `c${nextId.current}`));
      };
      if (!isServer) {
        dropLocal();
        return;
      }
      // A failed delete keeps the row: the global MutationCache surfaces the
      // error (feedback defaults on) and the session still exists server-side.
      void deleteMut
        .mutateAsync({ sessionId: id })
        .then(() => {
          dropLocal();
          refresh();
        })
        .catch(() => undefined);
    },
    [deleteMut, refresh],
  );

  const rename = useCallback(
    (id: string, title: string) => {
      void renameMut
        .mutateAsync({ sessionId: id, title })
        .then(() => {
          setSessions((prev) => prev.map((s) => (s.id === id ? { ...s, title } : s)));
          setState((prev) => ({
            ...prev,
            conversations: prev.conversations.map((c) => (c.id === id ? { ...c, title } : c)),
          }));
        })
        // A failed rename keeps the old title everywhere; the global
        // MutationCache surfaces the error (feedback defaults on).
        .catch(() => undefined);
    },
    [renameMut],
  );

  const onSettled = useCallback(
    (conversationId: string, outcome: "finished" | "aborted" | "failed", sawDelta: boolean) => {
      const binding = bindingsRef.current[conversationId];
      if (outcome === "failed" || (outcome === "aborted" && !sawDelta)) {
        conversationModel.settleNothingPersisted(
          bindingsRef.current,
          conversationId,
          outcome,
          updateTurns,
        );
        return;
      }
      // finished, or aborted with visible text — the server persisted the
      // exchange. Rename the draft to its session id only now: the stream
      // controllers map is empty, so nothing still keys the draft id.
      if (binding && binding.sessionId !== "" && binding.sessionId !== conversationId) {
        bindingsRef.current[binding.sessionId] = { sessionId: binding.sessionId, persisted: true };
        delete bindingsRef.current[conversationId];
        setState((prev) =>
          conversationModel.withDraftRenamed(prev, conversationId, binding.sessionId),
        );
      } else if (binding) {
        binding.persisted = true;
      }
      refresh();
    },
    [refresh, updateTurns],
  );

  const activeTurns = state.conversations.find((c) => c.id === state.activeId)?.turns ?? [];
  const conversations = railConversations(state.conversations, sessions);

  return {
    activeId: state.activeId,
    activeSessionId: activeServerId,
    activeTurns,
    deepLinkPending: deepLink !== undefined,
    conversations,
    newChat,
    onMessagesPage,
    onRailSessions,
    onSettled,
    refreshNonce,
    remove,
    rename,
    select,
    store,
  };
}
