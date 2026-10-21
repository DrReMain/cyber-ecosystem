import { Pause, Timer } from "lucide-react";
import { Button } from "react-aria-components";
import { m } from "#/paraglide/messages";

interface TestHudProps {
  remainingMs: number;
  canPause: boolean;
  onPause: () => void;
}

/** Top bar: pause control + the live countdown (danger pulse under 10s — legacy law). */
export function TestHud({ remainingMs, canPause, onPause }: Readonly<TestHudProps>) {
  const seconds = (remainingMs / 1000).toFixed(2);
  const low = remainingMs < 10_000;
  return (
    <div className="flex items-center justify-between border-line border-b px-4 py-2">
      {/* The round fab look (prototype .fab: #1F2937 fill, secondary ink, circle). */}
      <Button
        aria-label={m.test_pause_action()}
        className="tap-zone flex size-8 cursor-pointer items-center justify-center rounded-full bg-surface-input text-ink-secondary disabled:opacity-40"
        isDisabled={!canPause}
        onPress={onPause}
      >
        <Pause aria-hidden size={16} />
      </Button>
      <div
        className={`flex items-center gap-1.5 font-bold text-sm tabular-nums ${
          low ? "animate-pulse text-danger-text" : "text-ink"
        }`}
      >
        <Timer aria-hidden size={15} />
        {m.test_time_left({ s: seconds })}
      </div>
    </div>
  );
}
