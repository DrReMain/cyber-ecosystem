import { useQuery as useConnectQuery } from "@connectrpc/connect-query";
import {
  getAbilityDigest,
  getCognitiveReport,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/report-ReportService_connectquery";
import { useNavigate } from "@tanstack/react-router";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { StateScreen } from "#/features/app/states/state-screen";
import { FlowHeader } from "#/features/modules/match/ui/flow-header";
import { abilityWord } from "#/features/modules/match/vocabulary";
import { m } from "#/paraglide/messages";
import { wireCognitiveOf } from "./cognitive";
import { wireDigestOf } from "./digest";
import { AbilityRadar } from "./ui/ability-radar";
import { CognitiveHead } from "./ui/cognitive-head";
import { dimensionWord, tierWord } from "./vocabulary";

// The six-word core vocabulary in display order — the counters' and the
// missing-abilities line's own set; the digest says which are covered.
const CORE_ABILITIES = [
  "reactionSpeed",
  "aimingAccuracy",
  "trackingStability",
  "skillReleaseAccuracy",
  "resourceAllocation",
  "decisionEfficiency",
] as const;

// The detail row's bare-letter chip keeps the tier ladder's own colors
// (the head's full chip shares the palette).
const DIM_TIER_STYLE: Record<string, string> = {
  s: "bg-[rgba(234,179,8,0.15)] text-grade-s",
  a: "bg-status-ok-bg text-[#6EE7B7]",
  b: "bg-[rgba(59,130,246,0.15)] text-accent-info",
  c: "bg-[rgba(249,115,22,0.15)] text-[#FDBA74]",
};

// The eight-dimension aggregate's page. The verdict rides in the same
// response, so the locked face renders from one call; behind the open
// gate the page follows the prototype's own law — a dimension with no
// data is an absent row (never a zero axis), the radar needs three axes
// to be a shape at all, and both counters state the honest fractions.
export function CognitivePage() {
  const navigate = useNavigate();
  const cogQuery = useConnectQuery(getCognitiveReport, {}, { throwOnError: false });
  const digestQuery = useConnectQuery(getAbilityDigest, {}, { throwOnError: false });

  if (cogQuery.isPending) {
    return (
      <AmbientLayout>
        <StateScreen variant="loading" />
      </AmbientLayout>
    );
  }
  if (cogQuery.isError) {
    return (
      <AmbientLayout>
        <StateScreen
          action={
            <button
              className="bg-cta-primary rounded-control cursor-pointer px-5 py-2.5 text-sm font-semibold text-white active:opacity-80"
              onClick={() => void cogQuery.refetch()}
              type="button"
            >
              {m.common_retry()}
            </button>
          }
          description={m.report_error_desc()}
          title={m.report_error_title()}
          variant="empty"
        />
      </AmbientLayout>
    );
  }

  const r = wireCognitiveOf(cogQuery.data);
  const digest = wireDigestOf(digestQuery.data);
  const covered = new Set(digest.abilities.map((a) => a.ability));
  const missing = CORE_ABILITIES.filter((a) => !covered.has(a));
  const axes = r.dimensions.map((d) => ({ label: dimensionWord(d.dimension), pct: d.value }));
  const radarReady = r.dimensions.length >= 3;
  const tier = tierWord(String(r.tier ?? ""));

  return (
    <AmbientLayout>
      <div className="flex w-full flex-col gap-3 pb-6">
        <FlowHeader
          onBack={() => void navigate({ to: "/report/center" })}
          title={m.report_cognitive_title()}
        />

        {!r.access.allowed ? (
          <div className="border-line bg-surface-card shadow-card rounded-card border px-4 py-6 text-center backdrop-blur-md">
            <div className="text-ink text-[14px] font-bold">
              {m.report_cognitive_locked_title()}
            </div>
            <div className="text-ink-tertiary mt-1.5 text-[12px] leading-relaxed">
              {m.report_cognitive_locked_desc()}
            </div>
            <button
              className="bg-cta-primary rounded-control mt-4 px-5 py-2.5 text-sm font-semibold text-white active:opacity-80"
              onClick={() => void navigate({ to: "/profile" })}
              type="button"
            >
              {m.report_cognitive_locked_cta()}
            </button>
          </div>
        ) : (
          <>
            <CognitiveHead
              composite={r.composite}
              covered={digest.abilities.length}
              tested={r.testedDimensions}
              tier={tier}
              total={r.totalDimensions}
            />

            {radarReady ? (
              <AbilityRadar axes={axes} mainCount={axes.length} subCount={0} />
            ) : (
              <div className="border-line bg-surface-card rounded-card border px-4 py-4 backdrop-blur-md">
                {/* The axis count rides the title only while dimensions
                    are missing — a complete radar needs no apology. */}
                <div className="text-ink text-[13px] font-bold">
                  {m.report_cognitive_radar_title()}
                  {r.testedDimensions < r.totalDimensions
                    ? m.report_cognitive_radar_axes({ n: r.testedDimensions })
                    : ""}
                </div>
                <div className="text-ink-tertiary mt-1.5 text-[11.5px] leading-relaxed">
                  {m.report_cognitive_radar_few()}
                </div>
                <div className="text-ink-tertiary text-[11.5px] leading-relaxed">
                  {m.report_cognitive_radar_hint()}
                </div>
              </div>
            )}

            <div className="border-line bg-surface-card rounded-card border px-4 py-3.5 backdrop-blur-md">
              <div className="text-ink text-[13px] font-bold">
                {m.report_cognitive_detail_title()}
              </div>
              <div className="text-ink-tertiary text-[10.5px]">
                {m.report_cognitive_detail_tested({
                  tested: r.testedDimensions,
                  total: r.totalDimensions,
                })}
              </div>
              {r.dimensions.length > 0 ? (
                <div className="divide-tier-trap mt-2 divide-y">
                  {r.dimensions.map((d) => (
                    <div className="flex items-center justify-between py-2" key={d.dimension}>
                      <span className="text-ink text-[12.5px]">{dimensionWord(d.dimension)}</span>
                      <span className="flex items-center gap-2">
                        {/* The row's verdict is the bare letter — the
                            prototype's detail row; the full "letter ·
                            word" chip stays the head's own shape. */}
                        <span className="text-ink-tertiary text-[11px] tabular-nums">
                          {d.value}
                        </span>
                        <span
                          className={`rounded-full px-2 py-0.5 text-[9.5px] font-extrabold ${
                            DIM_TIER_STYLE[tierWord(String(d.tier ?? "")).kind] ?? ""
                          }`}
                        >
                          {tierWord(String(d.tier ?? "")).letter}
                        </span>
                      </span>
                    </div>
                  ))}
                </div>
              ) : null}
              {missing.length > 0 ? (
                <div className="text-ink-tertiary mt-2 text-[11px] leading-relaxed">
                  <div>
                    {m.report_cognitive_detail_missing({
                      n: r.totalDimensions - r.testedDimensions,
                    })}
                  </div>
                  <div>
                    {m.report_cognitive_detail_missing_abilities({
                      abilities: missing.map((a) => abilityWord(a).cn).join("、"),
                    })}
                  </div>
                  <div>{m.report_cognitive_detail_full()}</div>
                  <button
                    className="text-identity mt-2 text-[12px] font-bold active:opacity-70"
                    onClick={() => void navigate({ to: "/match" })}
                    type="button"
                  >
                    {m.report_cognitive_go_test({ n: r.totalDimensions - r.testedDimensions })}
                  </button>
                </div>
              ) : null}
            </div>
          </>
        )}

        <div className="border-line bg-surface-card rounded-card border px-4 py-3 backdrop-blur-md">
          <div className="text-ink text-[12px] font-bold">
            {m.report_cognitive_disclaimer_title()}
          </div>
          <div className="text-ink-tertiary mt-1 text-[10.5px] leading-relaxed">
            {m.report_cognitive_disclaimer()}
          </div>
        </div>

        {/* The to-trend upsell (prototype report-cognitive): the orange
            CTA rides the trend page's advanced gate — it renders only
            while the advanced tier is closed, so a member never meets
            the upsell aimed at themselves. */}
        {!r.trendAdvancedAccess.allowed ? (
          <button
            className="rounded-control bg-grade-c w-full cursor-pointer px-4 py-3.5 text-[15px] font-bold text-white active:opacity-80"
            onClick={() => void navigate({ to: "/report/trend" })}
            type="button"
          >
            {m.report_cognitive_to_trend()}
          </button>
        ) : null}
        <button
          className="rounded-control border-line text-ink-secondary w-full border px-4 py-2.5 text-[13px] font-bold active:opacity-70"
          onClick={() => void navigate({ to: "/report/center" })}
          type="button"
        >
          {m.report_cognitive_back()}
        </button>
      </div>
    </AmbientLayout>
  );
}
