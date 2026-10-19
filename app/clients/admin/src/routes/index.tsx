import {
  createFileRoute,
  // redirect
} from "@tanstack/react-router";
import { HomePage } from "#/features/modules/home/page";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/")({
  head: () => ({
    meta: [{ name: "description", content: m.home_description() }],
  }),
  component: HomePage,

  // beforeLoad: () => {
  // throw redirect({ to: "/dashboard" });
  // },
});
