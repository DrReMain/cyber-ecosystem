import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { noindexHead } from "#/domains/seo";
import { sessionGuard } from "#/features/auth/guard";
import { PayConfirmPage } from "#/features/modules/pay/page";
import { paySearchOf } from "#/features/modules/pay/search";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/pay")({
  beforeLoad: sessionGuard,
  head: () => ({ meta: [...noindexHead().meta, { title: pageTitle(m.pay_title()) }] }),
  validateSearch: (search: Record<string, unknown>) => paySearchOf(search),
  component: PayConfirmPage,
});
