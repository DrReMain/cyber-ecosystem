import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { GenrePickPage } from "#/features/modules/match/genre-pick";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/match/pick")({
  validateSearch: (search: Record<string, unknown>): { why?: "tie" | "empty" } => ({
    ...(search.why === "tie" || search.why === "empty" ? { why: search.why } : {}),
  }),
  head: () => ({
    meta: [{ title: pageTitle(m.match_pick_title()) }],
  }),
  component: GenrePickPage,
});
