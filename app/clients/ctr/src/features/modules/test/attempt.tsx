import { useNavigate } from "@tanstack/react-router";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { ExitConfirmDialog } from "#/features/app/dialog/exit-confirm-dialog";
import { m } from "#/paraglide/messages";
import { getLocale } from "#/paraglide/runtime";
import { GameClock } from "./clock";
import { type AttemptRecord, EngineParams, type GameModule } from "./contract";
import { TestRunHost } from "./runtime";
import { EndFace } from "./ui/end-face";
import { PauseOverlay } from "./ui/pause-overlay";
import { StartGameOverlay } from "./ui/start-game-overlay";
import { TestHud } from "./ui/test-hud";

interface TestAttemptProps {
  readonly gameModule: GameModule;
  readonly testName: string;
  readonly introduction: string;
  readonly durationSeconds: number;
  readonly engineParams: Readonly<Record<string, string>>;
  readonly onRetry: () => void;
}

type Phase = "intro" | "running" | "ended";

/**
 * One attempt's lifetime: host, stage, overlays. Mounted fresh per attempt
 * (the parent keys on runId), so a retry starts from virgin state. The
 * engine params are frozen at mount — an attempt runs on the config it
 * booted with; later refetches never tear down a live run.
 */
export function TestAttempt({
  gameModule,
  testName,
  introduction,
  durationSeconds,
  engineParams,
  onRetry,
}: Readonly<TestAttemptProps>) {
  const navigate = useNavigate();
  const mountRef = useRef<HTMLDivElement>(null);
  const hostRef = useRef<TestRunHost | null>(null);
  const bootParams = useMemo(() => engineParams, [engineParams]);
  const [phase, setPhase] = useState<Phase>("intro");
  const [paused, setPaused] = useState(false);
  const [exitOpen, setExitOpen] = useState(false);
  const [remainingMs, setRemainingMs] = useState(durationSeconds * 1000);
  const [endRecord, setEndRecord] = useState<AttemptRecord | null>(null);

  useEffect(() => {
    const mount = mountRef.current;
    if (mount === null) {
      return;
    }
    const durationMs = durationSeconds * 1000;
    const clock = new GameClock({ durationMs });
    const host = new TestRunHost({
      module: gameModule,
      clock,
      mount,
      params: new EngineParams(bootParams),
      durationMs,
      locale: getLocale(),
      now: () => Date.now(),
      onEnd: (record) => {
        setEndRecord(record);
        setPhase("ended");
      },
    });
    hostRef.current = host;
    setRemainingMs(durationMs);
    const unsubscribe = clock.onTick((snapshot) => setRemainingMs(snapshot.remainingMs));
    return () => {
      unsubscribe();
      host.destroy();
      hostRef.current = null;
    };
  }, [gameModule, durationSeconds, bootParams]);

  const startRun = useCallback(() => {
    hostRef.current?.start();
    setPhase("running");
  }, []);

  const pauseRun = useCallback(() => {
    hostRef.current?.pause();
    setPaused(true);
  }, []);

  const resumeRun = useCallback(() => {
    hostRef.current?.resume();
    setPaused(false);
  }, []);

  const backToFlow = useCallback(() => {
    hostRef.current?.destroy();
    void navigate({ to: "/match" });
  }, [navigate]);

  return (
    <div className="safe-pb safe-pt relative isolate flex h-app flex-col overflow-hidden bg-canvas text-ink">
      <TestHud
        canPause={phase === "running" && !paused}
        onPause={pauseRun}
        remainingMs={remainingMs}
      />
      <div className="min-h-0 flex-1" ref={mountRef} />

      {phase === "intro" && (
        <StartGameOverlay
          durationSeconds={durationSeconds}
          introduction={introduction}
          onBack={backToFlow}
          onStart={startRun}
          testName={testName}
        />
      )}
      {phase === "running" && paused && (
        <PauseOverlay onExit={() => setExitOpen(true)} onResume={resumeRun} />
      )}
      {phase === "ended" && endRecord !== null && (
        <EndFace
          code={gameModule.code}
          engineParams={bootParams}
          onExit={backToFlow}
          onRetry={onRetry}
          record={endRecord}
        />
      )}

      <ExitConfirmDialog
        description={m.test_exit_desc()}
        onCancel={() => setExitOpen(false)}
        onConfirm={() => {
          setExitOpen(false);
          backToFlow();
        }}
        open={exitOpen}
        title={m.test_exit_title()}
      />
    </div>
  );
}
