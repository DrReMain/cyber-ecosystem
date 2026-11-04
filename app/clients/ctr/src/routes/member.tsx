import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { noindexHead } from "#/domains/seo";
import { sessionGuard } from "#/features/auth/guard";
import { MemberDetailPage } from "#/features/modules/member/page";
import { m } from "#/paraglide/messages";

// The membership benefits page (prototype member-detail): a login-guarded
// standalone — the profile card's detail link, the wall's renewal link,
// and the receipt's member button all land here.
export const Route = createFileRoute("/member")({
  beforeLoad: sessionGuard,
  head: () => ({ meta: [...noindexHead().meta, { title: pageTitle(m.member_detail_head()) }] }),
  component: MemberDetailPage,
});
