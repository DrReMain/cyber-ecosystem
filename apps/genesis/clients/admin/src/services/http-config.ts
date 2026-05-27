import { env, resolveApiBaseUrl } from "#/env"
import type { CreateClientConfig } from "#/services/openapi/client.gen"

export const createClientConfig: CreateClientConfig = (config) => ({
  ...config,
  baseUrl: resolveApiBaseUrl(env.HTTP_API_URL, "/http"),
  retry: 0,
  kyOptions: { throwHttpErrors: false },
})
