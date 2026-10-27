import { m } from "#/paraglide/messages";
import { type RadarAxis, radarGeometry } from "../vocabulary";

/** The N-gon radar card (prototype ②): one axis per measured ability —
 * a missing ability is an absent axis, never a zero axis; a measured
 * ability with no percentile keeps its spoke but the shape never reaches
 * it (score-only). Sub-abilities ride as satellites with their own note.
 * Markup-only over radarGeometry's pure output. */
export function AbilityRadar({
  axes,
  mainCount,
  subCount,
}: Readonly<{
  axes: RadarAxis[];
  mainCount: number;
  subCount: number;
}>) {
  const geo = radarGeometry(axes);
  return (
    <div className="border-line bg-surface-card shadow-card rounded-card border px-4 py-3.5 backdrop-blur-md">
      <div className="mb-1 flex items-baseline justify-between">
        <span className="text-ink text-[13.5px] font-bold">
          {m.report_radar_title({ n: axes.length })}
        </span>
        <span className="text-ink-tertiary text-[10px]">
          {m.report_radar_subs({ main: mainCount, subs: subCount })}
        </span>
      </div>
      <div className="w-full">
        {/* biome-ignore lint/a11y/noSvgWithoutTitle: the title below is
            localized through paraglide — a dynamic expression the rule's
            literal-only check cannot see. */}
        <svg className="block h-auto w-full" viewBox="0 0 220 220">
          {/* The chart's accessible name; the shapes themselves are
              decorative over the per-ability rows below. */}
          <title>{m.report_radar_title({ n: axes.length })}</title>
          {geo.grid.map((points) => (
            <polygon fill="none" key={points} points={points} stroke="#1F2937" strokeWidth={1} />
          ))}
          {geo.spokes.map((p) => (
            <line
              key={`${p.x},${p.y}`}
              stroke="#1F2937"
              strokeWidth={1}
              x1={110}
              x2={p.x}
              y1={110}
              y2={p.y}
            />
          ))}
          {geo.shape !== null ? (
            <polygon
              fill="rgba(6,182,212,0.18)"
              points={geo.shape}
              stroke="#22D3EE"
              strokeWidth={2}
            />
          ) : null}
          {geo.dots.map((p) => (
            <circle cx={p.x} cy={p.y} fill="#22D3EE" key={`${p.x},${p.y}`} r={3.5} />
          ))}
          {axes.map((axis, i) => (
            <text
              dominantBaseline="middle"
              fill="#9CA3AF"
              fontSize={10}
              key={axis.label}
              textAnchor={geo.labels[i]?.anchor}
              x={geo.labels[i]?.point.x}
              y={geo.labels[i]?.point.y}
            >
              {axis.label}
            </text>
          ))}
        </svg>
      </div>
      <div className="bg-surface-input text-ink-tertiary mt-1 rounded-lg px-3 py-2.5 text-[11px] leading-[1.7]">
        {subCount > 0 ? m.report_radar_sub_note() : m.report_radar_no_subs()}
      </div>
    </div>
  );
}
