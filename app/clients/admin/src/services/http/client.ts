import { createClient } from "@cyber-ecosystem/gen-openapi-ts/client/client.gen";
import { resolveHTTPBaseUrl } from "#/env";

export const httpClient = createClient({
  baseUrl: resolveHTTPBaseUrl(),
  throwOnError: true,
});

httpClient.interceptors.error.use((error, response) => {
  const base = typeof error === "object" && error !== null ? error : { message: String(error) };
  return response === undefined ? base : { ...base, status: response.status };
});
