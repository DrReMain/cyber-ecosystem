import { useQuery as useConnectQuery } from "@connectrpc/connect-query";
import type {
  GetReportResponse,
  ReportAbility,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/report_pb";
import {
  getAbilityDigest,
  getReport,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/report-ReportService_connectquery";
import type { Sku } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/sku_pb";
import { listSkus } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/sku-SkuService_connectquery";
import { useNavigate, useParams, useRouter } from "@tanstack/react-router";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { StateScreen } from "#/features/app/states/state-screen";
import { FlowHeader } from "#/features/modules/match/ui/flow-header";
import { abilityWord } from "#/features/modules/match/vocabulary";
import { m } from "#/paraglide/messages";
import { bestMapOf, wireDigestOf } from "./digest";
import { AbilityRadar } from "./ui/ability-radar";
import { AbilityRow } from "./ui/ability-row";
import { ComparisonCard, type ComparisonLine } from "./ui/comparison-card";
import { CompositeCard } from "./ui/composite-card";
import { FilingFooter } from "./ui/filing-footer";
import { ImprovementBoard, type ImprovementLine } from "./ui/improvement-board";
import { MemberCard } from "./ui/member-card";
import { UnlockCard } from "./ui/unlock-card";
import {
  deltaWord,
  priceYuanOf,
  type RadarAxis,
  type TierWord,
  tierOfPercentile,
  tierWord,
} from "./vocabulary";

// The transport's JSON flattening: enums arrive as value-name strings and
// WKT wrappers as scalars, so the generated Message types are narrowed by
// typeof at every read — the same wire-stamp idiom the resume card set.
type AbilityWire = ReportAbility & {
  ability?: string;
  score?: number;
  percentile?: number | null;
};

/** One ability's view row: identity, the cohort answer, the free-seat
 * flag, and the decorations the row renders. */
interface AbilityRowModel {
  ability: string;
  abilityCn: string;
  free: boolean;
  gradient: string;
  key: string;
  pct: number | null;
  previous: number | null;
  score: number;
  tier: TierWord;
  delta: { text: string; tone: "up" | "down" | "first" } | null;
}

/** The report's whole derived face, pure over the wire shapes — the page
 * component composes, this decides. */
interface ReportModel {
  unlocked: boolean;
  tier: TierWord;
  composite: number | undefined;
  axes: RadarAxis[];
  mainCount: number;
  subCount: number;
  tested: number;
  total: number;
  rows: AbilityRowModel[];
  subRows: AbilityRowModel[];
  comparison: ComparisonLine[] | null;
  improvements: ImprovementLine[];
  trialYuan: string;
  memberYuan: string;
}

// Main abilities first, highest percentile ahead, a cohort-less axis at
// the tail — the prototype's descending read; satellites never compete
// with the primaries for order.
function byPercentileDesc(a: AbilityWire, b: AbilityWire): number {
  return (b.percentile ?? -1) - (a.percentile ?? -1);
}

function pctOf(a: ReportAbility): number | null {
  return typeof a.percentile === "number" ? a.percentile : null;
}

function rowOf(
  a: ReportAbility,
  deltas: Map<string, { previous: number | null; delta: number | undefined }>,
  unlocked: boolean,
  free: boolean,
  keyPrefix: string,
): AbilityRowModel {
  const word = abilityWord(a.ability ?? "");
  const pct = pctOf(a);
  const entry = deltas.get(a.ability ?? "");
  const previous = typeof entry?.previous === "number" ? entry.previous : null;
  const delta =
    unlocked && pct !== null && entry !== undefined
      ? deltaWord(previous, typeof entry.delta === "number" ? entry.delta : null)
      : null;
  return {
    ability: a.ability ?? "",
    abilityCn: word.cn,
    free,
    gradient: word.gradient,
    key: `${keyPrefix}${a.ability ?? a.gameCode ?? ""}`,
    pct,
    previous,
    score: a.score ?? 0,
    tier: pct !== null ? tierOfPercentile(pct) : tierWord(""),
    delta,
  };
}

function reportModel(report: GetReportResponse, skus: Sku[]): ReportModel {
  const unlocked = report.unlocked ?? false;
  const abilities = (report.abilities ?? []) as AbilityWire[];
  const subAbilities = (report.subAbilities ?? []) as AbilityWire[];
  const deltas = new Map(
    (report.comparison?.deltas ?? []).map((d) => [
      d.ability ?? "",
      { previous: d.previousPercentile ?? null, delta: d.deltaPercentile },
    ]),
  );
  const sorted = [...abilities].sort(byPercentileDesc);
  // The free tier's own law (the E5-0 ruling the audit found dropped):
  // the best measured main ability's trio — tier, analysis, advice — is
  // free; every other row stays locked until the unlock. Satellites never
  // compete for the free seat, and a score-only lineup (no cohort answer
  // anywhere) keeps the seat vacant — the trio has no tier to show.
  const freeSeat = sorted.findIndex((a) => pctOf(a) !== null);
  const rows = sorted.map((a, i) => rowOf(a, deltas, unlocked, i === freeSeat, ""));
  const subRows = subAbilities.map((a) => rowOf(a, deltas, unlocked, false, "sub-"));
  const comparison: ComparisonLine[] | null =
    unlocked && report.comparison !== undefined
      ? rows.map((r) => ({
          abilityCn: r.abilityCn,
          previous: r.previous,
          delta: r.delta ?? deltaWord(null, null),
        }))
      : null;
  return {
    unlocked,
    tier: tierWord(String(report.tier ?? "")),
    composite:
      typeof report.compositePercentile === "number" ? report.compositePercentile : undefined,
    axes: [...abilities, ...subAbilities].map((a) => ({
      label: abilityWord(a.ability ?? "").cn,
      pct: pctOf(a),
    })),
    mainCount: abilities.length,
    subCount: subAbilities.length,
    tested: report.testedDimensions ?? 0,
    total: report.totalDimensions ?? 0,
    rows,
    subRows,
    comparison,
    improvements: (report.improvements ?? [])
      .filter((it) => typeof it.currentPercentile === "number")
      .map((it) => ({
        abilityCn: abilityWord(it.ability ?? "").cn,
        current: it.currentPercentile ?? 0,
        target: it.targetPercentile ?? 0,
        gap: it.gapPercent ?? 0,
      })),
    trialYuan: priceYuanOf(skus, "SKU_KIND_TRIAL", false),
    memberYuan: priceYuanOf(skus, "SKU_KIND_MEMBERSHIP", true),
  };
}

/**
 * The single-attempt ability report (prototype report-single): the free
 * tier's headline — composite percentile, tier badge, the N-gon radar
 * over measured axes only — with the per-ability detail, the last-attempt
 * comparison and the improvement board behind the unlock state, and the
 * two purchase faces (one-shot unlock, membership) reading their prices
 * off the SKU catalog, never a hardcoded number. The free tier's own
 * seat: the best measured main ability's trio (tier, analysis, advice)
 * renders unlocked for everyone, its history-best bar beside it.
 */
export function ReportPage() {
  const { recordId } = useParams({ from: "/report/$recordId" });
  const navigate = useNavigate();
  const router = useRouter();
  const reportQuery = useConnectQuery(getReport, { recordId });
  const skusQuery = useConnectQuery(listSkus, {});
  // The history-best bars ride the cross-session digest; a failed read
  // just leaves the bars absent — never blocks the report itself.
  const digestQuery = useConnectQuery(getAbilityDigest, {}, { throwOnError: false });
  const bestOf =
    digestQuery.isSuccess && digestQuery.data !== undefined
      ? bestMapOf(wireDigestOf(digestQuery.data))
      : new Map<string, number>();

  if (reportQuery.isPending) {
    return (
      <AmbientLayout>
        <StateScreen variant="loading" />
      </AmbientLayout>
    );
  }
  if (reportQuery.isError) {
    return (
      <AmbientLayout>
        <StateScreen
          action={
            <button
              className="bg-cta-primary rounded-control cursor-pointer px-5 py-2.5 text-sm font-semibold text-white active:opacity-80"
              onClick={() => void reportQuery.refetch()}
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

  const model = reportModel(reportQuery.data, skusQuery.data?.skus ?? []);
  const renderRow = (row: AbilityRowModel, isSub: boolean) => (
    <AbilityRow
      abilityCn={row.abilityCn}
      delta={row.delta}
      free={row.free}
      gradient={row.gradient}
      historyBest={bestOf.get(row.ability)}
      isSub={isSub}
      key={row.key}
      pct={row.pct}
      score={row.score}
      tier={row.tier}
      unlocked={model.unlocked}
    />
  );

  return (
    <AmbientLayout>
      <div className="flex w-full flex-col gap-3 pb-6">
        {/* The header's back follows the road the member arrived by
            (prototype: back to the page that linked here); a direct
            entry with no history falls to the lineup. */}
        <FlowHeader
          onBack={() => {
            if (router.history.canGoBack()) {
              router.history.back();
            } else {
              void navigate({ to: "/match" });
            }
          }}
          title={m.report_title()}
        />
        <CompositeCard composite={model.composite} tier={model.tier} />
        <AbilityRadar axes={model.axes} mainCount={model.mainCount} subCount={model.subCount} />
        <div className="text-ink-tertiary px-0.5 text-[11px] leading-relaxed">
          {m.report_dims({ tested: model.tested, total: model.total })}
          {model.tested < model.total ? ` · ${m.report_dims_more()}` : ""}
        </div>
        {model.rows.map((row) => renderRow(row, false))}
        {model.subRows.map((row) => renderRow(row, true))}
        {model.comparison !== null ? <ComparisonCard lines={model.comparison} /> : null}
        {model.improvements.length > 0 ? <ImprovementBoard lines={model.improvements} /> : null}
        <MemberCard priceYuan={model.memberYuan} unlocked={model.unlocked} />
        {!model.unlocked ? (
          <UnlockCard count={model.mainCount} priceYuan={model.trialYuan} />
        ) : null}
        <div className="px-2 text-center text-[10px] leading-[1.6] text-[#F87171]">
          {m.report_minor_note()}
        </div>
        <button
          className="bg-cta-practice shadow-cta-practice rounded-control w-full px-4 py-[14px] text-[15px] font-bold text-white active:opacity-80"
          onClick={() => void navigate({ to: "/profile" })}
          type="button"
        >
          {m.report_back_profile()}
        </button>
        <FilingFooter />
      </div>
    </AmbientLayout>
  );
}
