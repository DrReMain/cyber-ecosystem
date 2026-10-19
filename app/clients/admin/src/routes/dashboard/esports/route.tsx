import { createFileRoute } from "@tanstack/react-router";
import { Gamepad2 } from "lucide-react";

export const Route = createFileRoute("/dashboard/esports")({
  staticData: {
    title: "esports_title",
    menu: { icon: Gamepad2, dividerBefore: true, order: 60 },
  },
});
