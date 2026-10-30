import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { noindexHead } from "#/domains/seo";
import { MergeConfirmPage } from "#/features/modules/security/merge";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/profile/security/merge/")({
  head: () => ({
    meta: [...noindexHead().meta, { title: pageTitle(m.security_merge_title()) }],
  }),
  component: MergeConfirmPage,
});
