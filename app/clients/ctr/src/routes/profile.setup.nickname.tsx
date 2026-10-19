import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { SetupNicknamePage } from "#/features/modules/profile/setup/nickname";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/profile/setup/nickname")({
  head: () => ({
    meta: [{ title: pageTitle(m.profile_setup_nickname_head()) }],
  }),
  component: SetupNicknamePage,
});
