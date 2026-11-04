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

// The pay result's URL shape: the order whose truth the page reads, the
// member face's copy fork (D-AU's third parameter — the wall knows whether
// the purchase opened or renewed, the order row alone cannot say), and the
// wall's own from passthrough so the receipt page returns where the
// purchase began.
export interface PayResultSearch {
  order: string;
  face?: "open" | "renew";
  from?: string;
}

export function payResultSearchOf(search: Record<string, unknown>): PayResultSearch {
  return {
    // "" = a malformed entry (no order on the wire); the page's query
    // stays disabled and the error face answers.
    order: typeof search.order === "string" ? search.order : "",
    ...(search.face === "open" || search.face === "renew" ? { face: search.face } : {}),
    ...(typeof search.from === "string" &&
    search.from.startsWith("/") &&
    !search.from.startsWith("//")
      ? { from: search.from }
      : {}),
  };
}
