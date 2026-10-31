import { TestAdminService } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/test_admin_pb";
import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { op, requireOperations } from "#/features/layout-dashboard/auth/permissions";
import { EsportsTestsPage } from "#/features/modules/esports-tests";
import { m } from "#/paraglide/messages";

const operations = [
  op(TestAdminService, "listTests"),
  op(TestAdminService, "getTestConfig"),
  op(TestAdminService, "updateTestConfig"),
  op(TestAdminService, "listQuestionBank"),
];

export const Route = createFileRoute("/dashboard/assessment/game-config/")({
  staticData: { title: "esports_tests_title", menu: { order: 10 }, operations },
  head: () => ({ meta: [{ title: pageTitle(m.esports_tests_title()) }] }),
  beforeLoad: ({ context }) => requireOperations(operations, context.permissions),
  component: EsportsTestsPage,
});
