export function buildSearchPatch<T extends Record<string, unknown>>(
  prev: T,
  patch: Partial<T>,
): T {
  const next: Record<string, unknown> = { ...prev, ...patch }
  for (const [k, v] of Object.entries(patch)) {
    if (v === undefined) delete next[k]
  }
  return next as T
}
