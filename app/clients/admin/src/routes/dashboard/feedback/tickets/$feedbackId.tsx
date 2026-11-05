import { FeedbackAdminService } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/feedback_admin_pb";
import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { op, requireOperations } from "#/features/layout-dashboard/auth/permissions";
import { FeedbackDetailPage } from "#/features/modules/esports-feedback-detail";
import { m } from "#/paraglide/messages";

const operations = [op(FeedbackAdminService, "getFeedback")];

export const Route = createFileRoute("/dashboard/feedback/tickets/$feedbackId")({
  staticData: {
    menu: { hide: true },
    operations,
  },
  head: () => ({ meta: [{ title: pageTitle(m.esports_feedback_detail_title()) }] }),
  beforeLoad: ({ context }) => requireOperations(operations, context.permissions),
  component: FeedbackDetailPage,
});
