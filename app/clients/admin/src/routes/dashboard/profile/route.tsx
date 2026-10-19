import { createFileRoute } from "@tanstack/react-router";
import { ProfileLayout } from "#/features/modules/profile";

export const Route = createFileRoute("/dashboard/profile")({
  staticData: { title: "profile_title", menu: { hide: true } },
  component: ProfileLayout,
});
