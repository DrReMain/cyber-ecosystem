import { createFileRoute } from "@tanstack/react-router";
import { FileText } from "lucide-react";

export const Route = createFileRoute("/dashboard/reports")({
  staticData: {
    title: "esports_reports_title",
    menu: { icon: FileText, order: 30 },
  },
});
