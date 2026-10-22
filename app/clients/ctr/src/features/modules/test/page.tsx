import { useQuery as useConnectQuery } from "@connectrpc/connect-query";
import { listTests } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/test-TestService_connectquery";
import { useNavigate, useParams } from "@tanstack/react-router";
import { StateScreen } from "#/features/app/states/state-screen";
import { m } from "#/paraglide/messages";
import { TestBootstrap } from "./bootstrap";

/**
 * The /test/$code host page. The route param is the test code; the row id
 * (what the read face keys on) is resolved through the same listTests
 * cache the match flow already hydrated.
 */
export function TestPage() {
  const { code } = useParams({ from: "/test/$code" });
  const navigate = useNavigate();
  const listQuery = useConnectQuery(listTests, {});
  const row = (listQuery.data?.tests ?? []).find((test) => test.code === code) ?? null;

  if (listQuery.isPending) {
    return <StateScreen variant="loading" />;
  }
  if (listQuery.isError) {
    return (
      <StateScreen
        description={m.test_error_load_desc()}
        title={m.test_error_load_title()}
        variant="empty"
      />
    );
  }
  if (row === null || row.id === undefined) {
    return (
      <StateScreen
        action={
          <button
            className="border-line text-ink-secondary cursor-pointer rounded-sm border px-5 py-2.5 text-sm font-semibold active:opacity-80"
            onClick={() => void navigate({ to: "/match" })}
            type="button"
          >
            {m.test_unknown_exit()}
          </button>
        }
        description={m.test_unknown_desc()}
        title={m.test_unknown_title()}
        variant="empty"
      />
    );
  }
  return <TestBootstrap code={code} fallbackName={row.name ?? code} id={row.id} />;
}
