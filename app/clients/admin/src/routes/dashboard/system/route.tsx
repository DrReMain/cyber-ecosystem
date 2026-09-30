import { createFileRoute } from "@tanstack/react-router";
import { Cog } from "lucide-react";

export const Route = createFileRoute("/dashboard/system")({
  staticData: {
    title: "system_title",
    menu: { icon: Cog, dividerBefore: true, order: 100 },
  },
});
