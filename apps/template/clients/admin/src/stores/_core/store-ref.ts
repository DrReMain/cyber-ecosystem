import type { createStore } from "jotai"

type JotaiStore = ReturnType<typeof createStore>

let apiStore: JotaiStore | null = null

export function setApiStore(store: JotaiStore): void {
  if (typeof window !== "undefined") {
    apiStore = store
  }
}

export function getApiStore(): JotaiStore {
  if (!apiStore)
    throw new Error("API store not initialized — call setApiStore() first")
  return apiStore
}
