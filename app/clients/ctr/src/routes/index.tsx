import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { HomePage } from "#/features/modules/home/page";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/")({
  head: () => ({
    meta: [
      { name: "description", content: m.home_description() },
      { title: pageTitle(m.home_title()) },
    ],
  }),
  component: HomePage,
});
