import { useQuery as useConnectQuery, useQuery } from "@connectrpc/connect-query";
import { getProfile } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/profile-ProfileService_connectquery";
import {
  getAbilityDigest,
  getSessionReport,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/report-ReportService_connectquery";
import { listSkus } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/sku-SkuService_connectquery";
import { useQueryClient } from "@tanstack/react-query";
import { Link, useNavigate } from "@tanstack/react-router";
import { TrendingUp } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import type { AmbientOrb } from "#/features/app/ambient/ambient-backdrop";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { StateScreen } from "#/features/app/states/state-screen";
import { TabBar } from "#/features/app/tabbar";
import { clearWechatProfile } from "#/features/app/wechat-profile";
import { PROFILE_STALE_MS } from "#/features/auth/profile-cache";
import { broadcastSessionKick } from "#/features/auth/session-kick";
import { latestScoredOf, wireDigestOf } from "#/features/modules/report/digest";
import { MemberCard } from "#/features/modules/report/ui/member-card";
import { priceYuanOf } from "#/features/modules/report/vocabulary";
import { m } from "#/paraglide/messages";
import { logoutFn } from "./logout.fn";
import { LogoutSheet } from "./logout-sheet";
import { bestScoreOf, compTopOf, historySlotsOf, overviewRowsOf } from "./overview";
import { AbilityOverviewCard } from "./ui/ability-overview-card";
import { DataCard } from "./ui/data-card";
import { GenreStatsCard } from "./ui/genre-stats-card";
import { HistoryCard } from "./ui/history-card";
import { IdentityCard } from "./ui/identity-card";
import { SecurityCard } from "./ui/security-card";

/* The overview page's orb pair (prototype profile-overview). */
const overviewOrbs: readonly AmbientOrb[] = [
  { className: "top-[-70px] left-[-60px] size-[240px]", tone: "purple" },
  { className: "bottom-[-50px] right-[-60px] size-[200px]", tone: "cyan" },
];

export function ProfilePage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  // Unauthenticated never mounts this page: the /profile layout's session
  // guard (features/auth/guard.ts) redirects in beforeLoad, so this query
  // only ever fires authed. Errors escalate to the route boundary per the
  // global throwOnError — this page owns only the pending face.
  const profileQuery = useQuery(getProfile, {}, { staleTime: PROFILE_STALE_MS });
  // The overview line's reads (prototype ①④⑤⑥): the digest carries the
  // counts, the ability bests, and the latest session lines; the latest
  // scored session's own report answers the composite stat. Failed reads
  // keep the cards absent rather than taking the page down.
  const digestQuery = useConnectQuery(getAbilityDigest, {}, { throwOnError: false });
  const digest =
    digestQuery.isSuccess && digestQuery.data !== undefined ? wireDigestOf(digestQuery.data) : null;
  const latestScored = digest !== null ? latestScoredOf(digest) : undefined;
  const compositeQuery = useConnectQuery(
    getSessionReport,
    { sessionId: latestScored?.sessionId ?? "" },
    { enabled: latestScored !== undefined, throwOnError: false },
  );
  const compTop =
    compositeQuery.isSuccess && compositeQuery.data !== undefined
      ? compTopOf(compositeQuery.data.compositePercentile as number | undefined)
      : null;
  const skusQuery = useConnectQuery(listSkus, {}, { throwOnError: false });
  const memberYuan = priceYuanOf(skusQuery.data?.skus ?? [], "SKU_KIND_MEMBERSHIP", true);
  const [loggingOut, setLoggingOut] = useState(false);
  const [sheetOpen, setSheetOpen] = useState(false);

  if (profileQuery.isPending) {
    return (
      <AmbientLayout chrome={<TabBar active="me" />} orbs={overviewOrbs}>
        <StateScreen variant="loading" />
      </AmbientLayout>
    );
  }

  const profile = profileQuery.data?.profile ?? undefined;

  async function onLogout() {
    setLoggingOut(true);
    try {
      await logoutFn();
      queryClient.clear();
      // The parked wechat identity belonged to the seat that just left —
      // a later sign-in in this tab must not inherit it as a prefill.
      clearWechatProfile();
      // Boots the other tabs off the dead session (no expired notice — a
      // deliberate sign-out is expected, not a surprise).
      broadcastSessionKick();
      await navigate({ to: "/login" });
    } finally {
      setLoggingOut(false);
      setSheetOpen(false);
    }
  }

  return (
    <AmbientLayout chrome={<TabBar active="me" />} orbs={overviewOrbs}>
      <div className="flex w-full flex-col gap-3 pb-20 md:pb-0">
        <div className="flex items-center justify-between">
          <h1 className="text-ink text-[19px] font-bold">{m.profile_title()}</h1>
          <button
            className="tap-zone text-danger-text rounded-full border border-[rgba(239,68,68,0.6)] px-[11px] py-1 text-[11.5px] font-extrabold active:opacity-60 disabled:opacity-50"
            disabled={loggingOut}
            onClick={() => setSheetOpen(true)}
            type="button"
          >
            {m.profile_logout()}
          </button>
        </div>

        <IdentityCard
          profile={profile}
          stats={{
            total: digest?.sessionsTotal ?? null,
            compTop,
            best: digest !== null ? bestScoreOf(digest) : null,
          }}
        />
        <DataCard profile={profile} />
        <SecurityCard />

        {/* The membership face (prototype ③): price off the catalog, CTA
            parked until the pay client lands — the same stance the report
            page's purchase faces hold. */}
        <MemberCard priceYuan={memberYuan} unlocked={false} />

        {digest !== null ? <AbilityOverviewCard rows={overviewRowsOf(digest)} /> : null}
        {digest !== null ? <HistoryCard slots={historySlotsOf(digest)} /> : null}
        {digest !== null ? (
          <GenreStatsCard fps={digest.sessionsFps} moba={digest.sessionsMoba} />
        ) : null}

        {/* Training trend — the prototype's own Phase-2 placeholder, dashed. */}
        <div className="rounded-card border border-dashed border-[rgba(107,114,128,0.6)] bg-[rgba(17,23,41,0.5)] px-4 py-[15px]">
          <div className="mb-[6px] flex items-center justify-between">
            <span className="text-[14.5px] font-extrabold text-[#9ca3af]">
              {m.profile_training_title()}
            </span>
            <span className="rounded-full bg-[rgba(234,179,8,0.18)] px-2.5 py-[3px] text-[10px] font-semibold text-[#fde047]">
              {m.profile_training_soon()}
            </span>
          </div>
          <p className="text-ink-tertiary text-center text-[11.5px] leading-[1.7]">
            {m.profile_training_desc()}
          </p>
        </div>

        {/* The report-center row (prototype ⑧): the trend report it names
            is a later slice's face, so the entry parks on the honest soon
            toast rather than linking nowhere. */}
        <div className="border-line-accent-blue bg-identity rounded-card border px-4 py-[15px]">
          <div className="mb-[11px] flex items-center justify-between">
            <span className="text-ink text-[14.5px] font-extrabold">
              {m.profile_center_title()}
            </span>
          </div>
          <div className="flex items-center justify-between">
            <div className="flex min-w-0 items-center gap-2.5">
              <span
                aria-hidden
                className="flex h-[30px] w-[30px] shrink-0 items-center justify-center rounded-[9px] bg-[rgba(234,179,8,0.18)]"
              >
                <TrendingUp aria-hidden className="text-[#fde047]" size={14} />
              </span>
              <div className="min-w-0">
                <div className="text-ink text-[12.5px] font-extrabold">
                  {m.profile_center_trend()}
                </div>
                <div className="text-ink-tertiary mt-0.5 text-[10.5px]">
                  {m.profile_center_trend_desc()}
                </div>
              </div>
            </div>
            <button
              className="text-ink-secondary tap-zone rounded-control shrink-0 border-[1.5px] border-[#4b5563] px-3 py-1.5 text-[11.5px] font-medium active:opacity-60"
              onClick={() => toast.info(m.common_soon())}
              type="button"
            >
              {m.profile_center_view()}
            </button>
          </div>
        </div>

        <div className="border-line-accent-blue bg-identity rounded-card border px-4 py-[15px]">
          <div className="mb-[9px] flex items-center justify-between">
            <span className="text-ink text-[14.5px] font-extrabold">{m.profile_about()}</span>
            <span className="text-ink-tertiary text-xs">›</span>
          </div>
          <div className="text-ink-muted flex gap-[14px] text-xs">
            <Link className="tap-zone text-link py-1 active:opacity-60" to="/agreement/user">
              {m.profile_about_user()}
            </Link>
            <Link className="tap-zone text-link py-1 active:opacity-60" to="/agreement/privacy">
              {m.profile_about_privacy()}
            </Link>
            <button
              className="tap-zone py-1 active:opacity-60"
              onClick={() => toast.info(m.common_soon())}
              type="button"
            >
              {m.profile_about_icp()}
            </button>
          </div>
        </div>

        <div className="text-ink-tertiary mt-[14px] text-center text-[10.5px] leading-[1.7]">
          {m.common_footer_company()} · {m.common_footer_icp()}
          <br />
          {m.common_footer_email()}
        </div>
      </div>
      <LogoutSheet
        onCancel={() => setSheetOpen(false)}
        onConfirm={() => void onLogout()}
        open={sheetOpen}
      />
    </AmbientLayout>
  );
}
