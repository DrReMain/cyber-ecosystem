import { System } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/error_reason_pb";
import { useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { ArrowRight } from "lucide-react";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import { systemReason } from "#/domains/error";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { BottomConfirmSheet } from "#/features/app/dialog/bottom-confirm-sheet";
import { inputClass } from "#/features/modules/login/ui/sms-fields";
import { beginWechatLoginFn } from "#/features/modules/login/wechat.fn";
import { FlowHeader } from "#/features/modules/match/ui/flow-header";
import { m } from "#/paraglide/messages";
import { clearMergeFlow, readMergeFlow } from "./merge-flow";
import type { MergeOfferDTO, MergeSideDTO } from "./security.fn";
import { beginBindWechatFn, mergeSmsFn } from "./security.fn";
import { useBindSms } from "./use-bind-sms";

const CARD_CLASS = "border-line-accent-blue bg-identity rounded-card border";
const ORANGE_BTN =
  "rounded-control bg-[linear-gradient(to_right,#EA580C,#F59E0B)] px-4 py-3.5 text-[15px] font-bold text-white shadow-[0_10px_24px_rgba(234,88,12,0.35)] disabled:opacity-50";

type Side = MergeSideDTO | undefined;

// The scenario triple the page renders from: which side is the empty one
// (the corpus's fixed direction, never client-chosen) and which proof leg
// executes it — sms when the bind-phone flow carried the number, wechat
// otherwise, the identity switch when the caller is the empty side.
type Scenario =
  | { kind: "keepCallerSms"; poor: Side; rich: Side }
  | { kind: "keepCallerWechat"; poor: Side; rich: Side }
  | { kind: "keepOther"; poor: Side; rich: Side }
  | { kind: "blocked"; poor: Side; rich: Side };

function scenarioOf(offer: MergeOfferDTO, phone: string | undefined): Scenario {
  const flip = offer.status === "KEEP_OTHER";
  const poor = flip ? offer.caller : offer.other;
  const rich = flip ? offer.other : offer.caller;
  if (offer.status === "KEEP_OTHER") {
    return { kind: "keepOther", poor, rich };
  }
  if (offer.status === "KEEP_CALLER") {
    return { kind: phone !== undefined ? "keepCallerSms" : "keepCallerWechat", poor, rich };
  }
  return { kind: "blocked", poor, rich };
}

// The rejection story per reason — the merge's own trio plus the fallback.
function mergeFailureCopy(reason: string | undefined): string {
  switch (systemReason(reason)) {
    case System.SMS_CODE_INVALID:
      return m.error_SYSTEM_SMS_CODE_INVALID();
    case System.MERGE_NOT_MERGEABLE:
      return m.error_SYSTEM_MERGE_NOT_MERGEABLE();
    case System.MERGE_NOT_BOUND:
      return m.error_SYSTEM_MERGE_NOT_BOUND();
    default:
      return m.security_error_generic();
  }
}

function summaryLine(s: Side | null | undefined): string {
  if (s === null || s === undefined) {
    return "";
  }
  if (s.sessions === 0 && s.memberUntil === undefined && !s.unlocked) {
    return m.security_merge_summary_empty();
  }
  const parts: string[] = [m.security_merge_summary_sessions({ count: s.sessions })];
  if (s.memberUntil !== undefined) {
    const until = new Date(s.memberUntil).toISOString().slice(0, 10);
    parts.push(m.security_merge_summary_member({ until }));
  } else if (s.unlocked) {
    parts.push(m.security_merge_summary_unlocked());
  }
  return parts.join(" · ");
}

// The middle line is the bare mask, exactly as the prototype renders it —
// a "手机号" prefix wraps inside the narrow half-card.
function summaryLabel(s: Side | null | undefined): string {
  if (s === null || s === undefined) {
    return "";
  }
  if (s.phoneMask !== undefined && s.phoneMask !== "") {
    return s.phoneMask;
  }
  return s.wechat ? m.security_merge_label_wechat() : m.security_merge_label_bare();
}

function mainlandOk(phone: string | undefined): boolean {
  return phone !== undefined && /^1[3-9]\d{9}$/.test(phone);
}

// The merge-confirm face: the two-account card pair (poor →
// rich), the proof line, the amber irreversibility warning, and the
// scenario's own execution leg. BLOCKED renders the no-auto-merge pane.
export function MergeConfirmPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const flow = readMergeFlow();
  const offer = flow.offer;
  const [seedPhone] = useState(flow.phone ?? "");
  const bind = useBindSms(seedPhone);
  const [sheetOpen, setSheetOpen] = useState(false);
  const [merging, setMerging] = useState(false);
  const [wechatPending, setWechatPending] = useState(false);

  // An empty handoff honestly bounces — the offer is a one-shot read and
  // cannot be re-derived from the URL.
  useEffect(() => {
    if (offer === undefined) {
      void navigate({ to: "/profile/security", viewTransition: false });
    }
  }, [offer, navigate]);

  if (offer === undefined) {
    return (
      <main className="safe-px safe-pt min-h-app flex items-center justify-center">
        <p className="text-ink-tertiary">{m.security_merge_pending()}</p>
      </main>
    );
  }

  const scenario = scenarioOf(offer, flow.phone);
  if (scenario.kind === "blocked") {
    return <BlockedPane onBack={() => void navigate({ to: "/profile/security" })} />;
  }

  async function executeSmsMerge() {
    if (flow.phone === undefined) {
      return;
    }
    setMerging(true);
    const res = await mergeSmsFn({
      data: { code: bind.code, phone: flow.phone, state: bind.state },
    });
    setMerging(false);
    setSheetOpen(false);
    if (res.status === "rejected") {
      toast.error(mergeFailureCopy(res.reason));
      return;
    }
    await queryClient.invalidateQueries();
    await navigate({ to: "/profile/security/merge/success", viewTransition: false });
    // Cleared only after the navigation commits — while mounted, an emptied
    // hand would flash the fallback and race its bounce against this target.
    clearMergeFlow();
  }

  async function launchWechatMergeProof() {
    setWechatPending(true);
    const res = await beginBindWechatFn({
      data: { redirectUri: `${window.location.origin}/profile/security/merge/bind-callback` },
    });
    if (res.status === "rejected") {
      setWechatPending(false);
      toast.error(m.security_bind_wechat_begin_failed());
      return;
    }
    window.location.href = res.authorizeUrl;
  }

  async function launchWechatSwitch() {
    setWechatPending(true);
    try {
      // beginWechatLoginFn throws on failure (no verdict wrapper) — the
      // stock login leg's own shape.
      const res = await beginWechatLoginFn({
        data: { redirectUri: `${window.location.origin}/profile/security/merge/wechat-callback` },
      });
      window.location.href = res.authorizeUrl;
    } catch {
      setWechatPending(false);
      toast.error(m.security_bind_wechat_begin_failed());
    }
  }

  return (
    <AmbientLayout>
      <FlowHeader
        onBack={() => void navigate({ to: "/profile/security" })}
        title={m.security_merge_title()}
      />
      <div className="flex w-full flex-col gap-3 pt-3">
        <AccountPair poor={scenario.poor} rich={scenario.rich} />

        <div className="text-ink-secondary border-line-accent-blue bg-identity rounded-card border px-3.5 py-3 text-[12px] leading-[1.85]">
          <b className="text-ink">{m.security_merge_proof_head()}</b>
          <br />
          {proofCopyOf(scenario.kind)}
        </div>

        <div className="rounded-card border border-[rgba(245,158,11,0.45)] bg-[rgba(120,53,15,0.18)] px-3.5 py-3 text-[12px] leading-[1.85] text-[#FCD34D]">
          {m.security_merge_warning()}
        </div>

        {scenario.kind === "keepCallerSms" ? <SmsProofFields bind={bind} /> : null}

        <ScenarioAction
          kind={scenario.kind}
          merging={merging}
          onOpenSheet={() => setSheetOpen(true)}
          onWechatMerge={() => void launchWechatMergeProof()}
          onWechatSwitch={() => void launchWechatSwitch()}
          smsReady={bind.code.length === 6 && mainlandOk(flow.phone)}
          wechatPending={wechatPending}
        />

        <button
          className="border-line text-ink-secondary bg-surface-input rounded-control border-[1.5px] px-4 py-3.5 text-[15px] font-medium active:opacity-60"
          onClick={() => {
            void navigate({ to: "/profile/security", viewTransition: false }).then(() => {
              clearMergeFlow();
            });
          }}
          type="button"
        >
          {m.security_merge_later()}
        </button>
      </div>

      {/* The prototype's bottom-sheet double gate (CTS.sheet) — the code is
          typed, the sheet is the last deliberate yes on an irreversible
          move. */}
      {sheetOpen ? (
        <BottomConfirmSheet
          actions={[
            {
              label: merging ? m.security_pending() : m.security_merge_sheet_confirm(),
              kind: "hot",
              onPress: () => void executeSmsMerge(),
            },
            {
              label: m.security_merge_sheet_cancel(),
              kind: "ghost",
              onPress: () => setSheetOpen(false),
            },
          ]}
          body={m.security_merge_sheet_body()}
          onDismiss={() => setSheetOpen(false)}
          title={m.security_merge_sheet_title()}
        />
      ) : null}
    </AmbientLayout>
  );
}

function proofCopyOf(kind: Scenario["kind"]): string {
  if (kind === "keepOther") {
    return m.security_merge_proof_switch();
  }
  return kind === "keepCallerSms" ? m.security_merge_proof_sms() : m.security_merge_proof_wechat();
}

// The two-account card pair — direction is data, never client-chosen.
function AccountPair({ poor, rich }: { poor: Side; rich: Side }) {
  return (
    <div className={`${CARD_CLASS} px-4 py-4`}>
      <div className="text-ink mb-2.5 text-sm font-extrabold">
        {m.security_merge_confirm_head()}
      </div>
      <div className="flex items-stretch gap-2.5">
        <div className="border-line bg-surface-input rounded-card flex-1 border border-dashed px-3 py-3 text-center">
          <div className="text-ink-tertiary text-[10.5px]">{m.security_merge_poor_tag()}</div>
          <div className="text-ink mt-1 text-[13.5px] font-extrabold">{summaryLabel(poor)}</div>
          <div className="text-ink-tertiary mt-1 text-[10.5px]">{summaryLine(poor)}</div>
        </div>
        <ArrowRight aria-hidden className="text-ink-tertiary self-center" size={16} />
        <div className="border-focus/60 bg-surface-input rounded-card flex-1 border px-3 py-3 text-center">
          <div className="text-ink-tertiary text-[10.5px]">{m.security_merge_rich_tag()}</div>
          <div className="text-ink mt-1 text-[13.5px] font-extrabold">{summaryLabel(rich)}</div>
          <div className="text-ink-tertiary mt-1 text-[10.5px]">{summaryLine(rich)}</div>
        </div>
      </div>
    </div>
  );
}

const sendBtnClass =
  "border-focus text-link bg-transparent rounded-control w-[112px] shrink-0 border px-3 py-[13px] text-[13px] font-bold active:opacity-60 disabled:opacity-60";

function SmsProofFields({ bind }: { bind: ReturnType<typeof useBindSms> }) {
  return (
    <>
      <input
        aria-label={m.security_phone_label()}
        className={inputClass}
        inputMode="numeric"
        maxLength={11}
        onChange={(event) => bind.setPhone(event.target.value)}
        placeholder={m.security_bind_phone_placeholder()}
        value={bind.phone}
      />
      <div className="flex items-stretch gap-3">
        <input
          aria-label={m.security_code_label()}
          className={`${inputClass} flex-1 tracking-[4px]`}
          inputMode="numeric"
          maxLength={6}
          onChange={(event) => bind.setCode(event.target.value)}
          placeholder={m.security_bind_phone_code_placeholder()}
          value={bind.code}
        />
        <button
          className={sendBtnClass}
          disabled={bind.countdown > 0 || bind.sending}
          onClick={() => void bind.send()}
          type="button"
        >
          {bind.countdown > 0
            ? m.security_code_resend_in({ seconds: bind.countdown })
            : m.security_code_send()}
        </button>
      </div>
    </>
  );
}

function ScenarioAction(props: {
  kind: Scenario["kind"];
  merging: boolean;
  onOpenSheet: () => void;
  onWechatMerge: () => void;
  onWechatSwitch: () => void;
  smsReady: boolean;
  wechatPending: boolean;
}) {
  const { kind, merging, onOpenSheet, onWechatMerge, onWechatSwitch, smsReady, wechatPending } =
    props;
  if (kind === "keepOther") {
    return (
      <button
        className="bg-cta-primary shadow-cta-primary rounded-control px-4 py-3.5 text-[15px] font-bold text-white disabled:opacity-50"
        disabled={wechatPending}
        onClick={onWechatSwitch}
        type="button"
      >
        {wechatPending ? m.security_pending() : m.security_merge_switch_cta()}
      </button>
    );
  }
  if (kind === "keepCallerSms") {
    return (
      <button
        className={ORANGE_BTN}
        disabled={merging || !smsReady}
        onClick={onOpenSheet}
        type="button"
      >
        {merging ? m.security_pending() : m.security_merge_confirm_cta()}
      </button>
    );
  }
  return (
    <button className={ORANGE_BTN} disabled={wechatPending} onClick={onWechatMerge} type="button">
      {wechatPending ? m.security_pending() : m.security_merge_wechat_cta()}
    </button>
  );
}

// The blocked pane: both sides carry weight, the merge stays
// closed, and the manual path is the only offer.
function BlockedPane({ onBack }: { onBack: () => void }) {
  return (
    <AmbientLayout>
      <FlowHeader onBack={onBack} title={m.security_merge_title()} />
      <div className="flex w-full flex-col gap-3 pt-3">
        <div className={`${CARD_CLASS} px-4 py-4`}>
          <div className="text-danger-text mb-2 text-sm font-extrabold">
            {m.security_merge_blocked_head()}
          </div>
          <p className="text-ink-secondary text-[12.5px] leading-[1.85]">
            {m.security_merge_blocked_body()}
          </p>
          <p className="text-ink-tertiary mt-2.5 text-[11.5px] leading-[1.8]">
            {m.security_merge_blocked_manual()}
          </p>
        </div>
        <div className="text-ink-tertiary border-line-accent-blue bg-identity rounded-card border px-3.5 py-3 text-[11.5px] leading-[1.8]">
          <b className="text-ink-secondary">{m.security_merge_blocked_basis_head()}</b>
          <br />
          {m.security_merge_blocked_basis()}
        </div>
        <button
          className="border-line text-ink-secondary bg-surface-input rounded-control border-[1.5px] px-4 py-3.5 text-[15px] font-medium active:opacity-60"
          onClick={() => toast.info(m.common_soon())}
          type="button"
        >
          {m.security_merge_blocked_feedback_cta()}
        </button>
      </div>
    </AmbientLayout>
  );
}
