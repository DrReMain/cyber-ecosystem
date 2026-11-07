import { FeedbackAdminService } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/feedback_admin_pb";
import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { op, requireOperations } from "#/features/layout-dashboard/auth/permissions";
import {
  FeedbackTicketsPage,
  parseTicketsSearch,
} from "#/features/modules/esports-feedback-tickets";
import { m } from "#/paraglide/messages";

const operations = [op(FeedbackAdminService, "listFeedbacks")];

export const Route = createFileRoute("/dashboard/feedback/tickets/")({
  staticData: {
    menu: { hide: true },
    operations,
  },
  validateSearch: parseTicketsSearch,
  head: () => ({ meta: [{ title: pageTitle(m.esports_feedback_tickets_title()) }] }),
  beforeLoad: ({ context }) => requireOperations(operations, context.permissions),
  component: FeedbackTicketsPage,
});
