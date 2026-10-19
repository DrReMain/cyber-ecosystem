import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { SetupAgePage } from "#/features/modules/profile/setup/age";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/profile/setup/age")({
  head: () => ({
    meta: [{ title: pageTitle(m.profile_setup_age_head()) }],
  }),
  component: SetupAgePage,
});
