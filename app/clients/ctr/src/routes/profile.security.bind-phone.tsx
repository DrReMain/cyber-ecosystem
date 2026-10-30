import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { noindexHead } from "#/domains/seo";
import { BindPhonePage } from "#/features/modules/security/bind-phone";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/profile/security/bind-phone")({
  validateSearch: (search: Record<string, unknown>): { from?: "me" | "login" | "guide" } => ({
    ...(search.from === "login" || search.from === "guide" ? { from: search.from } : {}),
  }),
  head: () => ({
    meta: [...noindexHead().meta, { title: pageTitle(m.security_bind_phone_title()) }],
  }),
  component: BindPhonePage,
});
