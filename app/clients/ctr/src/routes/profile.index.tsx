import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { ProfilePage } from "#/features/modules/profile/page";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/profile/")({
  head: () => ({
    meta: [{ title: pageTitle(m.profile_title()) }],
  }),
  component: ProfilePage,
});
