import { useNavigate } from "@tanstack/react-router";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { ExitConfirmDialog } from "#/features/app/dialog/exit-confirm-dialog";
import { m } from "#/paraglide/messages";
import { getLocale } from "#/paraglide/runtime";
import { GameClock } from "./clock";
import { type AttemptRecord, EngineParams, type GameModule } from "./contract";
import { TestRunHost } from "./runtime";
import { buildSubmitAttempt } from "./submit";
import { EndFace } from "./ui/end-face";
import { PauseOverlay } from "./ui/pause-overlay";
import { StartGameOverlay } from "./ui/start-game-overlay";
import { TestHud } from "./ui/test-hud";
import type { Verdict } from "./ui/verdict-card";
import { useSubmitAttempt } from "./use-submit-attempt";

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
  const submitAttempt = useSubmitAttempt();
  // Stable across renders (react-query) — safe as an effect dependency,
  // unlike the mutation state object it hangs off.
  const submitMutate = submitAttempt.mutate;
  // The built request for the attempt that just ended — kept for resubmit
  // (transport refusals), read only in callbacks, never on the render path.
  const submittedInputRef = useRef<ReturnType<typeof buildSubmitAttempt>>(null);
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
        // One submit per natural end, fired from the end event itself (an
        // effect would double-fire under StrictMode). buildSubmitAttempt
        // refuses non-completed exits and unmapped codes — those stay local.
        const input = buildSubmitAttempt({ code: gameModule.code, record });
        submittedInputRef.current = input;
        if (input !== null) {
          submitMutate(input);
        }
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
  }, [gameModule, durationSeconds, bootParams, submitMutate]);

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

  const resubmit = useCallback(() => {
    const input = submittedInputRef.current;
    if (input !== null) {
      submitMutate(input);
    }
  }, [submitMutate]);

  // The verdict the end face renders: pending → error (resubmit-able) →
  // the receipt. Idle means the attempt never crossed the seam (unmapped
  // code or non-completed exit) — the face stays local-only, no card.
  const verdict: Verdict = submitAttempt.isPending
    ? { phase: "pending" }
    : submitAttempt.isError
      ? { phase: "error", error: submitAttempt.error, onResubmit: resubmit }
      : submitAttempt.isSuccess && submitAttempt.data !== undefined
        ? { phase: "done", receipt: submitAttempt.data }
        : { phase: "local" };

  return (
    <div className="safe-pb safe-pt h-app bg-canvas text-ink relative isolate flex flex-col overflow-hidden">
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
          verdict={verdict}
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
