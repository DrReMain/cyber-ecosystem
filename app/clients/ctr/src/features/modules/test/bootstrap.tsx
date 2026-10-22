import { useQuery as useConnectQuery } from "@connectrpc/connect-query";
import {
  getTest,
  getTestConfig,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/test-TestService_connectquery";
import { useNavigate } from "@tanstack/react-router";
import { Button } from "react-aria-components";
import { StateScreen } from "#/features/app/states/state-screen";
import { m } from "#/paraglide/messages";
import { getGameModule } from "./registry";
import { TestRun } from "./run";

interface TestBootstrapProps {
  readonly id: string;
  readonly code: string;
  readonly fallbackName: string;
}

// The id-keyed half of the bootstrap: both read-face queries fire once the
// 20-char row id is resolved (the /test/$code route param is the code, not
// the id). Faces: loading / retryable error / unknown engine.
export function TestBootstrap({ id, code, fallbackName }: Readonly<TestBootstrapProps>) {
  const navigate = useNavigate();
  const testQuery = useConnectQuery(getTest, { id });
  const configQuery = useConnectQuery(getTestConfig, { id });
  const gameModule = getGameModule(code);

  if (testQuery.isPending || configQuery.isPending) {
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
      introduction={testQuery.data?.introduction ?? m.test_start_rule_fallback()}
      testName={test?.name ?? fallbackName}
    />
  );
}
