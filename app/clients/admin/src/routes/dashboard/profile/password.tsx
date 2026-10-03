import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { PasswordPage } from "#/features/modules/profile";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/dashboard/profile/password")({
  staticData: { menu: { hide: true } },
  head: () => ({ meta: [{ title: pageTitle(m.profile_password_title()) }] }),
  component: PasswordPage,
});
