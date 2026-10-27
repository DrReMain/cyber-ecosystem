import { writeCookie } from "@cyber-ecosystem/shared-cookie";
import { getDefaultStore } from "jotai";
import { describe, expect, it, vi } from "vitest";
import { collectPersistedStores, defineStore } from "./define-store";

// The persist seam's laws: a persist-marked store writes its cookie on
// every set, an immer draft mutates in place, and collection reads only
// the persisted registry through the app-supplied reader.

const rawOf = (key: string) => {
  const pair = document.cookie.split("; ").find((c) => c.startsWith(`${key}=`));
  return pair ? decodeURIComponent(pair.slice(key.length + 1)) : undefined;
};

describe("defineStore", () => {
  it("persists through the cookie writer on every set", () => {
    const def = defineStore("e2e-persist", { n: 0 }, { persist: true });
    const store = getDefaultStore();

    store.set(def.atom, { n: 5 });
    expect(rawOf("e2e-persist")).toBeTruthy();
    store.set(def.atom, (draft) => {
      draft.n = 9;
    });
    expect(store.get(def.atom)).toEqual({ n: 9 });
  });

  it("non-persisted stores never touch the cookie jar", () => {
    const spy = vi.spyOn(document, "cookie", "set");
    const def = defineStore("e2e-ephemeral", { n: 0 });
    getDefaultStore().set(def.atom, { n: 1 });
    expect(spy).not.toHaveBeenCalled();
    spy.mockRestore();
  });
});

describe("collectPersistedStores", () => {
  it("collects only the persisted stores through the reader", () => {
    defineStore("e2e-collected", { n: 1 }, { persist: true });
    defineStore("e2e-uncollected", { n: 2 });
    writeCookie("e2e-collected", { n: 7 });

    const collected = collectPersistedStores((key) =>
      key === "e2e-collected" ? rawOf(key) : undefined,
    );
    expect(collected["e2e-collected"]).toEqual({ n: 7 });
    expect(Object.keys(collected)).not.toContain("e2e-uncollected");
  });

  it("answers a persisted store's initial when its cookie is absent", () => {
    defineStore("e2e-absent", { n: 3 }, { persist: true });
    const collected = collectPersistedStores(() => undefined);
    expect(collected["e2e-absent"]).toEqual({ n: 3 });
  });
});
