import type { Draft } from "immer"
import type { WritableAtom } from "jotai"
import { atom } from "jotai"
import { atomWithImmer } from "jotai-immer"
import type { ZodType } from "zod"
import { storage } from "@/lib/mmkv"

export interface StoreDefinition<T = unknown> {
  key: string
  initial: T
  persist: boolean
  schema?: ZodType<T>
  immerAtom: WritableAtom<T, [value: T | ((draft: Draft<T>) => void)], void>
  atom: WritableAtom<T, [update: T | ((draft: Draft<T>) => void)], void>
}

export interface StoreOptions<T = unknown> {
  persist?: boolean
  schema?: ZodType<T>
}

function readPersisted<T>(key: string, schema?: ZodType<T>): T | undefined {
  const raw = storage.getString(key)
  if (!raw) return undefined
  try {
    const parsed = JSON.parse(raw) as T
    return schema ? schema.parse(parsed) : parsed
  } catch {
    return undefined
  }
}

export function defineStore<T>(
  key: string,
  initial: T,
  options?: StoreOptions<T>,
): StoreDefinition<T> {
  const persist = options?.persist ?? false
  const schema = options?.schema

  const stored = persist ? readPersisted<T>(key, schema) : undefined
  const immerAtom = atomWithImmer<T>(stored ?? initial)

  const publicAtom = atom<T, [update: T | ((draft: Draft<T>) => void)], void>(
    (get) => get(immerAtom),
    (get, set, update) => {
      set(immerAtom, update)
      if (persist) {
        const value = get(immerAtom)
        storage.set(key, JSON.stringify(schema ? schema.parse(value) : value))
      }
    },
  )

  return {
    key,
    initial,
    persist,
    schema,
    immerAtom,
    atom: publicAtom,
  }
}
