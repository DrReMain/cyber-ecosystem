import { createFileRoute } from "@tanstack/react-router";
import { UsersRound } from "lucide-react";

export const Route = createFileRoute("/dashboard/people")({
  staticData: {
    title: "esports_people_title",
    menu: { icon: UsersRound, dividerBefore: true, order: 10 },
  },
});
