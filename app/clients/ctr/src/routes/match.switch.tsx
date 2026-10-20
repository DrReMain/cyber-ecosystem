import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { SwitchConfirmPage } from "#/features/modules/match/switch-confirm";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/match/switch")({
  validateSearch: (search: Record<string, unknown>): { from?: "fps" | "moba" } => ({
    ...(search.from === "fps" || search.from === "moba" ? { from: search.from } : {}),
  }),
  head: () => ({
    meta: [{ title: pageTitle(m.match_switch_title()) }],
  }),
  component: SwitchConfirmPage,
});
