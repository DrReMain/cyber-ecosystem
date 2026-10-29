import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { noindexHead } from "#/domains/seo";
import { MergeSuccessPage } from "#/features/modules/security/merge-success";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/profile/security/merge/success")({
  head: () => ({
    meta: [...noindexHead().meta, { title: pageTitle(m.merge_success_title()) }],
  }),
  component: MergeSuccessPage,
});
