import { PolicyService } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/policy_pb";
import { createFileRoute } from "@tanstack/react-router";
import { CalendarClock } from "lucide-react";
import { pageTitle } from "#/config";
import { op, requireOperations } from "#/features/layout-dashboard/auth/permissions";
import { PoliciesPage, parsePoliciesSearch } from "#/features/modules/system-policies";
import { m } from "#/paraglide/messages";

const operations = [op(PolicyService, "listPolicies"), op(PolicyService, "getPolicy")];

export const Route = createFileRoute("/dashboard/system/policies")({
  staticData: {
    title: "system_policies_title",
    menu: { icon: CalendarClock, order: 25 },
    operations,
  },
  validateSearch: parsePoliciesSearch,
  head: () => ({ meta: [{ title: pageTitle(m.system_policies_title()) }] }),
  beforeLoad: ({ context }) => requireOperations(operations, context.permissions),
  component: PoliciesPage,
});
