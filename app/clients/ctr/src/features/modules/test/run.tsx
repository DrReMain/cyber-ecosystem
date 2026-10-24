import { useCallback, useState } from "react";
import { TestAttempt } from "./attempt";
import type { GameModule, StepContext } from "./contract";

interface TestRunProps {
  readonly gameModule: GameModule;
  readonly testName: string;
  readonly introduction: string;
  readonly durationSeconds: number;
  readonly engineParams: Readonly<Record<string, string>>;
  /** The session step binding — undefined rides the run as practice. */
  readonly matchSessionId: string | undefined;
  /** The live session context the faces render (null = practice). */
  readonly step: StepContext | null;
  readonly category: string;
  readonly unitLabel: string;
}

/**
 * The run shell: retries remount the whole attempt by key (runId), so
 * every attempt — host, clock, overlays, state — starts virgin and the
 * dependency story stays exhaustive-clean inside TestAttempt.
 */
export function TestRun(props: Readonly<TestRunProps>) {
  const [runId, setRunId] = useState(0);
  const retry = useCallback(() => {
    setRunId((n) => n + 1);
  }, []);
  return <TestAttempt key={runId} onRetry={retry} {...props} />;
}
