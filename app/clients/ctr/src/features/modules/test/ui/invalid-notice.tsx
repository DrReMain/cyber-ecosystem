import { m } from "#/paraglide/messages";

// Ops reason → user copy; the server's reason vocabulary is ops-facing, so
// the resolution lives client-side with a generic fallback.
const INVALID_REASON_COPY: Record<string, () => string> = {
  BACKGROUND_OVER_BUDGET: m.test_invalid_background,
  ABSURD_REACTION_TIME: m.test_invalid_absurd,
  ABSURD_TIME_SPAN: m.test_invalid_absurd,
  DETAIL_MISMATCH: m.test_invalid_detail,
};

function invalidCopy(reason: string | undefined): string {
  const copy = reason !== undefined ? INVALID_REASON_COPY[reason] : undefined;
  return copy?.() ?? m.test_verdict_invalid();
}

/** The invalid attempt's amber card: the verdict word plus the
 * ops-vocabulary reason resolved to user copy. */
export function InvalidNotice({ reason }: Readonly<{ reason: string | undefined }>) {
  return (
    <div className="text-status-pending rounded-control mt-3 border border-[rgba(245,158,11,0.45)] bg-transparent px-3.25 py-3 text-left text-[12px] leading-[1.85]">
      <span className="font-bold">{m.test_verdict_invalid()}</span>
      <span className="opacity-80"> · {invalidCopy(reason)}</span>
    </div>
  );
}
