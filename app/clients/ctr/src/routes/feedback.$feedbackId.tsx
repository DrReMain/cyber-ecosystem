import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { noindexHead } from "#/domains/seo";
import { FeedbackDetailPage } from "#/features/modules/feedback/detail";
import { m } from "#/paraglide/messages";

// The feedback conversation (prototype feedback-detail?kind=fb): the
// ticket root, its thread of turns, and the follow-up bar. Guard comes
// from the layout parent; the page reads its own params.
export const Route = createFileRoute("/feedback/$feedbackId")({
  head: () => ({ meta: [...noindexHead().meta, { title: pageTitle(m.feedback_detail_head()) }] }),
  component: FeedbackDetailPage,
});
