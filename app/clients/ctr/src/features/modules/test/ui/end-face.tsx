import { Button } from "react-aria-components";
import { m } from "#/paraglide/messages";
import { type AttemptRecord, EngineParams } from "../contract";
import { deriveLocalResults } from "../results";

interface EndFaceProps {
  readonly code: string;
  readonly record: AttemptRecord;
  /** The attempt's boot params — the derivations re-read the same thresholds the game ran with. */
  readonly engineParams: Readonly<Record<string, string>>;
  readonly onRetry: () => void;
  readonly onExit: () => void;
}

// Row key → label copy (every derivation key has its bilingual home here).
const ROW_LABELS: Record<string, () => string> = {
  taps: m.test_result_taps,
  hits: m.test_result_hits,
  avg_reaction: m.test_result_avg_reaction,
  misses: m.test_result_misses,
  center_hits: m.test_result_center_hits,
  hit_rate: m.test_result_hit_rate,
  avg_aim: m.test_result_avg_aim,
  score: m.test_result_score,
  high_hits: m.test_result_high_hits,
  trap_hits: m.test_result_trap_hits,
  missed_high: m.test_result_missed_high,
  incorrect_clicks: m.test_result_incorrect_clicks,
  avg_decision: m.test_result_avg_decision,
  enemy_hits: m.test_result_enemy_hits,
  enemy_missed: m.test_result_enemy_missed,
  friendly_misclicks: m.test_result_friendly_misclicks,
  neutral_misclicks: m.test_result_neutral_misclicks,
  blank_clicks: m.test_result_blank_clicks,
};

// The per-game local result face: facts and local aggregates only — the
// authoritative score/grade arrive with the server submit seam and the
// report face. Markup-only; every number comes from results.ts.
export function EndFace({ code, record, engineParams, onRetry, onExit }: Readonly<EndFaceProps>) {
  const rows = deriveLocalResults(code, record, new EngineParams(engineParams));
  return (
    <div className="absolute inset-0 z-40 flex flex-col overflow-y-auto bg-black/80 p-6 backdrop-blur-sm">
      <div className="mx-auto flex w-full max-w-xs flex-1 flex-col justify-center py-6">
        <h2 className="text-ink text-center text-xl font-extrabold">{m.test_end_title()}</h2>
        <div className="mt-5 grid grid-cols-2 gap-2.5">
          {rows.map((row) => (
            <div
              className="rounded-card border-line bg-surface-card border px-3 py-2.5"
              key={row.key}
            >
              <div className="text-ink-tertiary text-xs">
                {(ROW_LABELS[row.key] ?? m.test_result_taps)()}
              </div>
              <div className="text-ink mt-0.5 text-lg font-bold">{row.value}</div>
            </div>
          ))}
        </div>
        <div className="mt-6 flex flex-col gap-3">
          {/* Retry is the practice/orange CTA (prototype btn-org), not the primary blue. */}
          <Button
            className="rounded-control bg-cta-practice w-full cursor-pointer px-4 py-3 font-semibold text-white active:opacity-80"
            onPress={onRetry}
          >
            {m.test_end_retry()}
          </Button>
          <Button
            className="rounded-control border-line text-ink-secondary w-full cursor-pointer border px-4 py-3 font-semibold active:opacity-80"
            onPress={onExit}
          >
            {m.test_end_exit()}
          </Button>
        </div>
      </div>
    </div>
  );
}
