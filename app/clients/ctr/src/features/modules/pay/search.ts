// The pay wall's URL shape: the test record a trial unlock targets (the
// order's scope_ref), and the internal path to land on after a single
// unlock pays.
export interface PaySearch {
  ref?: string;
  from?: string;
}

export function paySearchOf(search: Record<string, unknown>): PaySearch {
  return {
    ...(typeof search.ref === "string" && search.ref !== "" ? { ref: search.ref } : {}),
    // Same-origin only, the login next guard's rule: protocol-relative
    // "//host" never passes.
    ...(typeof search.from === "string" &&
    search.from.startsWith("/") &&
    !search.from.startsWith("//")
      ? { from: search.from }
      : {}),
  };
}
