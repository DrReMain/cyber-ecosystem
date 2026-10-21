import type { ErrorPolicy } from "@cyber-ecosystem/shared-error";
import { MutationCache, QueryCache } from "@tanstack/react-query";
import { errorDomain } from ".";

declare module "@tanstack/react-query" {
  interface Register {
    queryMeta: ErrorPolicy;
    mutationMeta: ErrorPolicy;
  }
}

// An auth-kind fault means the server-side session truth changed while the
// cached probe may still believe otherwise — the wiring layer passes a
// re-verification hook so the one session truth re-runs immediately.
export function errorQueryCache(onAuthFault?: () => void): QueryCache {
  return new QueryCache({
    onError: (error, query) => {
      errorDomain.capture(error, {
        feedback: query.meta?.feedback ?? false,
        report: query.meta?.report,
      });
      if (onAuthFault !== undefined && errorDomain.view(error).kind === "auth") {
        onAuthFault();
      }
    },
  });
}

export function errorMutationCache(onAuthFault?: () => void): MutationCache {
  return new MutationCache({
    onError: (error, _variables, _result, mutation) => {
      errorDomain.capture(error, {
        feedback: mutation.meta?.feedback ?? true,
        report: mutation.meta?.report,
      });
      if (onAuthFault !== undefined && errorDomain.view(error).kind === "auth") {
        onAuthFault();
      }
    },
  });
}

// A dead session is session state, never a page error: auth faults never
// escalate into boundaries — the session watcher owns the exit, and the
// re-verification hook above is what wakes it.
export function shouldThrow(error: unknown): boolean {
  const view = errorDomain.view(error);
  return !view.silent && view.kind !== "auth";
}

export function shouldRetry(failureCount: number, error: unknown): boolean {
  const view = errorDomain.view(error);
  return failureCount < 3 && view.kind !== "canceled" && view.kind !== "auth";
}
