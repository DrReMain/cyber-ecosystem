import { createEnv } from "@t3-oss/env-core";
import { createIsomorphicFn } from "@tanstack/react-start";
import { z } from "zod";

export const env = createEnv({
  emptyStringAsUndefined: true,
  runtimeEnv: {
    ...import.meta.env,
    CONNECT_API_URL: process.env.CONNECT_API_URL,
  },
  server: {
    CONNECT_API_URL: z.url(),
  },
  clientPrefix: "VITE_",
  client: {
    VITE_CONNECT_API_PROXY: z.string().default("/connect"),
    VITE_GLITCHTIP_DSN: z.url().optional(),
  },
});

export const resolveCONNECTBaseUrl = createIsomorphicFn()
  .server(() => env.CONNECT_API_URL)
  .client(() => env.VITE_CONNECT_API_PROXY);

export const getSiteUrl = createIsomorphicFn()
  .server(async () => {
    const { getRequestUrl } = await import("@tanstack/react-start/server");
    return getRequestUrl({ xForwardedHost: true }).origin;
  })
  .client(() => window.location.origin);
