import { createEnv } from "@t3-oss/env-core";
import { createIsomorphicFn } from "@tanstack/react-start";
import { z } from "zod";

export const env = createEnv({
  emptyStringAsUndefined: true,
  runtimeEnv: {
    ...import.meta.env,
  },
  clientPrefix: "VITE_",
  client: {
    VITE_CONNECT_API_PROXY: z.string().default("/connect"),
    VITE_GLITCHTIP_DSN: z.url().optional(),
  },
});

export const getSiteUrl = createIsomorphicFn()
  .server(async () => {
    const { getRequestUrl } = await import("@tanstack/react-start/server");
    return getRequestUrl({ xForwardedHost: true }).origin;
  })
  .client(() => window.location.origin);
