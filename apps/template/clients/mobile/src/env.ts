import { createEnv } from "@t3-oss/env-core"
import { z } from "zod"

export const env = createEnv({
  runtimeEnv: process.env,
  emptyStringAsUndefined: true,

  clientPrefix: "EXPO_PUBLIC_",
  client: {
    EXPO_PUBLIC_CONNECT_API_URL: z.string().default("http://localhost:13000"),
    EXPO_PUBLIC_HTTP_API_URL: z.string().default("http://localhost:11000"),
    EXPO_PUBLIC_GLITCHTIP_DSN: z.string().optional(),
    EXPO_PUBLIC_OTEL_URL: z.string().optional(),
  },
})
