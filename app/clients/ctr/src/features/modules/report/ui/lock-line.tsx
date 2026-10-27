import { Lock } from "lucide-react";
import { m } from "#/paraglide/messages";

/** The locked detail's teaser line: blurred copy under a lock mask — the
 * axis name and score stay public, the analysis behind them does not. */
export function LockLine() {
  return (
    <div className="border-line relative mt-2 h-[30px] overflow-hidden rounded-md border">
      <div
        aria-hidden
        className="text-ink-tertiary px-2.5 text-[10.5px] leading-[30px] blur-[3px] select-none"
      >
        {m.report_lock_blur()}
      </div>
      <div className="absolute inset-0 flex items-center justify-center gap-1.5 bg-[rgba(17,24,39,0.5)] text-[#FDBA74]">
        <Lock aria-hidden size={11} strokeWidth={2} />
        <span className="text-[10px] font-semibold">{m.report_lock_cta()}</span>
      </div>
    </div>
  );
}
