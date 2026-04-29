import { useCallback, useMemo } from "react"
import { getPartition } from "./schema"
import type { FilterControls, SearchStore } from "./types"

export function useFilter<T extends Record<string, unknown>>(
  store: SearchStore<T>,
  schema: {
    shape: Record<string, unknown>
    safeParse: (input: unknown) => { success: boolean; data?: T }
  },
): FilterControls<T> {
  const partition = getPartition(schema)

  const defaults = useMemo(() => {
    const result = schema.safeParse({})
    return result.success ? (result.data as Record<string, unknown>) : {}
  }, [schema])

  const values = useMemo(() => {
    const result: Record<string, unknown> = {}
    for (const [key, value] of Object.entries(store.state)) {
      if (!partition.nonFilterKeys.has(key)) {
        result[key] = value
      }
    }
    return result as Partial<T>
  }, [store.state, partition])

  const onFilter = useCallback(
    (formValues: Partial<T>) => {
      const s = store.state as Record<string, unknown>
      const valuesMap = formValues as Record<string, unknown>
      const patch: Record<string, unknown> = {}
      let hasChanges = false

      for (const [key, raw] of Object.entries(valuesMap)) {
        if (partition.nonFilterKeys.has(key)) continue
        const normalized = raw === "" || raw === null ? undefined : raw
        const current = s[key]

        if (normalized !== undefined) {
          if (normalized !== current) hasChanges = true
          patch[key] = normalized
        } else if (current === undefined) {
          if (key in defaults) {
            hasChanges = true
            patch[key] = null
          }
        } else if (current === null) {
          // already cleared
        } else {
          hasChanges = true
          patch[key] = key in defaults ? null : undefined
        }
      }

      if (
        !hasChanges &&
        ((s[partition.pageNoKey ?? "pageNo"] as number) ?? 1) === 1
      )
        return
      if (partition.pageNoKey) patch[partition.pageNoKey] = 1
      store.patch(patch as Partial<T>)
    },
    [store.state, store.patch, defaults, partition],
  )

  const onReset = useCallback(() => {
    const s = store.state as Record<string, unknown>
    const patch: Record<string, unknown> = {}

    for (const [key, value] of Object.entries(defaults)) {
      if (!partition.nonFilterKeys.has(key)) patch[key] = value
    }

    for (const key of Object.keys(s)) {
      if (partition.nonFilterKeys.has(key)) continue
      if (!(key in defaults)) {
        patch[key] = undefined
      }
    }

    let needsNavigate = false
    for (const [key, value] of Object.entries(patch)) {
      if (!Object.is(s[key], value)) {
        needsNavigate = true
        break
      }
    }
    if (!needsNavigate) return

    store.patch(patch as Partial<T>)
  }, [store.state, store.patch, defaults, partition])

  return { values, onFilter, onReset }
}
