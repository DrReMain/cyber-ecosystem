import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { SetupGamesPage } from "#/features/modules/profile/setup/games";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/profile/setup/games")({
  head: () => ({
    meta: [{ title: pageTitle(m.profile_setup_games_head()) }],
  }),
  component: SetupGamesPage,
});
