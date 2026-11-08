import type { StepRef } from "../contract";

interface ProgStripProps {
  readonly lineup: readonly StepRef[];
  readonly doneCodes: readonly string[];
  readonly currentIndex: number;
}

/** The lineup strip (prototype .prog): done segments run cyan→blue, the
 * current one glows amber, untested ones stay dark. */
export function ProgStrip({ lineup, doneCodes, currentIndex }: Readonly<ProgStripProps>) {
  return (
    <div className="flex gap-1.25">
      {lineup.map((ref, i) => {
        const done = doneCodes.includes(ref.code);
        const cur = i === currentIndex;
        return (
          <div
            className={`h-1.25 flex-1 rounded-full ${
              done
                ? "from-link to-ambient-blue bg-gradient-to-r"
                : cur
                  ? "bg-progress-cur shadow-[0_0_10px_rgba(245,158,11,0.8)]"
                  : "bg-tier-trap"
            }`}
            key={ref.code}
          />
        );
      })}
    </div>
  );
}
