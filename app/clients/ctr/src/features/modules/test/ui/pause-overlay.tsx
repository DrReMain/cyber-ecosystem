import { Button } from "react-aria-components";
import { m } from "#/paraglide/messages";

interface PauseOverlayProps {
  onResume: () => void;
  onExit: () => void;
}

/** The pause cover: the clock is frozen while this is up; exit runs through the confirm dialog. */
export function PauseOverlay({ onResume, onExit }: Readonly<PauseOverlayProps>) {
  return (
    <div className="absolute inset-0 z-40 flex items-center justify-center bg-black/80 p-6 backdrop-blur-sm">
      <div className="rounded-card border-line bg-surface-card w-full max-w-xs border p-6">
        <h2 className="text-ink text-center text-lg font-bold">{m.test_pause_title()}</h2>
        <div className="mt-6 flex flex-col gap-3">
          <Button
            className="rounded-control bg-cta-primary w-full cursor-pointer px-4 py-3 font-semibold text-white active:opacity-80"
            onPress={onResume}
          >
            {m.test_pause_resume()}
          </Button>
          <Button
            className="rounded-control border-line text-ink-secondary w-full cursor-pointer border px-4 py-3 font-semibold active:opacity-80"
            onPress={onExit}
          >
            {m.test_pause_exit()}
          </Button>
        </div>
      </div>
    </div>
  );
}
