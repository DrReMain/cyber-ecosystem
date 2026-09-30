import { FileService } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/file_pb";
import { createFileRoute } from "@tanstack/react-router";
import { List } from "lucide-react";
import { pageTitle } from "#/config";
import { op, requireOperations } from "#/features/layout-dashboard/auth/permissions";
import { FilesPage, parseFilesSearch } from "#/features/modules/files";
import { m } from "#/paraglide/messages";

const operations = [op(FileService, "listFiles")];

export const Route = createFileRoute("/dashboard/files/list")({
  staticData: {
    title: "files_title",
    menu: { icon: List, order: 10 },
    operations,
  },
  validateSearch: parseFilesSearch,
  head: () => ({ meta: [{ title: pageTitle(m.files_title()) }] }),
  beforeLoad: ({ context }) => requireOperations(operations, context.permissions),
  component: FilesPage,
});
