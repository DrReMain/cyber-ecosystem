import { createFileRoute } from "@tanstack/react-router";
import { FolderOpen } from "lucide-react";

export const Route = createFileRoute("/dashboard/files")({
  staticData: {
    title: "files_mgmt_title",
    menu: { icon: FolderOpen, order: 200 },
  },
});
