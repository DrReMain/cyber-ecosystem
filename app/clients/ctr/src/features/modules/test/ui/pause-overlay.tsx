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
      <div className="w-full max-w-xs rounded-card border border-line bg-surface-card p-6">
        <h2 className="text-center font-bold text-ink text-lg">{m.test_pause_title()}</h2>
        <div className="mt-6 flex flex-col gap-3">
          <Button
            className="w-full cursor-pointer rounded-control bg-cta-primary px-4 py-3 font-semibold text-white active:opacity-80"
            onPress={onResume}
          >
            {m.test_pause_resume()}
          </Button>
          <Button
            className="w-full cursor-pointer rounded-control border border-line px-4 py-3 font-semibold text-ink-secondary active:opacity-80"
            onPress={onExit}
          >
            {m.test_pause_exit()}
          </Button>
        </div>
      </div>
    </div>
  );
}
