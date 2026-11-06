import { useQuery as useConnectQuery } from "@connectrpc/connect-query";
import { getOrder } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/order-OrderService_connectquery";
import { listSkus } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/sku-SkuService_connectquery";
import { useQuery } from "@tanstack/react-query";
import { useNavigate, useSearch } from "@tanstack/react-router";
import { Ban, CircleCheck, Crown, TriangleAlert } from "lucide-react";
import type { ComponentType } from "react";
import { Button } from "react-aria-components";
import type { AmbientOrb } from "#/features/app/ambient/ambient-backdrop";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { localDateOf, localDateTimeOf } from "#/features/app/date";
import { fenToYuan } from "#/features/app/fen";
import { StateScreen } from "#/features/app/states/state-screen";
import { sessionQuery } from "#/features/auth/session.fn";
import { m } from "#/paraglide/messages";
import { ORDER_SCOPE_TRIAL, ORDER_STATE_CLOSED, ORDER_STATE_PAID } from "./order";
import { PayFilingFooter } from "./ui/filing-footer";

// The prototype's single state blob: one 250px orb whose color states the
// verdict before any copy is read (green paid, purple member, red failed,
// slate cancelled).
const RESULT_ORB_CLASS = "top-[-80px] left-[-60px] size-[250px]";

// The order row as the transport delivers it: WKT wrappers flattened (the
// strings plain, the fen a bigint, the stamps ISO strings) and enums as
// value-name strings — the same wire idiom the wall reads.
interface WireResultOrder {
  orderNo?: string;
  skuCode?: string;
  scope?: unknown;
  scopeRef?: string;
  channel?: string;
  amountFen?: bigint;
  state?: unknown;
  transactionId?: string;
  createdAt?: string;
  paidAt?: string;
}

type Verdict =
  | { kind: "loading" }
  | { kind: "error" }
  | { kind: "trial-ok" }
  | { kind: "member"; renew: boolean }
  | { kind: "fail" }
  | { kind: "cancel" };

type FaceVerdict = Exclude<Verdict, { kind: "loading" } | { kind: "error" }>;

// The page's whole state machine, derived: GetOrder's truth decides the
// face — paid picks the trial or member success by the order's own scope,
// closed is the cancel face, and a still-pending order is the failure face
// (nothing settled). The face parameter only picks the member success copy
// (opened vs renewed); the order row alone cannot say which it was.
function verdictOf(
  order: WireResultOrder | undefined,
  face: "open" | "renew" | undefined,
  pending: boolean,
  failed: boolean,
): Verdict {
  if (pending) {
    return { kind: "loading" };
  }
  if (failed || order === undefined) {
    return { kind: "error" };
  }
  const state = String(order.state);
  if (state === ORDER_STATE_PAID) {
    return String(order.scope) === ORDER_SCOPE_TRIAL
      ? { kind: "trial-ok" }
      : { kind: "member", renew: face === "renew" };
  }
  if (state === ORDER_STATE_CLOSED) {
    return { kind: "cancel" };
  }
  return { kind: "fail" };
}

const FACE_ART: Record<
  Exclude<Verdict["kind"], "loading" | "error">,
  {
    Icon: ComponentType<{ "aria-hidden"?: boolean; className?: string; size?: number }>;
    orb: AmbientOrb["tone"];
    titleColor: string;
  }
> = {
  "trial-ok": { Icon: CircleCheck, orb: "green", titleColor: "text-[#22C55E]" },
  member: { Icon: Crown, orb: "purple", titleColor: "text-[#A855F7]" },
  fail: { Icon: TriangleAlert, orb: "red", titleColor: "text-[#F87171]" },
  cancel: { Icon: Ban, orb: "slate", titleColor: "text-[#9CA3AF]" },
};

function ReceiptRow({ label, value }: Readonly<{ label: string; value: string }>) {
  return (
    <div className="flex items-center justify-between">
      <span>{label}</span>
      <span className="text-ink-secondary tabular-nums">{value}</span>
    </div>
  );
}

function channelLabelOf(channel: string | undefined): string {
  switch (channel) {
    case "wechat_h5":
      return m.pay_result_channel_h5();
    case "wechat_jsapi":
      return m.pay_result_channel_jsapi();
    case "wechat_native":
      return m.pay_result_channel_native();
    default:
      return channel ?? "—";
  }
}

function timeCellOf(order: WireResultOrder): string {
  const stamp = order.paidAt ?? order.createdAt;
  if (typeof stamp !== "string") {
    return "—";
  }
  const ms = Date.parse(stamp);
  return Number.isNaN(ms) ? "—" : localDateTimeOf(ms);
}

// The member sub-line's two facts: the sku row owns the day count, the
// session's standing window (refreshed by the wall's invalidation) owns
// the end date; an unreadable window speaks the bare sentence rather
// than a wrong date.
function memberSubOf(periodDays: number | undefined, memberUntilMs: number): string {
  if (Number.isNaN(memberUntilMs)) {
    return m.pay_result_sub_member_bare();
  }
  const date = localDateOf(memberUntilMs);
  return periodDays !== undefined
    ? m.pay_result_sub_member_days({ date, days: String(periodDays) })
    : m.pay_result_sub_member({ date });
}

function titleOf(verdict: Verdict): string {
  switch (verdict.kind) {
    case "member":
      return verdict.renew ? m.pay_result_title_renew() : m.pay_result_title_member();
    case "trial-ok":
      return m.pay_result_title_ok();
    case "fail":
      return m.pay_result_title_fail();
    default:
      return m.pay_result_title_cancel();
  }
}

function subOf(verdict: Verdict, periodDays: number | undefined, memberUntilMs: number): string {
  switch (verdict.kind) {
    case "member":
      return memberSubOf(periodDays, memberUntilMs);
    case "trial-ok":
      return m.pay_result_sub_ok();
    case "fail":
      return m.pay_result_sub_fail();
    default:
      return m.pay_result_sub_cancel();
  }
}

// Where a settled purchase returns: the wall's from passthrough wins, else
// the order itself names the target — a trial's unlocked record, a
// membership the report hub.
function viewHrefOf(order: WireResultOrder, from: string | undefined): string {
  if (from !== undefined) {
    return from;
  }
  return String(order.scope) === ORDER_SCOPE_TRIAL && order.scopeRef !== undefined
    ? `/report/${order.scopeRef}`
    : "/report/center";
}

// The retry link back into the wall: a trial order carries its record as
// the scope target, and the from passthrough keeps the round trip whole.
function retrySearchOf(
  order: WireResultOrder | undefined,
  from: string | undefined,
): { ref?: string; from?: string } {
  return {
    ...(String(order?.scope) === ORDER_SCOPE_TRIAL && order?.scopeRef !== undefined
      ? { ref: order.scopeRef }
      : {}),
    ...(from !== undefined ? { from } : {}),
  };
}

// The drawn face, one component: the verdict art, the success receipt,
// the failure/cancel notes, and the button pair — every conditional flat,
// the parent only decides which face exists.
function ResultFace({
  onHome,
  onMember,
  onRetry,
  onView,
  order,
  sub,
  title,
  verdict,
}: Readonly<{
  onHome: () => void;
  onMember: () => void;
  onRetry: () => void;
  onView: () => void;
  order: WireResultOrder;
  sub: string;
  title: string;
  verdict: FaceVerdict;
}>) {
  const success = verdict.kind === "trial-ok" || verdict.kind === "member";
  const { Icon, orb, titleColor } = FACE_ART[verdict.kind];
  return (
    <AmbientLayout orbs={[{ className: RESULT_ORB_CLASS, tone: orb }]}>
      <div className="flex w-full flex-col gap-3.5 pb-6">
        <div className="pt-8 text-center">
          <Icon aria-hidden className={titleColor} size={44} />
          <h1 className={`mt-2 text-[19px] font-black ${titleColor}`}>{title}</h1>
          <p className="text-ink-muted mt-2 text-[12.5px] leading-[1.8]">{sub}</p>
        </div>

        {success ? (
          <div className="border-line bg-surface-card shadow-card rounded-card text-ink-tertiary border p-3.5 text-[12px] leading-[1.9]">
            <ReceiptRow label={m.pay_result_row_order()} value={order.orderNo ?? "—"} />
            <ReceiptRow
              label={m.pay_result_row_amount()}
              value={order.amountFen !== undefined ? `¥${fenToYuan(order.amountFen)}` : "—"}
            />
            <ReceiptRow label={m.pay_result_row_channel()} value={channelLabelOf(order.channel)} />
            <ReceiptRow label={m.pay_result_row_time()} value={timeCellOf(order)} />
          </div>
        ) : null}

        {verdict.kind === "fail" ? (
          <div className="rounded-card bg-surface-card border border-[rgba(245,158,11,0.45)] p-3.5 text-[11.5px] leading-[1.85] text-[#FCD34D]">
            {m.pay_result_tip_fail()}
          </div>
        ) : null}

        {verdict.kind === "cancel" ? (
          <div className="border-line bg-surface-card rounded-card text-ink-tertiary border p-3.5 text-[11.5px] leading-[1.85]">
            {m.pay_result_tip_cancel()}
          </div>
        ) : null}

        {success ? (
          <Button
            className="bg-cta-primary shadow-cta-primary w-full rounded-[12px] py-[14px] text-[15px] font-bold text-white active:opacity-80"
            onPress={onView}
          >
            {verdict.kind === "member" ? m.pay_result_btn_view_member() : m.pay_result_btn_view()}
          </Button>
        ) : (
          <Button
            className="bg-cta-practice shadow-cta-practice w-full rounded-[12px] py-[14px] text-[15px] font-bold text-white active:opacity-80"
            onPress={onRetry}
          >
            {m.pay_result_btn_retry()}
          </Button>
        )}
        <Button
          className="bg-surface-input text-ink-secondary w-full rounded-[12px] border-[1.5px] border-[#4B5563] py-[14px] text-[15px] font-medium active:opacity-80"
          onPress={onHome}
        >
          {m.pay_result_btn_home()}
        </Button>
        {verdict.kind === "member" ? (
          <Button
            className="bg-surface-input text-ink-secondary w-full rounded-[12px] border-[1.5px] border-[#4B5563] py-[14px] text-[15px] font-medium active:opacity-80"
            onPress={onMember}
          >
            {m.pay_result_btn_member()}
          </Button>
        ) : null}

        <div className="text-ink-tertiary text-center text-[10.5px] leading-[1.7]">
          {m.pay_result_note()}
        </div>
        <PayFilingFooter />
      </div>
    </AmbientLayout>
  );
}

/** The pay result (prototype pay-result, the receipt every purchase exit
 * lands on): the order's own truth draws one of four faces — single
 * success, member success (opened or renewed — the wall says which, D-AU's
 * third parameter), still-pending failure with the same-order retry, or
 * cancelled. A failed read renders the honest error screen, never an
 * invented verdict. */
export function PayResultPage() {
  const search = useSearch({ from: "/pay/result" });
  const navigate = useNavigate();
  const orderQuery = useConnectQuery(
    getOrder,
    { orderNo: search.order },
    // The verdict reads this query's error as its honest error face; the
    // global throw default would boundary it instead.
    { enabled: search.order !== "", throwOnError: false },
  );
  const skusQuery = useConnectQuery(listSkus, {});
  const session = useQuery({ ...sessionQuery, throwOnError: false });

  const order = orderQuery.data?.order as WireResultOrder | undefined;
  // A malformed entry (no order number on the wire) never starts the
  // query — it renders the error face directly, not an eternal loading.
  const noOrder = search.order === "";
  const verdict = verdictOf(
    order,
    search.face,
    orderQuery.isPending && !noOrder,
    orderQuery.isError || noOrder,
  );
  const drawn = verdict.kind === "loading" || verdict.kind === "error" ? null : verdict;

  // The member sub-line's inputs, narrowed at the seam: the catalog row
  // by the order's own sku code, the standing window off the session. A
  // session still refetching may hold the pre-purchase window (the wall
  // invalidates and navigates in one tick) — the bare sentence renders
  // until the read settles, never a stale date on a receipt.
  const skuRow = skusQuery.data?.skus?.find((s) => s.skuCode === order?.skuCode);
  const periodDays = typeof skuRow?.periodDays === "number" ? skuRow.periodDays : undefined;
  const memberUntil =
    !session.isFetching &&
    session.data?.status === "authed" &&
    typeof session.data.memberUntil === "string"
      ? Date.parse(session.data.memberUntil)
      : Number.NaN;

  function goHome() {
    void navigate({ to: "/" });
  }

  if (verdict.kind === "loading" || drawn === null) {
    return (
      <AmbientLayout>
        <StateScreen variant="loading" />
      </AmbientLayout>
    );
  }
  if (verdict.kind === "error") {
    return (
      <AmbientLayout>
        <StateScreen
          action={
            <Button
              className="bg-cta-primary rounded-[12px] px-5 py-2.5 text-sm font-semibold text-white active:opacity-80"
              onPress={goHome}
            >
              {m.pay_result_error_home()}
            </Button>
          }
          description={m.pay_result_error_desc()}
          title={m.pay_result_error_title()}
          variant="empty"
        />
      </AmbientLayout>
    );
  }

  return (
    <ResultFace
      onHome={goHome}
      onMember={() => void navigate({ to: "/member" })}
      onRetry={() =>
        void navigate({
          to: "/pay",
          search: retrySearchOf(order, search.from),
        })
      }
      onView={() => void navigate({ href: viewHrefOf(order as WireResultOrder, search.from) })}
      order={order as WireResultOrder}
      sub={subOf(drawn, periodDays, memberUntil)}
      title={titleOf(drawn)}
      verdict={drawn}
    />
  );
}
