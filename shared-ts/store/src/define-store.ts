import { readCookie, writeCookie } from "@cyber-ecosystem/shared-cookie";
import type { Draft } from "immer";
import type { WritableAtom } from "jotai";
import { atom } from "jotai";
import { atomWithImmer } from "jotai-immer";
import type { ZodType } from "zod";

export interface StoreDefinition<T = unknown> {
  key: string;
  initial: T;
  persist: boolean;
  schema?: ZodType<T>;
  immerAtom: WritableAtom<T, [value: T | ((draft: Draft<T>) => void)], void>;
  atom: WritableAtom<T, [update: T | ((draft: Draft<T>) => void)], void>;
}

export interface StoreOptions<T = unknown> {
  persist?: boolean;
  debugLabel?: string;
  schema?: ZodType<T>;
}

const registry: StoreDefinition[] = [];

export function defineStore<T>(
  key: string,
  initial: T,
  options?: StoreOptions<T>,
): StoreDefinition<T> {
  const persist = options?.persist ?? false;
  const debugLabel = options?.debugLabel;
  const schema = options?.schema;

  const immerAtom = atomWithImmer<T>(initial);

  const publicAtom = atom<T, [update: T | ((draft: Draft<T>) => void)], void>(
    (get) => get(immerAtom),
    (get, set, update) => {
      set(immerAtom, update);
      if (persist) {
        writeCookie(key, get(immerAtom));
      }
    },
  );

  if (debugLabel) {
    immerAtom.debugLabel = `__${debugLabel}`;
    publicAtom.debugLabel = debugLabel;
  }

  const def: StoreDefinition<T> = {
    key,
    initial,
    persist,
    schema,
    immerAtom,
    atom: publicAtom,
  };
  registry.push(def as StoreDefinition);

  return def;
}

export function getRegistry(): ReadonlyArray<StoreDefinition> {
  return registry;
}

// RPC-boundary serializability: server fns returning this data must satisfy
// TanStack's serializable validation — `unknown` values don't (the original
// app-side `Serializable` cast was load-bearing). Values come from readCookie
// (superjson parse), serializable by construction.
export type StoreSerializable =
  | null
  | boolean
  | number
  | string
  | StoreSerializable[]
  | { [key: string]: StoreSerializable };

// Pure SSR collection — the consuming app owns the server fn (composition
// root) and only decides how a cookie key reaches this function.
export function collectPersistedStores(
  readRaw: (key: string) => string | undefined,
): Record<string, StoreSerializable> {
  const data: Record<string, unknown> = {};
  for (const store of getRegistry()) {
    if (!store.persist) continue;
    data[store.key] = readCookie(readRaw(store.key), store.schema, store.initial);
  }
  return data as Record<string, StoreSerializable>;
}
