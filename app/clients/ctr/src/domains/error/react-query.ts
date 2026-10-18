import type { ErrorPolicy } from "@cyber-ecosystem/shared-error";
import { MutationCache, QueryCache } from "@tanstack/react-query";
import { errorDomain } from ".";

declare module "@tanstack/react-query" {
  interface Register {
    queryMeta: ErrorPolicy;
    mutationMeta: ErrorPolicy;
  }
}

export function errorQueryCache(): QueryCache {
  return new QueryCache({
    onError: (error, query) =>
      errorDomain.capture(error, {
        feedback: query.meta?.feedback ?? false,
        report: query.meta?.report,
      }),
  });
}

export function errorMutationCache(): MutationCache {
  return new MutationCache({
    onError: (error, _variables, _result, mutation) =>
      errorDomain.capture(error, {
        feedback: mutation.meta?.feedback ?? true,
        report: mutation.meta?.report,
      }),
  });
}

export function shouldThrow(error: unknown): boolean {
  return !errorDomain.view(error).silent;
}

export function shouldRetry(failureCount: number, error: unknown): boolean {
  return failureCount < 3 && errorDomain.view(error).kind !== "canceled";
}
