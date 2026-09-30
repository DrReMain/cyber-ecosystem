import { AuditService } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/audit_pb";
import { createFileRoute } from "@tanstack/react-router";
import { ScrollText } from "lucide-react";
import { pageTitle } from "#/config";
import { op, requireOperations } from "#/features/layout-dashboard/auth/permissions";
import { AuditPage, parseAuditSearch } from "#/features/modules/system-audit";
import { m } from "#/paraglide/messages";

const operations = [op(AuditService, "listAuditLogs")];

export const Route = createFileRoute("/dashboard/system/audit")({
  staticData: {
    title: "system_audit_title",
    menu: { icon: ScrollText, order: 35 },
    operations,
  },
  validateSearch: parseAuditSearch,
  head: () => ({ meta: [{ title: pageTitle(m.system_audit_title()) }] }),
  beforeLoad: ({ context }) => requireOperations(operations, context.permissions),
  component: AuditPage,
});
