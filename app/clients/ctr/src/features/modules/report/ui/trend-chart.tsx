import { abilityWord, narrowGenre } from "#/features/modules/match/vocabulary";
import { m } from "#/paraglide/messages";
import type { WireTrendSeries } from "../trend";

// The growth receipt's tone ladder (prototype row chips): an upward
// delta reads green, a downward one amber, a flat line grey.
function deltaChipClass(delta: number): string {
  if (delta > 0) {
    return "bg-[rgba(34,197,94,0.15)] text-status-ok";
  }
  if (delta < 0) {
    return "bg-[rgba(234,179,8,0.18)] text-[#FDE047]";
  }
  return "bg-[rgba(107,114,128,0.15)] text-ink-tertiary";
}

/** One ability's chronology (prototype trend row): the delta chip and the
 * latest value headline the row, then the polyline over the session
 * anchors — each point carrying its genre, FPS and MOBA ordering
 * different ability sets, so the marker shape says which vocabulary the
 * percentile spoke in — and the first→last receipt under the line. A
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
  const first = points[0]?.percentile ?? 0;
  const last = points.at(-1)?.percentile ?? 0;
  const delta = Math.round((last - first) * 10) / 10;
  const deltaText = `${delta > 0 ? "↑ +" : delta < 0 ? "↓ " : "— "}${delta}`;
  return (
    <div className="border-line border-t py-3 first:border-t-0">
      <div className="flex items-center justify-between gap-2">
        <span className="text-ink text-[12.5px] font-bold">{abilityWord(series.ability).cn}</span>
        <span className="flex items-center gap-2">
          <span
            className={`rounded-full px-1.5 py-0.5 text-[10px] font-bold tabular-nums ${deltaChipClass(delta)}`}
          >
            {deltaText}
          </span>
          <span className="text-ink text-[14px] font-extrabold tabular-nums">
            {Math.round(last)}
          </span>
        </span>
      </div>
      <svg className="mt-2 block h-auto w-full" viewBox={`0 0 ${width} ${height}`}>
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
      <div className="text-ink-tertiary mt-1.5 text-[10px] tabular-nums">
        {m.report_trend_footnote({ n: points.length, first: Math.round(first), last: Math.round(last) })}
      </div>
    </div>
  );
}
