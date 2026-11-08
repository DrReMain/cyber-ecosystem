import { useQuery as useConnectQuery } from "@connectrpc/connect-query";
import { getAbilityDigest } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/report-ReportService_connectquery";
import { listSkus } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/sku-SkuService_connectquery";
import { useNavigate, useRouter } from "@tanstack/react-router";
import { ChevronLeft, Crown } from "lucide-react";
import { Button } from "react-aria-components";
import type { AmbientOrb } from "#/features/app/ambient/ambient-backdrop";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { localDateOf } from "#/features/app/date";
import { StateScreen } from "#/features/app/states/state-screen";
import { purchaseFaces } from "#/features/modules/pay/order";
import { wireDigestOf } from "#/features/modules/report/digest";
import { m } from "#/paraglide/messages";

// The prototype's single state blob: one 250px purple orb, right-hung —
// this page's own arrangement (the receipt page hangs its verdict orb
// left).
const DETAIL_ORBS: readonly AmbientOrb[] = [
  { className: "-top-20 -right-17.5 size-62.5", tone: "purple" },
];

// The glass card the whole page speaks in.
const GLASS_CLASS =
  "border-line bg-surface-card shadow-card rounded-card border p-4 backdrop-blur-md";

type MemberWindow =
  | { kind: "never" }
  | { kind: "active"; untilMs: number }
  | { kind: "expired"; untilMs: number };

// The three-state fold off the digest's window fact: a parseable future
// end keeps the member face, a past one the expired face with the date it
// ended, anything else the never-bought pitch. member_until cannot draw
// this — its liveness read drops a lapsed window to null.
function windowOf(lastUntil: string | undefined): MemberWindow {
  if (lastUntil === undefined) {
    return { kind: "never" };
  }
  const ms = Date.parse(lastUntil);
  if (Number.isNaN(ms)) {
    return { kind: "never" };
  }
  return ms > Date.now() ? { kind: "active", untilMs: ms } : { kind: "expired", untilMs: ms };
}

// The month card: the catalog's own price and day count, plus the status
// strip the standing window states — purple when live, gray quoting the
// end date when lapsed, absent when never bought.
function MonthCard({
  periodDays,
  priceYuan,
  window: windowState,
}: Readonly<{ periodDays: number | undefined; priceYuan: string; window: MemberWindow }>) {
  return (
    <div className={`${GLASS_CLASS} border-[rgba(126,34,206,0.55)] py-5`}>
      <div className="flex items-start justify-between">
        <div>
          <div className="text-ink text-[15px] font-extrabold">{m.member_card_name()}</div>
          <div className="text-ink-tertiary mt-1 text-[11.5px]">
            {periodDays !== undefined
              ? m.member_card_days({ days: String(periodDays) })
              : m.member_card_days_bare()}
          </div>
        </div>
        <div className="text-right">
          <div className="text-grade-s text-[24px] leading-none font-black tabular-nums">
            ¥{priceYuan}
          </div>
          <div className="text-ink-tertiary mt-1 text-[10px]">{m.member_card_per_month()}</div>
        </div>
      </div>
      {windowState.kind === "active" ? (
        <div className="text-status-member mt-3.5 flex items-center gap-1.5 rounded-[12px] border border-[rgba(126,34,206,0.5)] bg-[rgba(126,34,206,0.12)] px-3 py-2.5 text-[12px]">
          <Crown aria-hidden size={13} />
          {m.member_status_active({ date: localDateOf(windowState.untilMs) })}
        </div>
      ) : null}
      {windowState.kind === "expired" ? (
        <div className="border-line bg-surface-input text-status-lock mt-3.5 rounded-[12px] border px-3 py-2.5 text-[12px]">
          {m.member_status_expired({ date: localDateOf(windowState.untilMs) })}
        </div>
      ) : null}
    </div>
  );
}

// The four benefit rows. The fourth is a live-catalog fact (the day
// count) — a degraded catalog read keeps the row honestly absent rather
// than inventing a number.
function BenefitCard({ periodDays }: Readonly<{ periodDays: number | undefined }>) {
  const rows = [
    { sub: m.member_benefit_1_sub(), title: m.member_benefit_1_title() },
    { sub: m.member_benefit_2_sub(), title: m.member_benefit_2_title() },
    { sub: m.member_benefit_3_sub(), title: m.member_benefit_3_title() },
    ...(periodDays !== undefined
      ? [
          {
            sub: m.member_benefit_4_sub({ next: String(periodDays + 1) }),
            title: m.member_benefit_4_title({ days: String(periodDays) }),
          },
        ]
      : []),
  ];
  return (
    <div className={GLASS_CLASS}>
      <div className="text-ink mb-1 text-[13px] font-bold">{m.member_benefit_title()}</div>
      <div className="divide-y divide-[rgba(55,65,81,0.4)]">
        {rows.map((row) => (
          <div className="flex gap-2.5 py-2.25" key={row.title}>
            <Crown aria-hidden className="mt-0.75 shrink-0 text-[#A855F7]" size={14} />
            <div>
              <div className="text-[12.5px] font-bold text-[#E5E7EB]">{row.title}</div>
              <div className="text-ink-tertiary mt-0.75 text-[10.5px] leading-[1.6]">{row.sub}</div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

// The five FAQ pairs — static marketing copy, prices quoted verbatim.
function FaqCard() {
  const rows = [
    { a: m.member_faq_1_a(), q: m.member_faq_1_q() },
    { a: m.member_faq_2_a(), q: m.member_faq_2_q() },
    { a: m.member_faq_3_a(), q: m.member_faq_3_q() },
    { a: m.member_faq_4_a(), q: m.member_faq_4_q() },
    { a: m.member_faq_5_a(), q: m.member_faq_5_q() },
  ];
  return (
    <div className={GLASS_CLASS}>
      <div className="text-ink mb-1 text-[13px] font-bold">{m.member_faq_title()}</div>
      <div className="divide-y divide-[rgba(55,65,81,0.4)]">
        {rows.map((row) => (
          <div className="py-2.5" key={row.q}>
            <div className="text-[12.5px] font-bold text-[#E5E7EB]">{row.q}</div>
            <div className="text-ink-muted mt-1.25 text-[11.5px] leading-[1.85]">{row.a}</div>
          </div>
        ))}
      </div>
    </div>
  );
}

/** The membership benefits page (prototype member-detail, matrix 23): the
 * digest's window fact draws one of three states — a live window's purple
 * strip, a lapsed window's gray strip quoting the end date, or no strip
 * for a never-member — over the month card, the four benefits, the five
 * FAQ pairs, and the purple CTA into the wall. The wall's ref-less entry
 * is member-only, so the CTA needs no record scope; the filing footer
 * itself stays on the pay page, this page points at it. */
export function MemberDetailPage() {
  const navigate = useNavigate();
  const router = useRouter();
  const digestQuery = useConnectQuery(getAbilityDigest, {}, { throwOnError: false });
  const skusQuery = useConnectQuery(listSkus, {}, { throwOnError: false });

  const digest =
    digestQuery.isSuccess && digestQuery.data !== undefined ? wireDigestOf(digestQuery.data) : null;
  const member = purchaseFaces(skusQuery.data?.skus ?? []).member;
  const periodDays = typeof member.row?.periodDays === "number" ? member.row.periodDays : undefined;

  function backOrHome() {
    if (router.history.canGoBack()) {
      router.history.back();
    } else {
      void navigate({ to: "/" });
    }
  }

  function goHome() {
    void navigate({ to: "/" });
  }

  if (digestQuery.isPending) {
    return (
      <AmbientLayout orbs={DETAIL_ORBS}>
        <StateScreen variant="loading" />
      </AmbientLayout>
    );
  }
  if (digest === null) {
    return (
      <AmbientLayout orbs={DETAIL_ORBS}>
        <StateScreen
          action={
            <Button
              className="bg-cta-primary rounded-[12px] px-5 py-2.5 text-sm font-semibold text-white active:opacity-80"
              onPress={goHome}
            >
              {m.member_error_home()}
            </Button>
          }
          title={m.member_error_title()}
          variant="empty"
        />
      </AmbientLayout>
    );
  }

  const windowState = windowOf(digest.lastMembershipUntil);
  return (
    <AmbientLayout orbs={DETAIL_ORBS}>
      <div className="flex w-full flex-col gap-3.5 pb-6">
        {/* Inline header — the wall's local-declaration pattern: the 40px
         * back fab and the page's own title. */}
        <div className="flex items-center gap-3 pt-5">
          <button
            aria-label={m.common_back()}
            className="bg-surface-input text-ink-secondary flex h-10 w-10 shrink-0 items-center justify-center rounded-full active:opacity-60"
            onClick={backOrHome}
            type="button"
          >
            <ChevronLeft aria-hidden size={20} />
          </button>
          <h1 className="text-ink text-[19px] font-bold">{m.member_detail_title()}</h1>
        </div>

        <MonthCard periodDays={periodDays} priceYuan={member.yuan} window={windowState} />
        <BenefitCard periodDays={periodDays} />
        <FaqCard />
        <Button
          className="bg-ambient-purple w-full rounded-[12px] py-3.5 text-[15px] font-bold text-white active:opacity-80 disabled:opacity-50"
          isDisabled={member.row === undefined}
          onPress={() => void navigate({ to: "/pay", search: { from: "/member" } })}
        >
          {windowState.kind === "active"
            ? m.member_cta_renew({ price: member.yuan })
            : m.member_cta_open({ price: member.yuan })}
        </Button>
        <div className="text-ink-tertiary text-center text-[10.5px] leading-[1.7]">
          {m.member_filing_note()}
        </div>
      </div>
    </AmbientLayout>
  );
}
