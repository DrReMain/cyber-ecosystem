import type { Timestamp } from "@bufbuild/protobuf/wkt";
import { Crown } from "lucide-react";
import { localDateOf } from "#/features/app/date";
import { m } from "#/paraglide/messages";

// The profile's own member-face colors: the purple identity the prototype
// paints this card with, the green of a standing window.
const CARD_CLASS =
  "border-line bg-surface-card shadow-card rounded-card border border-[rgba(168,85,247,0.5)] p-3.5 backdrop-blur-md";
const ACTIVE_TAG =
  "rounded-full bg-[rgba(34,197,94,0.18)] px-2 py-0.5 text-[10px] font-extrabold text-[#6EE7B7]";
const MONTH_TAG =
  "rounded-full bg-[rgba(124,58,237,0.22)] px-2 py-0.5 text-[10px] font-extrabold text-[#D8B4FE]";
const CTA_CLASS =
  "shrink-0 cursor-pointer rounded-[10px] bg-[#7C3AED] px-[14px] py-[7px] text-[12px] font-bold text-white opacity-70 active:opacity-60";

// The wire stamps a Timestamp two ways: the JSON transport flattens it to
// an ISO scalar, a hydrated message carries epoch seconds — read both,
// never trust one shape (the agreement face's own law). An unreadable
// stamp falls to the em dash, never a thrown face.
function dayOf(value: Timestamp): string {
  const v = value as unknown as string | { seconds?: unknown };
  const ms = typeof v === "string" ? new Date(v).getTime() : Number(v.seconds ?? Number.NaN) * 1000;
  return Number.isNaN(ms) ? "—" : localDateOf(ms);
}

/** The profile's membership face (prototype ③), its own card apart from
 * the report page's ④: the member state states the standing window off
 * member_until (GetProfile carries it — the page's own query, zero extra
 * hop), the pitch state sells the window at the catalog's price. Both
 * CTAs stay parked until the pay client lands — the same stance the
 * purchase faces hold; the 详情 link rides with the member-detail page. */
export function ProfileMemberCard({
  memberUntil,
  priceYuan,
}: Readonly<{
  memberUntil: Timestamp | null;
  priceYuan: string;
}>) {
  const active = memberUntil != null;
  const until = active ? dayOf(memberUntil) : null;
  return (
    <div className={CARD_CLASS}>
      <div className="flex items-center gap-2">
        <span className="flex items-center gap-1.5 text-[14px] font-extrabold text-[#EDE9FE]">
          <Crown aria-hidden size={15} />
          {m.profile_member_title()}
        </span>
        {active ? (
          <span className={ACTIVE_TAG}>{m.profile_member_tag_active()}</span>
        ) : (
          <span className={MONTH_TAG}>{m.profile_member_tag_month()}</span>
        )}
        {active ? (
          <span className="ml-auto text-[10.5px] text-[#D8B4FE]">
            {m.profile_member_until({ date: until ?? "" })}
          </span>
        ) : null}
      </div>
      <div className="text-ink-tertiary mt-2 text-[11.5px] leading-[1.75]">
        {active ? m.profile_member_desc_active() : m.profile_member_desc_pitch()}
      </div>
      <div className="mt-[11px] flex items-center justify-between">
        {active ? (
          <span className="text-ink-tertiary text-[10px]">{m.report_pay_soon()}</span>
        ) : (
          <span>
            <span className="text-[19px] leading-none font-black text-[#FDE047]">¥{priceYuan}</span>
            <span className="text-ink-tertiary text-[11px]">{m.profile_member_per_month()}</span>
          </span>
        )}
        <button className={CTA_CLASS} disabled type="button">
          {active ? m.profile_member_cta_renew() : m.profile_member_cta_open()}
        </button>
      </div>
    </div>
  );
}
