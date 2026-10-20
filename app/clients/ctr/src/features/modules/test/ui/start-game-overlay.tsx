import { Timer } from "lucide-react";
import { useEffect, useState } from "react";
import { Button } from "react-aria-components";
import { m } from "#/paraglide/messages";

interface StartGameOverlayProps {
  testName: string;
  introduction: string;
  durationSeconds: number;
  onBack: () => void;
  onStart: () => void;
}

/**
 * The pre-run ritual (prototype StartGameOverlay): rules, the time-limit
 * strip, then a 3-2-1 countdown once Start is tapped — the clock stays
 * untouched until the ritual hands over. Backing out mid-countdown cancels
 * the handover.
 */
export function StartGameOverlay({
  testName,
  introduction,
  durationSeconds,
  onBack,
  onStart,
}: Readonly<StartGameOverlayProps>) {
  const [countdown, setCountdown] = useState<number | null>(null);

  useEffect(() => {
    if (countdown === null) {
      return;
    }
    if (countdown === 0) {
      onStart();
      return;
    }
    const id = setTimeout(() => setCountdown(countdown - 1), 1000);
    return () => clearTimeout(id);
  }, [countdown, onStart]);

  return (
    <div className="absolute inset-0 z-40 flex items-center justify-center bg-black/80 p-6 backdrop-blur-sm">
      {countdown !== null ? (
        <div className="flex flex-col items-center gap-3">
          <div className="font-extrabold text-7xl text-ink tabular-nums">{countdown}</div>
          <div className="text-ink-secondary text-sm">{m.test_countdown_ready()}</div>
          <button
            className="mt-4 text-ink-tertiary text-sm underline active:opacity-60"
            onClick={onBack}
            type="button"
          >
            {m.test_countdown_cancel()}
          </button>
        </div>
      ) : (
        <div className="w-full max-w-md rounded-card border border-line bg-surface-card p-6">
          <h1 className="text-center font-extrabold text-ink text-xl">{testName}</h1>
          <p className="mt-3 text-center text-ink-secondary text-sm leading-relaxed">
            {introduction}
          </p>
          <div className="mt-4 flex items-center justify-center gap-2 rounded-sm border border-[rgba(245,158,11,0.5)] bg-surface-deep px-3 py-2 text-status-pending text-xs">
            <Timer aria-hidden size={14} />
            {m.test_start_duration({ n: durationSeconds })}
          </div>
          <div className="mt-6 flex gap-3">
            <Button
              className="flex-1 cursor-pointer rounded-sm border border-line px-4 py-3 font-semibold text-ink-secondary active:opacity-80"
              onPress={onBack}
            >
              {m.test_intro_back()}
            </Button>
            <Button
              className="flex-1 cursor-pointer rounded-sm bg-cta-primary px-4 py-3 font-semibold text-white active:opacity-80"
              onPress={() => setCountdown(3)}
            >
              {m.test_start_action()}
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
