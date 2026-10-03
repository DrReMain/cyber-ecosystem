import { createFileRoute } from "@tanstack/react-router";
import { Bot } from "lucide-react";

export const Route = createFileRoute("/dashboard/agents")({
  staticData: {
    title: "agents_title",
    menu: { icon: Bot, type: "group", order: 90 },
  },
});
