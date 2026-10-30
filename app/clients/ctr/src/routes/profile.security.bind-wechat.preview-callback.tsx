import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { noindexHead } from "#/domains/seo";
import { BindWechatPreviewCallbackPage } from "#/features/modules/security/bind-wechat-preview-callback";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/profile/security/bind-wechat/preview-callback")({
  validateSearch: (search: Record<string, unknown>): { code?: string; state?: string } => ({
    ...(typeof search.code === "string" ? { code: search.code } : {}),
    ...(typeof search.state === "string" ? { state: search.state } : {}),
  }),
  head: () => ({
    meta: [...noindexHead().meta, { title: pageTitle(m.security_bind_wechat_pending()) }],
  }),
  component: BindWechatPreviewCallbackPage,
});
