import { createFileRoute, Outlet } from "@tanstack/react-router";
import { Banknote } from "lucide-react";

export const Route = createFileRoute("/dashboard/pricing")({
  staticData: {
    title: "esports_pricing_title",
    menu: { icon: Banknote, order: 25 },
  },
  component: Outlet,
});
