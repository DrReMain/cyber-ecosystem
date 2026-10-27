import { m } from "#/paraglide/messages";
import type { DeltaWord } from "../vocabulary";

/** The unlocked detail's advice box: the five-zone improvement wording
 * with the last-attempt delta riding beside it. Markup-only. */
export function AdviceBox({
  advice,
  delta,
}: Readonly<{
  advice: string;
  delta: DeltaWord | null;
}>) {
  const tone =
    delta?.tone === "up"
      ? "text-status-ok"
      : delta?.tone === "down"
        ? "text-[#F87171]"
        : "text-ink-tertiary";
  return (
    <div className="mt-2 rounded-lg border border-[rgba(55,65,81,0.5)] bg-[rgba(255,255,255,0.03)] px-2.5 py-2">
      <span className="text-[11px] font-bold text-[#FDE047]">{m.report_advice_title()}</span>
      <span className="text-ink-tertiary text-[11px]"> · {advice}</span>
      {delta !== null ? (
        <span className={`ml-1.5 text-[11px] font-bold ${tone}`}>{delta.text}</span>
      ) : null}
    </div>
  );
}
