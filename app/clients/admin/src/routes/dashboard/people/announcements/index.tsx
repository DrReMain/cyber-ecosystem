import { MessageAdminService } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/message_admin_pb";
import { createFileRoute } from "@tanstack/react-router";
import { BellRing } from "lucide-react";
import { pageTitle } from "#/config";
import { op, requireOperations } from "#/features/layout-dashboard/auth/permissions";
import { AnnouncementsPage } from "#/features/modules/esports-announcements";
import { m } from "#/paraglide/messages";

const operations = [op(MessageAdminService, "listAnnouncements")];

export const Route = createFileRoute("/dashboard/people/announcements/")({
  staticData: {
    title: "esports_announcements_title",
    menu: { icon: BellRing, order: 30 },
    operations,
  },
  head: () => ({ meta: [{ title: pageTitle(m.esports_announcements_title()) }] }),
  beforeLoad: ({ context }) => requireOperations(operations, context.permissions),
  component: AnnouncementsPage,
});
