import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { noindexHead } from "#/domains/seo";
import { MessageDetailPage } from "#/features/modules/messages/detail";
import { m } from "#/paraglide/messages";

// One letter's full body (prototype feedback-detail?kind=notice): opening
// it is the read — the badge and the list dot walk back to truth. Guard
// comes from the layout parent; the page reads its own params.
export const Route = createFileRoute("/messages/$messageId")({
  head: () => ({ meta: [...noindexHead().meta, { title: pageTitle(m.messages_detail_head()) }] }),
  component: MessageDetailPage,
});
