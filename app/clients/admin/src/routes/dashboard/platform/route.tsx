import { createFileRoute } from "@tanstack/react-router";
import { Globe } from "lucide-react";

export const Route = createFileRoute("/dashboard/platform")({
  staticData: {
    title: "platform_title",
    menu: { icon: Globe, order: 300 },
  },
});
