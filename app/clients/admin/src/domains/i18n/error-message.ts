import type { AppError } from "#/domains/errors/app-error";
import * as m from "#/paraglide/messages";

// Copy projection: AppError identity → human text via the paraglide seeds
// (`error_<REASON>` keys; fallback tiers error_network /error_unknown).
// Reason keys are dynamic (UPPER_SNAKE wire names arrive at
// runtime) so they go through a record view; the two fallback tiers are our
// own seeds and are accessed by name.
const table = m as unknown as Record<string, (() => string) | undefined>;

// Transport layers prefix messages with the code name ("[unauthenticated] …")
// — e.g. a ConnectError crossing the server-fn boundary keeps only such a
// message. Strip the prefix at the PRESENTATION tier (identity stays
// normalize's business); if nothing human remains, fall through to the
// unknown tier instead of showing a bare code name.
function stripCodePrefix(message: string): string {
  return message.replace(/^\[[^\]]*\]\s*/, "");
}

export function errorMessage(error: AppError): string {
  if (error.reason !== undefined) {
    const copy = table[`error_${error.reason}`];
    if (copy !== undefined) return copy();
  }
  if (error.network) return m.error_network();
  const message = stripCodePrefix(error.message);
  if (message !== "") return message;
  return m.error_unknown();
}
