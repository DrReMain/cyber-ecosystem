import { m } from "#/paraglide/messages";

// The report's client vocabulary: tier words, radar geometry, money and
// delta formatting — pure functions over the wire's JSON-flattened shapes
// (enums arrive as value-name strings, WKT wrappers as scalars).

export type TierKind = "s" | "a" | "b" | "c" | "none";

export interface TierWord {
  kind: TierKind;
  letter: string;
  cn: string;
}

// The wire carries the tier as a value-name string; UNSPECIFIED is the
// frozen snapshot's honest no-verdict (every cohort too thin for a
// percentile at pin time) — it renders as 未评定, never a made-up letter.
export function tierWord(tier: string): TierWord {
  switch (tier) {
    case "REPORT_TIER_S":
      return { kind: "s", letter: "S", cn: m.report_tier_s() };
    case "REPORT_TIER_A":
      return { kind: "a", letter: "A", cn: m.report_tier_a() };
    case "REPORT_TIER_B":
      return { kind: "b", letter: "B", cn: m.report_tier_b() };
    case "REPORT_TIER_C":
      return { kind: "c", letter: "C", cn: m.report_tier_c() };
  }
  return { kind: "none", letter: "·", cn: m.report_tier_none() };
}

// The per-ability badge is decorative client vocabulary — the thresholds
// restate the contract's tier ladder (S ≥95, A ≥85, B ≥60, else C), the
// same words the server judges the composite by.
export function tierOfPercentile(pct: number): TierWord {
  if (pct >= 95) return { kind: "s", letter: "S", cn: m.report_tier_s() };
  if (pct >= 85) return { kind: "a", letter: "A", cn: m.report_tier_a() };
  if (pct >= 60) return { kind: "b", letter: "B", cn: m.report_tier_b() };
  return { kind: "c", letter: "C", cn: m.report_tier_c() };
}

// One radar axis: the ability's label and its percentile — null means the
// cohort was too thin to say anything, so the axis keeps its spoke and
// label but the shape never reaches it (score-only, never a zero axis).
export interface RadarAxis {
  label: string;
  pct: number | null;
}

export interface RadarPoint {
  x: number;
  y: number;
}

export interface RadarGeometry {
  grid: string[];
  spokes: RadarPoint[];
  shape: string | null;
  dots: RadarPoint[];
  labels: Array<{ point: RadarPoint; anchor: "start" | "middle" | "end" }>;
}

// The N-gon radar (prototype geometry): three concentric rings, one spoke
// and label per measured ability, the member's shape drawn only through
// axes that carry a percentile. A missing ability is an absent axis — the
// list only ever holds measured ones.
export function radarGeometry(axes: RadarAxis[]): RadarGeometry {
  const n = axes.length;
  const cx = 110;
  const cy = 110;
  const r = 84;
  const point = (i: number, radius: number): RadarPoint => {
    const angle = -Math.PI / 2 + (i * 2 * Math.PI) / n;
    return { x: cx + radius * Math.cos(angle), y: cy + radius * Math.sin(angle) };
  };
  const polygon = (radius: number): string =>
    axes
      .map((_, i) => {
        const p = point(i, radius);
        return `${p.x.toFixed(1)},${p.y.toFixed(1)}`;
      })
      .join(" ");
  const grid = [0.66, 0.33, 1].map((f) => polygon(r * f));
  const spokes = axes.map((_, i) => point(i, r));
  const measured = axes.filter((a) => a.pct !== null);
  // A single measured axis has no area to fill — the shape stays absent.
  const shape =
    measured.length >= 2
      ? axes
          .map((a, i) => {
            const p = point(i, (r * Math.max(a.pct ?? 0, 3)) / 100);
            return `${p.x.toFixed(1)},${p.y.toFixed(1)}`;
          })
          .join(" ")
      : null;
  const dots = measured.map((a) => {
    const i = axes.indexOf(a);
    return point(i, (r * (a.pct ?? 0)) / 100);
  });
  const labels = axes.map((_, i) => {
    const p = point(i, r + 20);
    const anchor: "start" | "middle" | "end" =
      Math.abs(p.x - cx) < 8 ? "middle" : p.x > cx ? "start" : "end";
    return { point: p, anchor };
  });
  return { grid, spokes, shape, dots, labels };
}

// Fen amounts ride the wire as bigint (Int64Value flattened); yuan is a
// display concern, never a stored one.
export function fenToYuan(fen: bigint): string {
  const yuan = Number(fen) / 100;
  return Number.isInteger(yuan) ? String(yuan) : yuan.toFixed((yuan * 10) % 1 === 0 ? 1 : 2);
}

export interface DeltaWord {
  text: string;
  tone: "up" | "down" | "first";
}

// The last-attempt comparison: null previous percentile renders as a first
// measurement, never as a zero delta.
export function deltaWord(previous: number | null, delta: number | null): DeltaWord {
  if (previous === null || previous === undefined || delta === null || delta === undefined) {
    return { text: m.report_cmp_first(), tone: "first" };
  }
  const rounded = Math.round(delta * 10) / 10;
  return rounded >= 0
    ? { text: `+${rounded.toFixed(1)}`, tone: "up" }
    : { text: `−${Math.abs(rounded).toFixed(1)}`, tone: "down" };
}

export interface ZoneCopy {
  analysis: string;
  advice: string;
}

// The per-ability detail's three-zone wording (≥85 / ≥60 / below) — the
// client-held five-zone copy until its configuration face lands; null
// when the cohort said nothing (the row then renders score-only).
export function zoneCopyOf(abilityCn: string, pct: number | null, score: number): ZoneCopy | null {
  if (pct === null) {
    return null;
  }
  const vars = {
    ability: abilityCn,
    score: String(Math.round(score * 10) / 10),
    pct: Math.round(pct),
  };
  if (pct >= 85) {
    return { analysis: m.report_analysis_high(vars), advice: m.report_advice_high() };
  }
  if (pct >= 60) {
    return { analysis: m.report_analysis_mid(vars), advice: m.report_advice_mid() };
  }
  return { analysis: m.report_analysis_low(vars), advice: m.report_advice_low() };
}
