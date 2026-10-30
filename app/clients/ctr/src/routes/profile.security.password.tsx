import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { noindexHead } from "#/domains/seo";
import { PasswordPage } from "#/features/modules/security/password";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/profile/security/password")({
  validateSearch: (search: Record<string, unknown>): { mode?: "set" | "change" } => ({
    ...(search.mode === "change" ? { mode: "change" } : {}),
  }),
  head: () => ({
    meta: [...noindexHead().meta, { title: pageTitle(m.security_pw_set_title()) }],
  }),
  component: PasswordPage,
});
