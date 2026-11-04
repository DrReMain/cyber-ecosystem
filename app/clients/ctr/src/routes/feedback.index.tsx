import { createFileRoute } from "@tanstack/react-router";
import { FeedbackSubmitPage } from "#/features/modules/feedback/submit";

// The feedback form (prototype feedback-submit): the BLOCKED merge pane's
// 去意见反馈 CTA, the queue's 我要反馈 button, and the home bell all land
// here. Head and guard come from the layout parent.
export const Route = createFileRoute("/feedback/")({
  component: FeedbackSubmitPage,
});
