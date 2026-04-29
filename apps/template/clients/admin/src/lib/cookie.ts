import superjson from "superjson"
import type { ZodType } from "zod"

interface CookieOptions {
  maxAge?: number
  path?: string
  sameSite?: "strict" | "lax" | "none"
}

const DEFAULT_MAX_AGE = 31_536_000 // 1 year

export function writeCookie(
  key: string,
  value: unknown,
  options?: CookieOptions,
) {
  if (typeof document === "undefined") return

  const maxAge = options?.maxAge ?? DEFAULT_MAX_AGE
  const path = options?.path ?? "/"
  const sameSite = options?.sameSite ?? "lax"
  const secure = window.location.protocol === "https:" ? ";secure" : ""

  const encoded = encodeURIComponent(superjson.stringify(value))
  // biome-ignore lint/suspicious/noDocumentCookie: cookie must be client-accessible for SSR hydration
  document.cookie = `${key}=${encoded};path=${path};max-age=${maxAge};samesite=${sameSite}${secure}`
}

export function readCookie<T>(
  raw: string | undefined,
  schema: ZodType<T> | undefined,
  fallback: T,
): T {
  if (!raw) return fallback
  try {
    const parsed = superjson.parse(raw)
    if (!schema) return parsed as T
    const result = schema.safeParse(parsed)
    return result.success ? result.data : fallback
  } catch {
    return fallback
  }
}
