import type { SessionMessage } from "@cyber-ecosystem/gen-connect-ts/cyber/agent/v1/chat_pb";

export type ChatRole = "user" | "assistant";

export interface ChatTurn {
  key: string;
  role: ChatRole;
  content: string;
  reasoning: string;
}

export interface Conversation {
  id: string;
  turns: ChatTurn[];
  title?: string;
}

// The store seam the stream hook drives; the backing impl lives in the
// workspace hook so it can change without touching the hook.
export interface ConversationStore {
  getTurns(id: string): ChatTurn[];
  appendTurn(id: string, turn: ChatTurn): void;
  updateTurns(id: string, updater: (turns: ChatTurn[]) => ChatTurn[]): void;
  // "" = the next exchange creates a session server-side.
  sessionIdOf(id: string): string;
  // Adopts the id echoed in the first frame of a newly created session.
  bindSession(id: string, sessionId: string): void;
  // Send-time activity signal: re-tops the conversation's rail row.
  touch(id: string): void;
}

export interface WorkspaceState {
  conversations: Conversation[];
  activeId: string;
}

export interface SessionSummary {
  id: string;
  title?: string;
}

export interface SessionBinding {
  sessionId: string;
  persisted: boolean;
}

export const sessionsPageSize = 20;
export const messagesPageSize = 50;
// The draft every workspace starts on; a pending URL deep link only applies
// while it is still pristine (no new chat, no first send).
export const initialDraftId = "c0";

// An abort before any visible delta leaves the empty assistant bubble the
// stream hook appended — locally rolled back (nothing was persisted).
function withoutTrailingEmptyAssistant(turns: ChatTurn[]): ChatTurn[] {
  const last = turns[turns.length - 1];
  if (last?.role === "assistant" && last.content === "" && last.reasoning === "") {
    return turns.slice(0, -1);
  }
  return turns;
}

// Settle-time rename: one atomic flip of every id-keyed state; seeding the
// clipped title marks the row server-backed.
function withDraftRenamed(
  state: WorkspaceState,
  draftId: string,
  sessionId: string,
): WorkspaceState {
  return {
    ...state,
    conversations: state.conversations.map((c) =>
      c.id === draftId ? { ...c, id: sessionId, title: c.title ?? conversationTitle(c) } : c,
    ),
    activeId: state.activeId === draftId ? sessionId : state.activeId,
  };
}

function hydrateTurns(
  list: readonly SessionMessage[],
  sessionId: string,
  pageNo: number,
): ChatTurn[] {
  const hydrated: ChatTurn[] = [];
  for (const [i, msg] of list.entries()) {
    hydrated.push({
      key: msg.id ?? `${sessionId}:${pageNo}:${i}`,
      role: msg.role === "assistant" ? "assistant" : "user",
      content: msg.content ?? "",
      reasoning: msg.reasoning ?? "",
    });
  }
  return hydrated.reverse();
}

function withoutConversation(state: WorkspaceState, id: string, successor: string): WorkspaceState {
  const conversations = state.conversations.filter((c) => c.id !== id);
  if (state.activeId !== id) return { ...state, conversations };
  return { conversations: [...conversations, { id: successor, turns: [] }], activeId: successor };
}

// Earlier pages prepend before the turns already held (page 1 = newest).
function withTurnsPrepended(
  state: WorkspaceState,
  sessionId: string,
  prefix: ChatTurn[],
): WorkspaceState {
  return {
    ...state,
    conversations: state.conversations.map((c) =>
      c.id === sessionId ? { ...c, turns: [...prefix, ...c.turns] } : c,
    ),
  };
}

// A pump with more rows than held means accumulation grew: keep the live
// order (an optimistic send-time re-top survives) and append only the new
// ids. Anything else re-truths the server's order wholesale.
export function mergeRailSessions(
  prev: SessionSummary[],
  incoming: readonly SessionSummary[],
): SessionSummary[] {
  if (incoming.length > prev.length) {
    const known = new Set(prev.map((s) => s.id));
    return [...prev, ...incoming.filter((s) => !known.has(s.id))];
  }
  return [...incoming];
}

// Settle bookkeeping when nothing reached the database: a draft unbinds so
// its retry rebuilds with an empty session id; an existing session keeps its
// binding — its earlier turns live on the server and a retry must append.
function settleNothingPersisted(
  bindings: Record<string, SessionBinding>,
  conversationId: string,
  outcome: "finished" | "aborted" | "failed",
  updateTurns: ConversationStore["updateTurns"],
): void {
  const binding = bindings[conversationId];
  if (binding && !binding.persisted) delete bindings[conversationId];
  if (outcome === "aborted") updateTurns(conversationId, withoutTrailingEmptyAssistant);
}

// A deep link is a mount-time intent: it lapses once the user has left the
// pristine initial draft behind (new chat, first send on it).
export function deepLinkStillEligible(state: WorkspaceState): boolean {
  const draft = state.conversations.find((c) => c.id === initialDraftId);
  return state.activeId === initialDraftId && draft !== undefined && draft.turns.length === 0;
}

// Rail list: live drafts newest-first (sending is the newest activity), then
// the server sessions in updated_at DESC order — bumped at send time too.
export function railConversations(
  state: Conversation[],
  sessions: readonly SessionSummary[],
): Conversation[] {
  const byId = new Map(state.map((c) => [c.id, c]));
  const drafts = state
    .filter((c) => c.turns.length > 0 && !sessions.some((s) => s.id === c.id))
    .toReversed();
  const serverConvs = sessions.map(
    (s) => byId.get(s.id) ?? { id: s.id, title: s.title, turns: [] as ChatTurn[] },
  );
  return [...drafts, ...serverConvs];
}

export const conversationModel = {
  hydrateTurns,
  mergeRailSessions,
  settleNothingPersisted,
  withDraftRenamed,
  withTurnsPrepended,
  withoutConversation,
};

export function conversationTitle(conversation: Conversation): string {
  const first = conversation.turns.find((turn) => turn.role === "user")?.content.trim() ?? "";
  const flat = first.replace(/\s+/g, " ");
  return flat.length > 20 ? `${flat.slice(0, 20)}…` : flat;
}
