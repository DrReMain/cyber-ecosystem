import { FeedbackAdminService } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/feedback_admin_pb";
import { createFileRoute, Outlet } from "@tanstack/react-router";
import { Inbox } from "lucide-react";
import { op, requireOperations } from "#/features/layout-dashboard/auth/permissions";

const operations = [op(FeedbackAdminService, "listFeedbacks")];

export const Route = createFileRoute("/dashboard/feedback/tickets")({
  staticData: {
    title: "esports_feedback_tickets_title",
    menu: { icon: Inbox, order: 10 },
    operations,
  },
  beforeLoad: ({ context }) => requireOperations(operations, context.permissions),
  component: Outlet,
});
