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
    <div className="flex gap-[5px]">
      {lineup.map((ref, i) => {
        const done = doneCodes.includes(ref.code);
        const cur = i === currentIndex;
        return (
          <div
            className={`h-[5px] flex-1 rounded-full ${
              done
                ? "bg-gradient-to-r from-[#22d3ee] to-[#2563eb]"
                : cur
                  ? "bg-[#f59e0b] shadow-[0_0_10px_rgba(245,158,11,0.8)]"
                  : "bg-[#1f2937]"
            }`}
            key={ref.code}
          />
        );
      })}
    </div>
  );
}
