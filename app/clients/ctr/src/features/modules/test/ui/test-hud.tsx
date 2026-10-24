import { Pause, Sun, Timer } from "lucide-react";
import { Button } from "react-aria-components";
import { abilityWord } from "#/features/modules/match/vocabulary";
import { m } from "#/paraglide/messages";
import type { StepContext } from "../contract";

interface TestHudProps {
  readonly remainingMs: number;
  readonly canPause: boolean;
  readonly onPause: () => void;
  readonly testName: string;
  readonly step: StepContext | null;
}

/** The run's top bar (prototype test-running header row): pause control,
 * the run title over the keep-awake law, and the live countdown (danger
 * pulse under 10s — legacy law) above the lineup position. */
export function TestHud({
  remainingMs,
  canPause,
  onPause,
  testName,
  step,
}: Readonly<TestHudProps>) {
  const seconds = (remainingMs / 1000).toFixed(2);
  const low = remainingMs < 10_000;
  return (
    <div className="border-line flex items-center gap-2.5 border-b px-3.5 py-2">
      <Button
        aria-label={m.test_pause_action()}
        className="tap-zone bg-surface-input text-ink-secondary flex size-8 shrink-0 cursor-pointer items-center justify-center rounded-full disabled:opacity-40"
        isDisabled={!canPause}
        onPress={onPause}
      >
        <Pause aria-hidden size={16} />
      </Button>
      <div className="min-w-0 flex-1">
        <div className="text-ink truncate text-[14.5px] font-extrabold">
          {m.test_run_title({ name: testName })}
        </div>
        <div className="text-ink-tertiary mt-px flex items-center gap-1 text-[10px] leading-snug">
          <Sun aria-hidden className="shrink-0" size={10} />
          <span className="truncate">{m.test_run_hint()}</span>
        </div>
      </div>
      <div className="flex shrink-0 flex-col items-end">
        <div
          className={`flex items-center gap-1.5 text-sm font-bold tabular-nums ${
            low ? "text-danger-text animate-pulse" : "text-ink"
          }`}
        >
          <Timer aria-hidden size={15} />
          {m.test_time_left({ s: seconds })}
        </div>
        {step !== null ? (
          <div className="text-ink-tertiary mt-px text-[10.5px] tabular-nums">
            {m.test_run_idx({
              ability: abilityWord(step.ability).cn,
              i: step.index + 1,
              total: step.lineup.length,
            })}
          </div>
        ) : null}
      </div>
    </div>
  );
}
