import { Crown } from "lucide-react";
import { m } from "#/paraglide/messages";

/** The membership card (prototype ④): the unlocked state states the
 * standing benefit — the prototype splits it by how the unlock happened
 * (a member window speaks the membership note, a single purchase says
 * so); the locked state pitches the window (price off the SKU read face)
 * with a CTA into the wall's member purchase. */
export function MemberCard({
  memberActive,
  onPay,
  priceYuan,
  unlocked,
}: Readonly<{
  memberActive: boolean;
  onPay: () => void;
  priceYuan: string;
  unlocked: boolean;
}>) {
  if (unlocked) {
    return (
      <div className="border-line bg-surface-card shadow-card rounded-card border border-[rgba(168,85,247,0.5)] p-3.5 backdrop-blur-md">
        <div className="flex items-center justify-between">
          <span className="flex items-center gap-1.5 text-[14px] font-extrabold text-[#EDE9FE]">
            <Crown aria-hidden size={15} />
            {m.report_member_title()}
          </span>
          <span className="rounded-full bg-[rgba(34,197,94,0.18)] px-2 py-0.5 text-[10px] font-extrabold text-[#6EE7B7]">
            {m.report_member_unlocked()}
          </span>
        </div>
        <div className="text-status-soon mt-2 text-[11px] leading-[1.7]">
          {memberActive ? m.report_member_note() : m.report_member_unlocked_single()}
        </div>
      </div>
    );
  }
  return (
    <div className="rounded-card bg-surface-card shadow-card relative overflow-hidden border border-[rgba(168,85,247,0.5)] p-3.5 backdrop-blur-md">
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 bg-[linear-gradient(120deg,rgba(124,58,237,0.24),transparent_70%)]"
      />
      <div className="relative">
        <div className="flex items-center justify-between">
          <span className="flex items-center gap-1.5 text-[14px] font-extrabold text-[#EDE9FE]">
            <Crown aria-hidden size={15} />
            {m.report_member_title()}
          </span>
          <span className="text-grade-s rounded-full bg-[rgba(234,179,8,0.18)] px-2 py-0.5 text-[10px] font-extrabold">
            {m.report_member_better()}
          </span>
        </div>
        <div className="text-status-soon mt-2 text-[11px] leading-[1.7]">
          {m.report_member_pitch({ price: priceYuan })}
        </div>
        <div className="mt-2.5 flex flex-col items-center gap-1">
          <button
            className="bg-ambient-purple w-full cursor-pointer rounded-[10px] px-4 py-2.5 text-[13px] font-bold text-white active:opacity-60"
            onClick={onPay}
            type="button"
          >
            {m.report_member_cta({ price: priceYuan })}
          </button>
        </div>
      </div>
    </div>
  );
}
