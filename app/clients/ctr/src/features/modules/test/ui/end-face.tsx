import { Button } from "react-aria-components";
import { m } from "#/paraglide/messages";
import type { AttemptRecord } from "../contract";

interface EndFaceProps {
  record: AttemptRecord;
  onRetry: () => void;
  onExit: () => void;
}

/** The post-run face: facts only (tap/hit counts) — scoring is server-side. */
export function EndFace({ record, onRetry, onExit }: Readonly<EndFaceProps>) {
  const clicks = record.events.length;
  const hits = record.events.filter((event) => event.isHit === true).length;
  return (
    <div className="absolute inset-0 z-40 flex items-center justify-center bg-black/80 p-6 backdrop-blur-sm">
      <div className="w-full max-w-xs rounded-card border border-line bg-surface-card p-6">
        <h2 className="text-center font-extrabold text-ink text-xl">{m.test_end_title()}</h2>
        <p className="mt-3 text-center text-ink-secondary text-sm">
          {m.test_end_summary({ clicks, hits })}
        </p>
        <div className="mt-6 flex flex-col gap-3">
          <Button
            className="w-full cursor-pointer rounded-sm bg-cta-primary px-4 py-3 font-semibold text-white active:opacity-80"
            onPress={onRetry}
          >
            {m.test_end_retry()}
          </Button>
          <Button
            className="w-full cursor-pointer rounded-sm border border-line px-4 py-3 font-semibold text-ink-secondary active:opacity-80"
            onPress={onExit}
          >
            {m.test_end_exit()}
          </Button>
        </div>
      </div>
    </div>
  );
}
