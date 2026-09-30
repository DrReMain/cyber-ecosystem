import type { DescMethod, DescService } from "@bufbuild/protobuf";
import { makeEnvelope } from "@cyber-ecosystem/shared-error";

// Derives an operation string from a generated proto service descriptor, so
// typos in service or method names fail at compile time. Method keys are the
// camelCase local names of the generated service object (e.g. "listUsers").
export function op<S extends DescService>(service: S, method: keyof S["method"] & string): string {
  const desc = (service.method as Record<string, DescMethod>)[method];
  if (!desc) throw new Error(`unknown method "${method}" on ${service.typeName}`);
  return `/${service.typeName}/${desc.name}`;
}

export function isOperationAllowed(patterns: readonly string[], operation: string): boolean {
  return patterns.some(
    (pattern) =>
      operation === pattern ||
      (pattern.endsWith("/*") && operation.startsWith(pattern.slice(0, -1))),
  );
}

// Route guard for direct links: routes without operations stay public.
// Thrown pre-normalized so every consumer shares one pipeline.
export function requireOperations(
  operations: readonly string[] | undefined,
  patterns: readonly string[],
): void {
  for (const operation of operations ?? []) {
    if (!isOperationAllowed(patterns, operation)) {
      throw makeEnvelope("forbidden", operation);
    }
  }
}
