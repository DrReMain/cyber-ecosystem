import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { noindexHead } from "#/domains/seo";
import { profileFlowGuard } from "#/features/auth/guard";
import { TestPage } from "#/features/modules/test/page";
import { m } from "#/paraglide/messages";

// The game runtime's parameterized face (prototype test-running — the only
// landscape page): the member flow guard, per guard.ts's own reservation
// ("the test runtime later"), and the personal-flow noindex stance.
export const Route = createFileRoute("/test/$code")({
  beforeLoad: profileFlowGuard,
  head: () => ({ meta: [...noindexHead().meta, { title: pageTitle(m.test_title()) }] }),
  component: TestPage,
});
