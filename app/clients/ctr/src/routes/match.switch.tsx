import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { SwitchConfirmPage } from "#/features/modules/match/switch-confirm";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/match/switch")({
  head: () => ({
    meta: [{ title: pageTitle(m.match_switch_title()) }],
  }),
  component: SwitchConfirmPage,
});
