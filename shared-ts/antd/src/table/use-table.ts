import { useSyncExternalStore } from "react";
import { storageGet, storageSet } from "../storage";

export type TableSize = "small" | "middle" | "large";

// Density is one global preference, not per-table state: every toolbar in
// the app toggles the same value (vben semantics), persisted per browser so
// a reload keeps the choice. Module scope keeps the setter referentially
// stable across consumers.
const STORAGE_KEY = "cyber.table-size";
const DEFAULT_SIZE: TableSize = "middle";

const listeners = new Set<() => void>();
let current: TableSize | null = null;

function snapshot(): TableSize {
  if (current === null) {
    current = DEFAULT_SIZE;
    const saved = storageGet(STORAGE_KEY);
    if (saved === "small" || saved === "large") current = saved;
  }
  return current;
}

function setTableSize(size: TableSize): void {
  current = size;
  storageSet(STORAGE_KEY, size);
  for (const notify of listeners) notify();
}

function subscribe(listener: () => void): () => void {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

export function useTable() {
  // Server snapshot stays the default: a persisted client choice surfaces
  // right after hydration as a plain re-render, never a mismatch.
  const tableSize = useSyncExternalStore(subscribe, snapshot, () => DEFAULT_SIZE);
  return { tableSize, setTableSize };
}
