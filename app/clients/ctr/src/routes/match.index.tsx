import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { MatchPage } from "#/features/modules/match/page";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/match/")({
  validateSearch: (search: Record<string, unknown>): { genre?: "fps" | "moba" } => ({
    ...(search.genre === "fps" || search.genre === "moba" ? { genre: search.genre } : {}),
  }),
  head: () => ({
    meta: [{ title: pageTitle(m.match_title()) }],
  }),
  component: MatchPage,
});
