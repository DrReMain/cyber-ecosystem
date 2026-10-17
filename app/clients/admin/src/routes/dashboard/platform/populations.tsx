import { PopulationService } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/population_pb";
import { createFileRoute } from "@tanstack/react-router";
import { Users } from "lucide-react";
import { pageTitle } from "#/config";
import { op, requireOperations } from "#/features/layout-dashboard/auth/permissions";
import { PopulationsPage, parsePopulationsSearch } from "#/features/modules/platform-populations";
import { m } from "#/paraglide/messages";

const operations = [op(PopulationService, "listPopulations")];

export const Route = createFileRoute("/dashboard/platform/populations")({
  staticData: {
    title: "platform_populations_title",
    menu: { icon: Users, order: 20 },
    operations,
  },
  validateSearch: parsePopulationsSearch,
  head: () => ({ meta: [{ title: pageTitle(m.platform_populations_title()) }] }),
  beforeLoad: ({ context }) => requireOperations(operations, context.permissions),
  component: PopulationsPage,
});
