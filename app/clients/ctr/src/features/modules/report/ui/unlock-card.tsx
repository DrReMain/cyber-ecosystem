import { Lock } from "lucide-react";
import { m } from "#/paraglide/messages";

/** The one-shot unlock card (prototype ⑤, paystripe): what the unlock
 * buys, the price straight off the SKU read face (never a hardcoded
 * number), and the permanence note. The CTA stays parked until the pay
 * client lands — an honest "coming" note, never a dead button. */
export function UnlockCard({
  count,
  priceYuan,
}: Readonly<{
  count: number;
  priceYuan: string;
}>) {
  return (
    <div
      className="border-line bg-surface-card shadow-card rounded-card border p-4 backdrop-blur-md"
      style={{
        backgroundImage:
          "repeating-linear-gradient(45deg, rgba(255,255,255,0.05) 0 8px, transparent 8px 16px)",
      }}
    >
      <div className="flex items-center justify-between">
        <span className="text-ink text-sm font-extrabold">
          {m.report_unlock_title({ n: count })}
        </span>
        <Lock aria-hidden className="text-[#FDBA74]" size={16} />
      </div>
      <div className="text-ink-tertiary mt-2 text-[11px] leading-[1.7]">
        {m.report_unlock_desc({ n: count })}
      </div>
      <div className="mt-3 flex items-baseline justify-center gap-2">
        <span className="text-[30px] leading-none font-black text-[#FDE047]">¥{priceYuan}</span>
        <span className="text-ink-tertiary text-[11px]">{m.report_unlock_once()}</span>
      </div>
      <div className="mt-3 flex flex-col items-center gap-1.5">
        <button
          className="w-full cursor-pointer rounded-[10px] bg-[#F97316] px-4 py-3 text-[15px] font-bold text-white opacity-70 active:opacity-60"
          disabled
          type="button"
        >
          {m.report_unlock_cta({ price: priceYuan })}
        </button>
        <div className="text-ink-tertiary text-[10px]">{m.report_pay_soon()}</div>
      </div>
      <div className="text-ink-tertiary mt-2 text-center text-[10px]">
        {m.report_unlock_forever()}
      </div>
    </div>
  );
}
