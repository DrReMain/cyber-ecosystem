import { useMutation } from "@connectrpc/connect-query";
import { submitTestAttempt } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/test-TestService_connectquery";
import { useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { ExitConfirmDialog } from "#/features/app/dialog/exit-confirm-dialog";
import { invalidateMatchSession } from "#/features/modules/match/session";
import { invalidateAbilityDigest } from "#/features/modules/report/digest";
import { invalidateSessionReport } from "#/features/modules/report/session";
import { m } from "#/paraglide/messages";
import { getLocale } from "#/paraglide/runtime";
import { GameClock } from "./clock";
import { type AttemptRecord, EngineParams, type GameModule, type StepContext } from "./contract";
import { TestRunHost } from "./runtime";
import { buildSubmitRequest, type SubmitVerdict } from "./submit";
import { EndFace } from "./ui/end-face";
import { IntroFace } from "./ui/intro-face";
import { PauseOverlay } from "./ui/pause-overlay";
import { TestHud } from "./ui/test-hud";

interface TestAttemptProps {
  readonly gameModule: GameModule;
  readonly testName: string;
  readonly introduction: string;
  readonly durationSeconds: number;
  readonly engineParams: Readonly<Record<string, string>>;
  /** The session binding frozen at boot — undefined rides as practice. */
  readonly matchSessionId: string | undefined;
  /** Land on the stored verdict, not the run (the done step's read-back). */
  readonly review: boolean;
  /** The live session context the faces render (null = practice). */
  readonly step: StepContext | null;
  readonly category: string;
  readonly unitLabel: string;
  readonly onRetry: () => void;
}

type Phase = "intro" | "running" | "ended";

// The mutation's face — the structural slice verdictOf reads. The plain
// proto3 scalars (valid, retestLeft) are zero-omitted in the transport's
// JSON, so they read as optional at runtime and the guard defaults
// restate the wire truth.
interface SubmitState {
  readonly isPending: boolean;
  readonly isError: boolean;
  readonly data?: {
    valid?: boolean;
    retestLeft?: number;
    invalidReason?: string;
    authoritativeScore?: number;
    grade?: string;
  };
}

// The submit chain's verdict as one pure derivation — the WKT wrappers
// arrive JSON-flattened to scalars on this transport, so the casts
// restate the runtime shape.
function verdictOf(
  endRecord: AttemptRecord | null,
  submit: SubmitState,
  inSession: boolean,
): SubmitVerdict {
  if (endRecord === null || endRecord.exit !== "completed") {
    return { phase: "idle" };
  }
  if (submit.isPending) {
    return { phase: "submitting" };
  }
  if (submit.isError) {
    return { phase: "error" };
  }
  if (submit.data === undefined) {
    return { phase: "idle" };
  }
  const data = submit.data;
  return {
    phase: "done",
    valid: data.valid ?? false,
    invalidReason: data.invalidReason as string | undefined,
    authoritativeScore: data.authoritativeScore as number | undefined,
    grade: data.grade as string | undefined,
    retestLeft: data.retestLeft ?? 0,
    inSession,
  };
}

// The time bar under the stage (prototype .bar): elapsed share of the
// run, cyan→blue, pinned to a 6px minimum once anything has elapsed.
function elapsedPctOf(remainingMs: number, durationSeconds: number): number {
  const totalMs = durationSeconds * 1000;
  return Math.max(0, Math.min(100, ((totalMs - remainingMs) / totalMs) * 100));
}

function RunTimeBar({ pct }: Readonly<{ pct: number }>) {
  return (
    <div className="flex-none px-4 pt-1.5 pb-2">
      <div className="h-[6px] w-full overflow-hidden rounded-full bg-[#111827]">
        <div
          className="h-full rounded-full bg-gradient-to-r from-[#0e7490] to-[#22d3ee]"
          style={{ minWidth: pct > 0 ? "6px" : undefined, width: `${pct}%` }}
        />
      </div>
    </div>
  );
}

/**
 * One attempt's lifetime: host, stage, overlays. Mounted fresh per attempt
 * (the parent keys on runId), so a retry starts from virgin state. The
 * engine params are frozen at mount — an attempt runs on the config it
 * booted with; later refetches never tear down a live run. The step
 * context stays live so the result face reads the post-submit truth.
 */
export function TestAttempt({
  gameModule,
  testName,
  introduction,
  durationSeconds,
  engineParams,
  matchSessionId,
  review,
  step,
  category,
  unitLabel,
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
  // The read-back mode starts and stays on the stored verdict — the
  // intro/run never mount a host start.
  const [phase, setPhase] = useState<Phase>(review ? "ended" : "intro");
  const [paused, setPaused] = useState(false);
  const [exitOpen, setExitOpen] = useState(false);
  const [remainingMs, setRemainingMs] = useState(durationSeconds * 1000);
  const [endRecord, setEndRecord] = useState<AttemptRecord | null>(null);
  const submit = useMutation(submitTestAttempt, {
    onSuccess: () => {
      invalidateMatchSession(queryClient);
      // The attempt just became the scoreboard's next line — the
      // settlement tags and the attempts note read the refreshed truth,
      // and every aggregate the digest carries (counts, bests, the
      // latest lines) moved with it.
      if (bootSessionId !== undefined) {
        invalidateSessionReport(queryClient, bootSessionId);
      }
      invalidateAbilityDigest(queryClient);
    },
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

  const verdict = verdictOf(endRecord, submit, bootSessionId !== undefined);

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

  // The exit dialog names the session's progress when the run is bound.
  const exitDesc =
    step !== null
      ? m.test_exit_desc_session({ done: step.doneCodes.length, total: step.lineup.length })
      : m.test_exit_desc();

  return (
    <div className="safe-pb safe-pt bg-canvas text-ink h-app relative isolate flex flex-col overflow-hidden">
      <TestHud
        canPause={phase === "running" && !paused}
        onPause={pauseRun}
        remainingMs={remainingMs}
        step={step}
        testName={testName}
      />
      <div className="min-h-0 flex-1" ref={mountRef} />
      {phase === "running" ? <RunTimeBar pct={elapsedPctOf(remainingMs, durationSeconds)} /> : null}

      {phase === "intro" && (
        <IntroFace
          category={category}
          introduction={introduction}
          onBack={backToFlow}
          onStart={startRun}
          step={step}
          testName={testName}
          unitLabel={unitLabel}
        />
      )}
      {phase === "running" && paused && (
        <PauseOverlay onExit={() => setExitOpen(true)} onResume={resumeRun} />
      )}
      {phase === "ended" && (review || endRecord !== null) && (
        <EndFace
          code={gameModule.code}
          engineParams={bootParams}
          onExit={backToFlow}
          onResubmit={resubmit}
          onRetry={onRetry}
          record={review ? null : endRecord}
          review={review}
          sessionId={bootSessionId}
          step={step}
          testName={testName}
          verdict={verdict}
        />
      )}

      <ExitConfirmDialog
        description={exitDesc}
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
