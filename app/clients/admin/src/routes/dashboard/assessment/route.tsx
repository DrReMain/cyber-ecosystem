import { createFileRoute } from "@tanstack/react-router";
import { ClipboardList } from "lucide-react";

export const Route = createFileRoute("/dashboard/assessment")({
  staticData: {
    title: "esports_assessment_title",
    menu: { icon: ClipboardList, order: 20 },
  },
});
