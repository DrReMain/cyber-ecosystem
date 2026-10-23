import { useMutation } from "@connectrpc/connect-query";
import { submitTestAttempt } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/test-TestService_connectquery";
import { useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { ExitConfirmDialog } from "#/features/app/dialog/exit-confirm-dialog";
import { invalidateMatchSession } from "#/features/modules/match/session";
import { m } from "#/paraglide/messages";
import { getLocale } from "#/paraglide/runtime";
import { GameClock } from "./clock";
import { type AttemptRecord, EngineParams, type GameModule } from "./contract";
import { TestRunHost } from "./runtime";
import { buildSubmitRequest, type SubmitVerdict } from "./submit";
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
  /** The session step binding frozen at boot — undefined rides as practice. */
  readonly matchSessionId: string | undefined;
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
  matchSessionId,
  onRetry,
}: Readonly<TestAttemptProps>) {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const mountRef = useRef<HTMLDivElement>(null);
  const hostRef = useRef<TestRunHost | null>(null);
  const bootParams = useMemo(() => engineParams, [engineParams]);
  // The session binding freezes at boot, same law as the engine params: a
  // mid-run invalidation must not rewire the attempt already in flight.
  const [bootSessionId] = useState(matchSessionId);
  const [phase, setPhase] = useState<Phase>("intro");
  const [paused, setPaused] = useState(false);
  const [exitOpen, setExitOpen] = useState(false);
  const [remainingMs, setRemainingMs] = useState(durationSeconds * 1000);
  const [endRecord, setEndRecord] = useState<AttemptRecord | null>(null);
  const submit = useMutation(submitTestAttempt, {
    onSuccess: () => invalidateMatchSession(queryClient),
  });
  const { mutate: submitAttempt } = submit;
  // One submit per attempt — the guard survives StrictMode's double effect
  // pass (refs persist) and retries (a new record births a new attempt).
  const submittedRef = useRef<AttemptRecord | null>(null);

  useEffect(() => {
    if (endRecord === null || submittedRef.current === endRecord) {
      return;
    }
    // Only a completed run reaches the server: an aborted or failed run
    // never spends a retake (facts of that run are simply not recorded).
    if (endRecord.exit !== "completed") {
      return;
    }
    const request = buildSubmitRequest({
      code: gameModule.code,
      record: endRecord,
      matchSessionId: bootSessionId,
    });
    if (request === null) {
      return;
    }
    submittedRef.current = endRecord;
    submitAttempt(request);
  }, [bootSessionId, endRecord, gameModule.code, submitAttempt]);

  const resubmit = useCallback(() => {
    if (endRecord === null || endRecord.exit !== "completed") {
      return;
    }
    const request = buildSubmitRequest({
      code: gameModule.code,
      record: endRecord,
      matchSessionId: bootSessionId,
    });
    if (request !== null) {
      submitAttempt(request);
    }
  }, [bootSessionId, endRecord, gameModule.code, submitAttempt]);

  const verdict: SubmitVerdict =
    endRecord === null || endRecord.exit !== "completed"
      ? { phase: "idle" }
      : submit.isPending
        ? { phase: "submitting" }
        : submit.isError
          ? { phase: "error" }
          : submit.data === undefined
            ? { phase: "idle" }
            : {
                // WKT wrappers arrive JSON-flattened to scalars on this
                // transport — the casts restate the runtime shape.
                phase: "done",
                valid: submit.data.valid,
                invalidReason: submit.data.invalidReason as string | undefined,
                authoritativeScore: submit.data.authoritativeScore as number | undefined,
                grade: submit.data.grade as string | undefined,
                retestLeft: submit.data.retestLeft,
                inSession: bootSessionId !== undefined,
              };

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
          onResubmit={resubmit}
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
