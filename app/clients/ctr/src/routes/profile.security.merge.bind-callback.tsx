import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { noindexHead } from "#/domains/seo";
import { MergeBindCallbackPage } from "#/features/modules/security/merge-callbacks";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/profile/security/merge/bind-callback")({
  validateSearch: (search: Record<string, unknown>): { code?: string; state?: string } => ({
    ...(typeof search.code === "string" ? { code: search.code } : {}),
    ...(typeof search.state === "string" ? { state: search.state } : {}),
  }),
  head: () => ({
    meta: [...noindexHead().meta, { title: pageTitle(m.merge_pending()) }],
  }),
  component: MergeBindCallbackPage,
});
