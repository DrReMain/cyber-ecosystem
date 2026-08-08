import type { Mutation, Query } from "@tanstack/react-query";
import { errorHandler } from "./error-handler";

export function onQueryError(
  error: unknown,
  _query: Query<unknown, unknown, unknown, readonly unknown[]>,
): void {
  errorHandler.handle(error, "query", { feedback: false });
}

export function onMutationError(
  error: unknown,
  _mutation: Mutation<unknown, unknown, unknown, unknown>,
): void {
  errorHandler.handle(error, "mutation");
}
