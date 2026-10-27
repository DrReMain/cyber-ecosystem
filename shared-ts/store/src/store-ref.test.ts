import { createStore } from "jotai";
import { expect, it } from "vitest";
import { getApiStore, setApiStore } from "./store-ref";

// The API-store singleton fails loud before its first set — hydration must
// never silently share an uninitialized store.
it("fails loud before initialization, then serves the set instance", () => {
  expect(() => getApiStore()).toThrow(/not initialized/);
  const instance = createStore();
  setApiStore(instance);
  expect(getApiStore()).toBe(instance);
});
