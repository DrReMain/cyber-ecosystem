import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { noindexHead } from "#/domains/seo";
import { BindWechatCallbackPage } from "#/features/modules/security/bind-wechat-callback";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/profile/security/bind-wechat/callback")({
  validateSearch: (search: Record<string, unknown>): { code?: string; state?: string } => ({
    ...(typeof search.code === "string" ? { code: search.code } : {}),
    ...(typeof search.state === "string" ? { state: search.state } : {}),
  }),
  head: () => ({
    meta: [...noindexHead().meta, { title: pageTitle(m.security_bind_wechat_pending()) }],
  }),
  component: BindWechatCallbackPage,
});
