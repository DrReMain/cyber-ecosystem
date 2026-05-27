import { createEnv } from "@t3-oss/env-core"
import { z } from "zod"

function isServer(): boolean {
  return typeof window === "undefined"
}

export const env = createEnv({
  runtimeEnv: import.meta.env,
  emptyStringAsUndefined: true,

  shared: {
    HTTP_API_URL: z.string().default("http://localhost:11001"),
    CONNECT_API_URL: z.string().default("http://localhost:13001"),
  },

  clientPrefix: "VITE_",
  client: {
    VITE_SITE_URL: z.url().default("https://cyber-ecosystem.com"),
    VITE_GLITCHTIP_DSN: z.string().optional(),
    VITE_OTEL_URL: z.string().optional(),
  },
})

export function getSiteUrl(): string {
  return env.VITE_SITE_URL
}

export function resolveApiBaseUrl(
  serverUrl: string,
  clientUrl: string,
): string {
  return isServer() ? serverUrl : clientUrl
}
