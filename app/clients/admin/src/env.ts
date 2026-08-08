import { createEnv } from "@t3-oss/env-core";
import { z } from "zod";

export const env = createEnv({
  runtimeEnv: import.meta.env,
  emptyStringAsUndefined: true,
  server: {
    HTTP_API_URL: z.string().default("http://localhost:11001"),
    CONNECT_API_URL: z.string().default("http://localhost:13001"),
    SESSION_SECRET: z.string().min(32).default("11223344556677889900aabbccddeeff"),
  },
  clientPrefix: "VITE_",
  client: {
    VITE_HTTP_API_URL: z.string().default("http://localhost:5173/http"),
    VITE_CONNECT_API_URL: z.string().default("http://localhost:5173/connect"),
    VITE_SITE_URL: z.url().default("http://localhost:5173"),
  },
});

function isServer(): boolean {
  return typeof window === "undefined";
}

export function resolveCONNECTBaseUrl(): string {
  return isServer() ? env.CONNECT_API_URL : env.VITE_CONNECT_API_URL;
}

export function resolveHTTPBaseUrl(): string {
  return isServer() ? env.HTTP_API_URL : env.VITE_HTTP_API_URL;
}

export function getSiteUrl(): string {
  return env.VITE_SITE_URL.replace(/\/+$/, "");
}
