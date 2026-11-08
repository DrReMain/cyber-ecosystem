import { useQuery as useConnectQuery, useMutation } from "@connectrpc/connect-query";
import { markPaid } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/mock_cashier-MockCashierService_connectquery";
import {
  closeOrder,
  createOrder,
  getOrder,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/order-OrderService_connectquery";
import { listSkus } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/sku-SkuService_connectquery";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useNavigate, useRouter, useSearch } from "@tanstack/react-router";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { useEffect, useState } from "react";
import { Button } from "react-aria-components";
import type { AmbientOrb } from "#/features/app/ambient/ambient-backdrop";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { ageLineReached, ageOf } from "#/features/app/birth";
import { localDateOf } from "#/features/app/date";
import { BottomConfirmSheet, type SheetAction } from "#/features/app/dialog/bottom-confirm-sheet";
import { sessionQuery } from "#/features/auth/session.fn";
import { m } from "#/paraglide/messages";
import {
  CREDENTIAL_INVOKE,
  CREDENTIAL_JUMP_URL,
  CREDENTIAL_QR_CODE,
  channelOf,
  invokeWechatPay,
  ORDER_STATE_CLOSED,
  ORDER_STATE_PAID,
  type PurchaseFace,
  purchaseFaces,
} from "./order";
import type { PaySearch } from "./search";
import { BirthCard } from "./ui/birth-card";
import { PayFilingFooter } from "./ui/filing-footer";
import { MinorCard } from "./ui/minor-card";
import { OptionCard } from "./ui/option-card";
import { PayWayCard } from "./ui/pay-way-card";
import { QrTile } from "./ui/qr-tile";
import { StateBanner } from "./ui/state-banner";

// 态②'s two-orb art direction — the membership tint up top, the cyan
// floor — the dual-option states ride the shared login trio instead.
const RENEW_ORBS: readonly AmbientOrb[] = [
  { className: "-top-15 -left-10 size-70", tone: "purple" },
  { className: "-bottom-10 -right-15 size-55", tone: "cyan" },
];

type PurchaseKind = "member" | "trial";

// The receipt's member copy fork (opened vs renewed). Captured at click
// time — the standing window the purchase started under is the fact; a
// session refetch mid-poll must not rewrite the verdict the receipt quotes.
type MemberFace = "open" | "renew";

interface SheetState {
  face: MemberFace;
  orderNo: string;
  price: string;
  kind: PurchaseKind;
}

interface QrState {
  face: MemberFace;
  kind: PurchaseKind;
  orderNo: string;
  value: string;
}

// The session probe's authed shape, read structurally — the wall needs
// the birth pair and the window history (liveness is lastMemberUntil's
// own derived fact here: future end = active, past end = expired).
interface SessionView {
  status: string;
  lastMemberUntil?: string | null;
  profile?: { birthYear?: number; birthMonth?: number };
}

// The create answer, read structurally: the provider routes mock vs real
// continuation, the credential kind picks the real leg.
interface CreateResult {
  order?: { orderNo?: string; provider?: string } | undefined;
  credential?: { kind?: unknown; value?: string } | undefined;
}

interface PayGateView {
  payAgeLine?: number;
  minorPrompt?: unknown;
  missingBirthPrompt?: unknown;
}

interface WallState {
  birthMissing: boolean;
  minor: boolean;
  memberActive: boolean;
  memberExpired: boolean;
  untilDate: string;
  ageLine: number;
  ageNow: number;
  minorPrompt?: string;
  missingBirthPrompt?: string;
}

// The standing window: a parseable future date keeps the member face,
// anything else that exists is an expired one.
function windowOf(memberUntil: string | null | undefined): {
  active: boolean;
  expired: boolean;
  untilDate: string;
} {
  if (memberUntil == null) {
    return { active: false, expired: false, untilDate: "" };
  }
  const parsed = Date.parse(memberUntil);
  if (Number.isNaN(parsed)) {
    return { active: false, expired: true, untilDate: "" };
  }
  return {
    active: parsed > Date.now(),
    expired: parsed <= Date.now(),
    untilDate: localDateOf(parsed),
  };
}

// The wall's state, derived: the standing window decides the face, the
// pay_gate fold decides the two refusals. The wrapper prompts arrive
// flattened (a plain string or absent); typeof narrows what the
// generated types still claim as Message wrappers.
function wallStateOf(sdata: SessionView | undefined, payGate: PayGateView | undefined): WallState {
  const authed = sdata?.status === "authed";
  const birthYear = authed ? sdata?.profile?.birthYear : undefined;
  const birthMonth = authed ? sdata?.profile?.birthMonth : undefined;
  const window = windowOf(authed ? sdata?.lastMemberUntil : null);
  const birthMissing = authed && (birthYear === undefined || birthMonth === undefined);
  const ageLine = payGate?.payAgeLine ?? 18;
  const minor =
    !birthMissing &&
    birthYear !== undefined &&
    birthMonth !== undefined &&
    !ageLineReached(birthYear, birthMonth, ageLine);
  return {
    birthMissing,
    minor,
    memberActive: window.active,
    memberExpired: window.expired,
    untilDate: window.untilDate,
    ageLine,
    ageNow: birthYear !== undefined && birthMonth !== undefined ? ageOf(birthYear, birthMonth) : 0,
    minorPrompt: typeof payGate?.minorPrompt === "string" ? payGate.minorPrompt : undefined,
    missingBirthPrompt:
      typeof payGate?.missingBirthPrompt === "string" ? payGate.missingBirthPrompt : undefined,
  };
}

// Where a created order continues: the mock provider answers through the
// simulated cashier sheet; a real provider continues by the credential
// kind — jump away, invoke inside WeChat, or show the QR sheet.
function continueWithCredential(
  res: CreateResult,
  kind: PurchaseKind,
  face: MemberFace,
  price: string,
  onMockSheet: (state: SheetState) => void,
  onInvokePaid: (orderNo: string) => void,
  onQr: (state: QrState) => void,
): void {
  const orderNo = res.order?.orderNo ?? "";
  if ((res.order?.provider ?? "") === "mock") {
    onMockSheet({ face, kind, orderNo, price });
    return;
  }
  const credentialKind = String(res.credential?.kind ?? "");
  const value = res.credential?.value ?? "";
  if (credentialKind === CREDENTIAL_JUMP_URL && value !== "") {
    window.location.href = value;
    return;
  }
  if (credentialKind === CREDENTIAL_INVOKE && value !== "") {
    invokeWechatPay(value, (ok) => {
      if (ok) {
        onInvokePaid(orderNo);
      }
    });
    return;
  }
  if (credentialKind === CREDENTIAL_QR_CODE && value !== "") {
    onQr({ face, kind, orderNo, value });
  }
}

// The create input this wall builds, structurally: the catalog code, the
// environment-decided channel, and the two channel-dependent hints.
interface CreateInput {
  skuCode: string;
  channel: string;
  scopeRef?: string;
  returnUrl?: string;
}

// The H5 leg's post-payment landing: the absolute receipt URL carrying
// the face the purchase started under (and the wall's own from
// passthrough). The order number cannot be known here — it is minted by
// the same create call this rides into — so the server stitches it onto
// this URL before the channel sees it, and the cashier's redirect lands
// the payer on a receipt reading its own row, never back on a wall that
// lost the order context.
function h5ReturnUrl(face: MemberFace, from: string | undefined): string {
  const params = new URLSearchParams({ face });
  if (from !== undefined) {
    params.set("from", from);
  }
  return `${window.location.origin}/pay/result?${params.toString()}`;
}

// Create and continue: one purchase attempt end to end. The refusal
// toasts ride the mutation cache; nothing here swallows a state change.
async function createAndContinue(
  kind: PurchaseKind,
  face: PurchaseFace,
  memberFace: MemberFace,
  search: PaySearch,
  mutate: (input: CreateInput) => Promise<CreateResult>,
  onMockSheet: (state: SheetState) => void,
  onInvokePaid: (orderNo: string) => void,
  onQr: (state: QrState) => void,
): Promise<void> {
  if (face.row === undefined) {
    return;
  }
  const channel = channelOf(navigator.userAgent);
  try {
    const res = await mutate({
      skuCode: face.row.skuCode ?? "",
      channel,
      ...(kind === "trial" && search.ref !== undefined ? { scopeRef: search.ref } : {}),
      ...(channel === "wechat_h5" ? { returnUrl: h5ReturnUrl(memberFace, search.from) } : {}),
    });
    continueWithCredential(res, kind, memberFace, face.yuan, onMockSheet, onInvokePaid, onQr);
  } catch {
    // The mutation cache already toasts the reason copy.
  }
}

function sheetActionsOf(
  settle: () => void,
  onFail: () => void,
  onCancel: () => void,
): SheetAction[] {
  return [
    { label: m.pay_sheet_success(), kind: "primary", onPress: settle },
    { label: m.pay_sheet_fail(), kind: "ghost", onPress: onFail },
    { label: m.pay_sheet_cancel(), kind: "ghost", onPress: onCancel },
  ];
}

function qrActionsOf(onCheck: () => void, onCancel: () => void): SheetAction[] {
  return [
    { label: m.pay_qr_check(), kind: "primary", onPress: onCheck },
    { label: m.pay_sheet_cancel(), kind: "ghost", onPress: onCancel },
  ];
}

/** 态②, the slice's own subject: renewal only — an active window hides
 * the single unlock entirely. Below the green button the P7 draft's text
 * link hands off to the benefits page. */
function RenewalFace({
  disabled,
  onDetail,
  onPay,
  price,
  untilDate,
}: Readonly<{
  disabled: boolean;
  onDetail: () => void;
  onPay: () => void;
  price: string;
  untilDate: string;
}>) {
  return (
    <>
      <StateBanner tag={m.pay_member_tag()} text={m.pay_member_until({ date: untilDate })} />
      <div className="rounded-card shadow-card border border-[rgba(168,85,247,0.55)] bg-[linear-gradient(135deg,rgba(168,85,247,0.1),rgba(17,24,39,0.92))] p-4 backdrop-blur-md">
        <div className="flex items-center justify-between">
          <span className="text-ink text-[15px] font-extrabold">{m.pay_opt_member_name()}</span>
          <span className="text-grade-s text-[22px] font-black tabular-nums">¥{price}</span>
        </div>
        <p className="text-ink-muted mt-2 text-[11.5px] leading-[1.7]">{m.pay_renew_desc()}</p>
      </div>
      <PayWayCard title={m.pay_pay_way()} way={m.pay_pay_wechat()} />
      <Button
        className="bg-cta-success shadow-cta-success w-full rounded-[12px] py-3.5 text-[15px] font-bold text-white active:opacity-80 disabled:opacity-50"
        isDisabled={disabled}
        onPress={onPay}
      >
        {m.pay_act_renew({ price })}
      </Button>
      <button
        className="tap-zone text-accent-info flex w-full items-center justify-center gap-0.5 text-[11px] active:opacity-60"
        onClick={onDetail}
        type="button"
      >
        {m.pay_benefit_link()}
        <ChevronRight aria-hidden size={11} />
      </button>
    </>
  );
}

/** 态①/③: the dual options, member first — an expired window says
 * so on the neutral banner before the pitch. The single unlock targets
 * one test record, so it renders only when a record is in scope (the
 * report context entries); a ref-less arrival (the profile card, the
 * member retry) pitches member only — the same member-only face the
 * prototype's profile entry carries. */
function DualFace({
  expiredText,
  memberDisabled,
  memberPrice,
  onPay,
  showTrial,
  trialDisabled,
  trialPrice,
}: Readonly<{
  expiredText?: string;
  memberDisabled: boolean;
  memberPrice: string;
  onPay: (kind: PurchaseKind) => void;
  showTrial: boolean;
  trialDisabled: boolean;
  trialPrice: string;
}>) {
  return (
    <>
      {expiredText !== undefined ? <StateBanner expired text={expiredText} /> : null}
      <div className="border-line bg-surface-card shadow-card rounded-card border p-4 backdrop-blur-md">
        <div className="text-ink mb-3 text-[13px] font-bold">{m.pay_choose_way()}</div>
        <div className="flex flex-col gap-2.5">
          {showTrial ? (
            <OptionCard
              desc={m.pay_opt_trial_desc()}
              name={m.pay_opt_trial_name()}
              price={`¥${trialPrice}`}
            />
          ) : null}
          <OptionCard
            bestBadge={m.pay_opt_member_best()}
            desc={m.pay_opt_member_desc()}
            highlighted
            name={m.pay_opt_member_name()}
            price={`¥${memberPrice}`}
          />
        </div>
      </div>
      <Button
        className="bg-cta-primary shadow-cta-primary w-full rounded-[12px] py-3.5 text-[15px] font-bold text-white active:opacity-80 disabled:opacity-50"
        isDisabled={memberDisabled}
        onPress={() => onPay("member")}
      >
        {m.pay_act_member({ price: memberPrice })}
      </Button>
      {showTrial ? (
        <Button
          className="bg-surface-input text-ink-secondary border-ambient-slate w-full rounded-[12px] border-[1.5px] py-3.5 text-[15px] font-medium active:opacity-80 disabled:opacity-50"
          isDisabled={trialDisabled}
          onPress={() => onPay("trial")}
        >
          {m.pay_act_trial({ price: trialPrice })}
        </Button>
      ) : null}
    </>
  );
}

// The wall's interaction flow, one hook: create-and-continue, the mock
// cashier sheet over the dev bridge, the QR sheet, and the settle poll
// that hands the outcome to the receipt page — the prototype routes every
// cashier exit there, the wall never renders the verdict itself.
function usePayFlow(faces: ReturnType<typeof purchaseFaces>, search: PaySearch) {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [sheet, setSheet] = useState<SheetState | null>(null);
  const [qr, setQr] = useState<QrState | null>(null);
  const [polling, setPolling] = useState<{
    kind: PurchaseKind;
    orderNo: string;
    face: MemberFace;
  } | null>(null);
  const createOrderMut = useMutation(createOrder);
  // The simulated cashier's own face — a mock provider's dev RPC, one the
  // service refuses (CASHIER_DISABLED) wherever the default provider is
  // not mock, so the sheet's success leg exists only in dev.
  const markPaidMut = useMutation(markPaid);
  const closeOrderMut = useMutation(closeOrder);
  // Convergence read: the bridge settles synchronously server-side, so the
  // first poll usually answers PAID; the interval only covers the race.
  const orderQuery = useConnectQuery(
    getOrder,
    { orderNo: polling?.orderNo ?? "" },
    { enabled: polling !== null, refetchInterval: 700 },
  );

  // The settle finisher: PAID observed → the session's embedded window
  // refreshes, then the receipt page owns the outcome (a CLOSED row —
  // closed by this page's cancel or expired server-side — renders the
  // cancel face from its own truth). The face each state carries is the
  // one captured at click time — the receipt quotes the window the
  // purchase started under, never a mid-poll refetch's rewrite. The state
  // rides the wire as a value-name string — String() restates what the
  // transport delivers.
  useEffect(() => {
    if (polling === null) {
      return;
    }
    const state = String(orderQuery.data?.order?.state ?? "");
    if (state !== ORDER_STATE_PAID && state !== ORDER_STATE_CLOSED) {
      return;
    }
    setPolling(null);
    if (state === ORDER_STATE_PAID) {
      void queryClient.invalidateQueries({ queryKey: sessionQuery.queryKey });
    }
    void navigate({
      to: "/pay/result",
      search: {
        order: polling.orderNo,
        face: polling.face,
        ...(search.from !== undefined ? { from: search.from } : {}),
      },
    });
  }, [orderQuery.data, polling, queryClient, navigate, search.from]);

  function toResult(orderNo: string, face: MemberFace) {
    void navigate({
      to: "/pay/result",
      search: { order: orderNo, face, ...(search.from !== undefined ? { from: search.from } : {}) },
    });
  }

  // The cancel leg: close the order first so the channel order cannot be
  // paid into behind the receipt, then land on it — the row answers the
  // face. A close refusal toasts through the mutation cache and the order
  // still expires on its own.
  function abandonTo(orderNo: string, face: MemberFace) {
    void (async () => {
      try {
        await closeOrderMut.mutateAsync({ orderNo });
      } catch {
        // The receipt page reads the row's own truth either way.
      }
      toResult(orderNo, face);
    })();
  }

  async function pay(kind: PurchaseKind, face: MemberFace) {
    if (createOrderMut.isPending) {
      return;
    }
    await createAndContinue(
      kind,
      kind === "member" ? faces.member : faces.trial,
      face,
      search,
      createOrderMut.mutateAsync,
      setSheet,
      (orderNo) => setPolling({ kind, orderNo, face }),
      setQr,
    );
  }

  function settleViaBridge() {
    if (sheet === null) {
      return;
    }
    const target = sheet;
    setSheet(null);
    void (async () => {
      try {
        await markPaidMut.mutateAsync({ orderNo: target.orderNo });
        setPolling({ kind: target.kind, orderNo: target.orderNo, face: target.face });
      } catch {
        // The mutation cache already toasts the reason copy; the sheet
        // returns so its button can replay the idempotent call.
        setSheet(target);
      }
    })();
  }

  const closeSheet = () => setSheet(null);
  const closeQr = () => setQr(null);

  const sheetActions =
    sheet !== null
      ? sheetActionsOf(
          settleViaBridge,
          () => {
            const target = sheet;
            setSheet(null);
            toResult(target.orderNo, target.face);
          },
          () => {
            const target = sheet;
            setSheet(null);
            abandonTo(target.orderNo, target.face);
          },
        )
      : [];

  const qrActions =
    qr !== null
      ? qrActionsOf(
          () => {
            const target = qr;
            setQr(null);
            setPolling({ kind: target.kind, orderNo: target.orderNo, face: target.face });
          },
          () => {
            const target = qr;
            setQr(null);
            abandonTo(target.orderNo, target.face);
          },
        )
      : [];

  return {
    sheet,
    qr,
    sheetActions,
    qrActions,
    createOrderMut,
    pay,
    closeSheet,
    closeQr,
  };
}

// The wall's own body, one function per refusal/face: the early returns
// ARE the state machine — backfill, refusal, renewal, dual options.
function WallBody({
  faces,
  onBackfill,
  onBackOrHome,
  onDetail,
  onHome,
  onPay,
  pending,
  showTrial,
  wall,
}: Readonly<{
  faces: ReturnType<typeof purchaseFaces>;
  onBackfill: () => void;
  onBackOrHome: () => void;
  onDetail: () => void;
  onHome: () => void;
  onPay: (kind: PurchaseKind) => void;
  pending: boolean;
  showTrial: boolean;
  wall: WallState;
}>) {
  if (wall.birthMissing) {
    return (
      <BirthCard
        actionLabel={m.pay_birth_act()}
        copy={wall.missingBirthPrompt ?? m.pay_birth_default()}
        onBackfill={onBackfill}
      />
    );
  }
  if (wall.minor) {
    return (
      <MinorCard
        copy={
          wall.minorPrompt ??
          m.pay_minor_default({ age: String(wall.ageNow), line: String(wall.ageLine) })
        }
        freeLabel={m.pay_minor_free()}
        homeLabel={m.pay_minor_home()}
        onFree={onBackOrHome}
        onHome={onHome}
      />
    );
  }
  if (wall.memberActive) {
    return (
      <RenewalFace
        disabled={faces.member.row === undefined || pending}
        onDetail={onDetail}
        onPay={() => onPay("member")}
        price={faces.member.yuan}
        untilDate={wall.untilDate}
      />
    );
  }
  return (
    <DualFace
      expiredText={wall.memberExpired ? m.pay_member_expired({ date: wall.untilDate }) : undefined}
      memberDisabled={faces.member.row === undefined || pending}
      memberPrice={faces.member.yuan}
      onPay={onPay}
      showTrial={showTrial}
      trialDisabled={faces.trial.row === undefined || pending}
      trialPrice={faces.trial.yuan}
    />
  );
}

/**
 * The pay wall (three states): a
 * non-member and an expired member both face the dual options, an active
 * member faces renewal only. The server's CreateOrder refusal stays the
 * authority — this page's birth and age pre-checks (the pay_gate fold off
 * the catalog read) only keep the wall from pitching a purchase the gate
 * refuses: a missing birth renders the backfill card, a minor the refusal
 * card with its two exits, and no pay button renders under either.
 */
export function PayConfirmPage() {
  const search = useSearch({ from: "/pay/" });
  const navigate = useNavigate();
  const router = useRouter();
  const session = useQuery({ ...sessionQuery, throwOnError: false });
  const skusQuery = useConnectQuery(listSkus, {});

  const wall = wallStateOf(session.data, skusQuery.data?.payGate);
  const faces = purchaseFaces(skusQuery.data?.skus ?? []);
  const flow = usePayFlow(faces, search);
  // The receipt's member copy fork (opened vs renewed) keys off the
  // standing window at click time — captured in the click handler, so a
  // session refetch mid-poll can never rewrite the verdict.
  const payWith = (kind: PurchaseKind) => void flow.pay(kind, wall.memberActive ? "renew" : "open");

  function backOrHome() {
    if (router.history.canGoBack()) {
      router.history.back();
    } else {
      void navigate({ to: "/" });
    }
  }

  return (
    <AmbientLayout orbs={wall.memberActive ? RENEW_ORBS : undefined}>
      <div className="flex w-full flex-col gap-3.5 pb-6">
        {/* Inline header — the agreement page's local-declaration pattern
         * (match's FlowHeader is another module's territory); same shape:
         * the 40px back fab and the state's own title. */}
        <div className="flex items-center gap-3 pt-5">
          <button
            aria-label={m.common_back()}
            className="bg-surface-input text-ink-secondary flex h-10 w-10 shrink-0 items-center justify-center rounded-full active:opacity-60"
            onClick={backOrHome}
            type="button"
          >
            <ChevronLeft aria-hidden size={20} />
          </button>
          <h1 className="text-ink text-[19px] font-bold">
            {wall.memberActive ? m.pay_head_renew() : m.pay_head_unlock()}
          </h1>
        </div>

        <WallBody
          faces={faces}
          onBackfill={() => void navigate({ to: "/age-required" })}
          onBackOrHome={backOrHome}
          onDetail={() => void navigate({ to: "/member" })}
          onHome={() => void navigate({ to: "/" })}
          onPay={payWith}
          pending={flow.createOrderMut.isPending}
          showTrial={search.ref !== undefined}
          wall={wall}
        />

        <PayFilingFooter />
      </div>

      {flow.sheet !== null ? (
        <BottomConfirmSheet
          actions={flow.sheetActions}
          body={m.pay_sheet_body()}
          onDismiss={flow.closeSheet}
          title={m.pay_sheet_title({ price: flow.sheet.price })}
        />
      ) : null}

      {flow.qr !== null ? (
        <BottomConfirmSheet
          actions={flow.qrActions}
          body={m.pay_qr_body()}
          onDismiss={flow.closeQr}
          title={m.pay_qr_title()}
          visual={<QrTile value={flow.qr.value} />}
        />
      ) : null}
    </AmbientLayout>
  );
}
