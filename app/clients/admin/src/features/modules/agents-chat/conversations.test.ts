import fc from "fast-check";
import { describe, expect, it, vi } from "vitest";

import {
  type ChatTurn,
  type Conversation,
  conversationModel,
  conversationTitle,
  deepLinkStillEligible,
  initialDraftId,
  railConversations,
  type SessionBinding,
  type WorkspaceState,
} from "./conversations";

const {
  hydrateTurns,
  mergeRailSessions,
  withDraftRenamed,
  withTurnsPrepended,
  withoutConversation,
} = conversationModel;

const state = (conversations: Conversation[], activeId: string): WorkspaceState => ({
  conversations,
  activeId,
});

describe("conversationTitle", () => {
  it("flattens whitespace and clips at 20 with an ellipsis", () => {
    fc.assert(
      fc.property(fc.string({ minLength: 1 }), (raw) => {
        const flat = raw.replace(/\s+/g, " ").trim();
        const got = conversationTitle({
          id: "c",
          turns: [{ key: "k", role: "user", content: raw, reasoning: "" }],
        });
        expect(got.length).toBeLessThanOrEqual(21);
        expect(got.startsWith(flat.slice(0, 20)) || flat.length > 20).toBe(true);
      }),
    );
  });

  it("falls back to the empty string when no user turn exists", () => {
    expect(
      conversationTitle({
        id: "c",
        turns: [{ key: "k", role: "assistant", content: "hi", reasoning: "" }],
      }),
    ).toBe("");
  });
});

describe("hydrateTurns", () => {
  it("reverses server order (page 1 = newest) and fills the wire omissions", () => {
    // Wire rows arrive protojson-shaped (unset scalars absent), not as
    // message instances — the transport family's standing reality.
    const rows = [
      { id: "m1", role: "assistant", content: "a", reasoning: undefined },
      { id: undefined, role: "operator", content: undefined, reasoning: "r" },
    ] as unknown as Parameters<typeof hydrateTurns>[0];
    const got = hydrateTurns(rows, "s1", 2);
    expect(got).toEqual([
      { key: "s1:2:1", role: "user", content: "", reasoning: "r" },
      { key: "m1", role: "assistant", content: "a", reasoning: "" },
    ]);
  });
});

describe("mergeRailSessions", () => {
  it("grows by appending only new ids in incoming order; anything else re-truths", () => {
    fc.assert(
      fc.property(
        fc.array(fc.string({ minLength: 1 }), { maxLength: 5, unique: true } as never),
        fc.array(fc.string({ minLength: 1 }), { maxLength: 5, unique: true } as never),
        (prevIds, nextIds) => {
          const prev = prevIds.map((id) => ({ id }));
          const incoming = nextIds.map((id) => ({ id }));
          const out = mergeRailSessions(prev, incoming);
          if (incoming.length > prev.length) {
            expect(out.map((s) => s.id)).toEqual([
              ...prevIds,
              ...nextIds.filter((id) => !prevIds.includes(id)),
            ]);
          } else {
            expect(out.map((s) => s.id)).toEqual(nextIds);
          }
        },
      ),
    );
  });
});

describe("workspace transitions", () => {
  it("withDraftRenamed flips every id-keyed slot atomically and seeds the title", () => {
    const s = state(
      [
        {
          id: initialDraftId,
          turns: [{ key: "k", role: "user", content: "hello world question", reasoning: "" }],
        },
        { id: "other", turns: [] },
      ],
      initialDraftId,
    );
    const out = withDraftRenamed(s, initialDraftId, "s9");
    expect(out.activeId).toBe("s9");
    expect(out.conversations[0]?.id).toBe("s9");
    expect(out.conversations[0]?.title).toBe("hello world question");
    expect(out.conversations[1]?.id).toBe("other");
  });

  it("withoutConversation re-homes the active id to the successor", () => {
    const out = withoutConversation(
      state(
        [
          { id: "a", turns: [] },
          { id: "b", turns: [] },
        ],
        "b",
      ),
      "b",
      initialDraftId,
    );
    expect(out.activeId).toBe(initialDraftId);
    expect(out.conversations.map((c) => c.id)).toEqual(["a", initialDraftId]);
  });

  it("withTurnsPrepended keeps earlier pages before the held turns, only for the target", () => {
    const prefix: ChatTurn[] = [{ key: "p", role: "user", content: "old", reasoning: "" }];
    const out = withTurnsPrepended(
      state(
        [
          { id: "s1", turns: [{ key: "n", role: "user", content: "new", reasoning: "" }] },
          { id: "s2", turns: [] },
        ],
        "s1",
      ),
      "s1",
      prefix,
    );
    expect(out.conversations[0]?.turns.map((t) => t.key)).toEqual(["p", "n"]);
    expect(out.conversations[1]?.turns).toEqual([]);
  });
});

describe("deepLinkStillEligible", () => {
  it("holds only on the pristine initial draft", () => {
    expect(deepLinkStillEligible(state([{ id: initialDraftId, turns: [] }], initialDraftId))).toBe(
      true,
    );
    expect(
      deepLinkStillEligible(
        state(
          [
            {
              id: initialDraftId,
              turns: [{ key: "k", role: "user", content: "x", reasoning: "" }],
            },
          ],
          initialDraftId,
        ),
      ),
    ).toBe(false);
    expect(deepLinkStillEligible(state([{ id: initialDraftId, turns: [] }], "other"))).toBe(false);
  });
});

describe("railConversations", () => {
  it("places live drafts newest-first ahead of the server-ordered sessions", () => {
    const convs: Conversation[] = [
      { id: "s1", turns: [{ key: "1", role: "user", content: "a", reasoning: "" }] },
      { id: "d2", turns: [{ key: "2", role: "user", content: "b", reasoning: "" }] },
      { id: "d1", turns: [{ key: "3", role: "user", content: "c", reasoning: "" }] },
    ];
    const out = railConversations(convs, [{ id: "s2" }, { id: "s1" }]);
    expect(out.map((c) => c.id)).toEqual(["d1", "d2", "s2", "s1"]);
    expect(out[2]?.turns).toEqual([]);
  });
});

describe("settleNothingPersisted", () => {
  it("unbinds an unpersisted draft but keeps a persisted session's binding; aborted rolls the empty bubble back", () => {
    const bindings: Record<string, SessionBinding> = {
      draft: { sessionId: "", persisted: false },
      live: { sessionId: "s1", persisted: true },
    };
    const updateTurns = vi.fn();
    conversationModel.settleNothingPersisted(bindings, "draft", "failed", updateTurns);
    expect(bindings.draft).toBeUndefined();
    conversationModel.settleNothingPersisted(bindings, "live", "failed", updateTurns);
    expect(bindings.live).toEqual({ sessionId: "s1", persisted: true });
    conversationModel.settleNothingPersisted(bindings, "live", "aborted", updateTurns);
    expect(updateTurns).toHaveBeenCalledTimes(1);
  });
});
