import { useQuery as useConnectQuery } from "@connectrpc/connect-query";
import type { MatchSession } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/match_pb";
import { getMatchSession } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/match-MatchService_connectquery";
import {
  getTest,
  getTestConfig,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/test-TestService_connectquery";
import { useNavigate } from "@tanstack/react-router";
import { Button } from "react-aria-components";
import { StateScreen } from "#/features/app/states/state-screen";
import { m } from "#/paraglide/messages";
import type { StepContext } from "./contract";
import { getGameModule } from "./registry";
import { TestRun } from "./run";

interface TestBootstrapProps {
  readonly id: string;
  readonly code: string;
  readonly fallbackName: string;
  /** Land on the stored verdict, not the run (the done step's read-back). */
  readonly review: boolean;
}

// The faces' session binding (null = practice): the lineup, this test's
// slot, the recorded codes, and the retake receipt. Pure — the bootstrap
// recomputes it on every session refetch so the result face reads the
// post-submit truth.
function buildStepContext(session: MatchSession, code: string): StepContext | null {
  const steps = session.steps ?? [];
  const index = steps.findIndex((s) => s.testCode === code);
  if (index < 0) {
    return null;
  }
  return {
    lineup: steps.map((s) => ({ ability: s.ability ?? "", code: s.testCode ?? "" })),
    index,
    ability: steps[index]?.ability ?? "",
    doneCodes: session.done ?? [],
    retestLeft: session.retestLeft?.[code] ?? 0,
    firstSession: (session.firstSession as boolean | undefined) ?? true,
    anchorRecordId: session.anchorRecordId as string | undefined,
  };
}

// The id-keyed half of the bootstrap: both read-face queries fire once the
// 20-char row id is resolved (the /test/$code route param is the code, not
// the id). Faces: loading / retryable error / unknown engine. The match
// session read binds the run to its step — a session that is missing,
// undecided, or failed to load degrades the run to practice (unbound
// submit) rather than blocking the game. The step context recomputes on
// every session refetch, so the result face reads the post-submit truth.
export function TestBootstrap({ id, code, fallbackName, review }: Readonly<TestBootstrapProps>) {
  const navigate = useNavigate();
  // Each read owns its failure below — the retry face for the test/config
  // pair, and a session load failure degrades the run to practice — so
  // none may hand its error to the route boundary.
  const testQuery = useConnectQuery(getTest, { id }, { throwOnError: false });
  const configQuery = useConnectQuery(getTestConfig, { id }, { throwOnError: false });
  const sessionQuery = useConnectQuery(getMatchSession, {}, { throwOnError: false });
  const gameModule = getGameModule(code);
  const session = sessionQuery.data?.session ?? null;
  const step = session !== null ? buildStepContext(session, code) : null;
  const boundSessionId = step !== null && session?.id != null ? session.id : undefined;

  if (testQuery.isPending || configQuery.isPending || sessionQuery.isPending) {
    return <StateScreen variant="loading" />;
  }
  if (testQuery.isError || configQuery.isError) {
    return (
      <StateScreen
        action={
          <Button
            className="bg-cta-primary cursor-pointer rounded-sm px-5 py-2.5 text-sm font-semibold text-white active:opacity-80"
            onPress={() => {
              void testQuery.refetch();
              void configQuery.refetch();
            }}
          >
            {m.common_retry()}
          </Button>
        }
        description={m.test_error_load_desc()}
        title={m.test_error_load_title()}
        variant="empty"
      />
    );
  }
  if (gameModule === null) {
    return (
      <StateScreen
        action={
          <Button
            className="border-line text-ink-secondary cursor-pointer rounded-sm border px-5 py-2.5 text-sm font-semibold active:opacity-80"
            onPress={() => void navigate({ to: "/match" })}
          >
            {m.test_no_engine_exit()}
          </Button>
        }
        description={m.test_no_engine_desc()}
        title={m.test_no_engine_title()}
        variant="empty"
      />
    );
  }

  const test = testQuery.data?.test;
  const config = configQuery.data?.config;
  // A proto3-unset int32 arrives as 0 (not nullish): the || chain keeps the
  // 30s default alive for exactly the misconfigured row it exists for.
  const durationSeconds = config?.durationSeconds || test?.durationSeconds || 30;
  return (
    <TestRun
      durationSeconds={durationSeconds}
      engineParams={config?.engineParams ?? {}}
      gameModule={gameModule}
      introduction={testQuery.data?.introduction ?? ""}
      matchSessionId={boundSessionId}
      review={review}
      step={step}
      testName={test?.name ?? fallbackName}
      unitLabel={test?.unitScoreLabel ?? ""}
    />
  );
}
