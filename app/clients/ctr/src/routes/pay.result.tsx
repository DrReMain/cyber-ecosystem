import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { noindexHead } from "#/domains/seo";
import { PayResultPage } from "#/features/modules/pay/result-page";
import { payResultSearchOf } from "#/features/modules/pay/search";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/pay/result")({
  head: () => ({ meta: [...noindexHead().meta, { title: pageTitle(m.pay_result_head()) }] }),
  validateSearch: (search: Record<string, unknown>) => payResultSearchOf(search),
  component: PayResultPage,
});
