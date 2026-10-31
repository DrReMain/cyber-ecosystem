import { createFileRoute } from "@tanstack/react-router";
import { Megaphone } from "lucide-react";

export const Route = createFileRoute("/dashboard/channels")({
  staticData: {
    title: "esports_channels_title",
    menu: { icon: Megaphone, order: 40 },
  },
});
