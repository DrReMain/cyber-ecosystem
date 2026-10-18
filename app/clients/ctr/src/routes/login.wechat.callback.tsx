import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { noindexHead } from "#/domains/seo";
import { WechatCallbackPage } from "#/features/modules/login/wechat-callback";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/login/wechat/callback")({
  validateSearch: (
    search: Record<string, unknown>,
  ): { code?: string; state?: string; next?: string } => ({
    ...(typeof search.code === "string" ? { code: search.code } : {}),
    ...(typeof search.state === "string" ? { state: search.state } : {}),
    ...(typeof search.next === "string" ? { next: search.next } : {}),
  }),
  head: () => ({
    meta: [...noindexHead().meta, { title: pageTitle(m.login_callback_pending()) }],
  }),
  component: WechatCallbackPage,
});
