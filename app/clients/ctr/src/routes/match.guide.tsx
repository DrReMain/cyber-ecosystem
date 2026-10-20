import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { TestGuidePage } from "#/features/modules/match/test-guide";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/match/guide")({
  head: () => ({
    meta: [{ title: pageTitle(m.match_guide_title()) }],
  }),
  component: TestGuidePage,
});
