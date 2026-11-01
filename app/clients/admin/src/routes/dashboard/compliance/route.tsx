import { createFileRoute } from "@tanstack/react-router";
import { Scale } from "lucide-react";

export const Route = createFileRoute("/dashboard/compliance")({
  staticData: {
    title: "esports_compliance_title",
    menu: { icon: Scale, order: 50 },
  },
});
