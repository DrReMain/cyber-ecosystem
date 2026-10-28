import { abilityWord, narrowGenre } from "#/features/modules/match/vocabulary";
import { m } from "#/paraglide/messages";
import type { WireTrendSeries } from "../trend";

/** One ability's chronology: an SVG polyline over the session anchors,
 * points ordered oldest-first as the wire carries them. Each point
 * carries its genre — FPS and MOBA order different ability sets, so the
 * marker shape says which vocabulary the percentile spoke in, and a
 * single-point series renders its lone marker without a line. Markup-only
 * over the wire's rows; the y-axis is the 0-100 percentile floor. */
export function TrendChart({ series }: Readonly<{ series: WireTrendSeries }>) {
  const { points } = series;
  const width = 260;
  const height = 90;
  const padX = 10;
  const padY = 10;
  const span = Math.max(points.length - 1, 1);
  const x = (i: number) => padX + (i * (width - padX * 2)) / span;
  const y = (pct: number) => height - padY - (pct / 100) * (height - padY * 2);
  const line = points.map((p, i) => `${x(i)},${y(p.percentile)}`).join(" ");
  return (
    <div className="border-line bg-surface-card shadow-card rounded-card border px-4 py-3.5 backdrop-blur-md">
      <div className="mb-1 flex items-center justify-between">
        <span className="text-ink text-[13px] font-bold">{abilityWord(series.ability).cn}</span>
        <span className="text-ink-tertiary text-[10px]">
          {m.report_trend_points({ n: points.length })}
        </span>
      </div>
      <svg className="block h-auto w-full" viewBox={`0 0 ${width} ${height}`}>
        <title>{abilityWord(series.ability).cn}</title>
        <line
          stroke="#1F2937"
          strokeWidth={1}
          x1={padX}
          x2={width - padX}
          y1={height - padY}
          y2={height - padY}
        />
        {points.length > 1 && (
          <polyline fill="none" points={line} stroke="#22D3EE" strokeWidth={1.5} />
        )}
        {points.map((p, i) => {
          const genre = narrowGenre(p.genre);
          return genre === "moba" ? (
            <circle
              cx={x(i)}
              cy={y(p.percentile)}
              fill="#0B1220"
              key={`${p.at}-${String(i)}`}
              r={3}
              stroke="#A78BFA"
              strokeWidth={1.5}
            />
          ) : (
            <circle
              cx={x(i)}
              cy={y(p.percentile)}
              fill={genre === "fps" ? "#22D3EE" : "#6B7280"}
              key={`${p.at}-${String(i)}`}
              r={3}
            />
          );
        })}
      </svg>
    </div>
  );
}
