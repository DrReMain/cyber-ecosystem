import { createFileRoute } from "@tanstack/react-router";
import { Headset } from "lucide-react";

export const Route = createFileRoute("/dashboard/feedback")({
  staticData: {
    title: "esports_feedback_title",
    menu: { icon: Headset, order: 45 },
  },
});
