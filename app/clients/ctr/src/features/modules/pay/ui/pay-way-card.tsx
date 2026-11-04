import { Check, CircleCheck } from "lucide-react";

/** The 态② payment-method row: WeChat Pay selected — the only
 * channel the current window serves, so the row states it with the
 * selected check rather than offering a choice. */
export function PayWayCard({ title, way }: Readonly<{ title: string; way: string }>) {
  return (
    <div className="border-line bg-surface-card shadow-card rounded-card border p-4 backdrop-blur-md">
      <div className="text-ink mb-2.5 text-[13px] font-bold">{title}</div>
      <div className="flex items-center justify-between rounded-[12px] border border-[rgba(34,197,94,0.35)] bg-[rgba(34,197,94,0.08)] px-3.5 py-3">
        <span className="flex items-center gap-2.5">
          <span className="flex size-[34px] items-center justify-center rounded-lg bg-[rgba(34,197,94,0.18)]">
            <CircleCheck aria-hidden className="text-[#6EE7B7]" size={18} />
          </span>
          <span className="text-ink text-[14px] font-bold">{way}</span>
        </span>
        <Check aria-hidden className="text-[#22C55E]" size={18} strokeWidth={2.5} />
      </div>
    </div>
  );
}
