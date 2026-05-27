import { createStore } from "jotai"
import { getRegistry } from "./define-store"

export function createMainStore(initialData?: Record<string, unknown>) {
  const store = createStore()
  if (initialData) {
    for (const def of getRegistry()) {
      if (def.persist && initialData[def.key] !== undefined) {
        store.set(def.immerAtom, initialData[def.key])
      }
    }
  }
  return store
}
