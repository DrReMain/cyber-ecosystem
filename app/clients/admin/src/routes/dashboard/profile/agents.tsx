import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { AgentsPage } from "#/features/modules/profile";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/dashboard/profile/agents")({
  staticData: { menu: { hide: true } },
  head: () => ({ meta: [{ title: pageTitle(m.profile_agents_title()) }] }),
  component: AgentsPage,
});
