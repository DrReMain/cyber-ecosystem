import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { noindexHead } from "#/domains/seo";
import { BindWechatPage } from "#/features/modules/security/bind-wechat";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/profile/security/bind-wechat/")({
  validateSearch: (
    search: Record<string, unknown>,
  ): { failed?: boolean; from?: "me" | "login" | "guide" } => ({
    ...(search.failed === true ? { failed: true } : {}),
    ...(search.from === "login" || search.from === "guide" ? { from: search.from } : {}),
  }),
  head: () => ({
    meta: [...noindexHead().meta, { title: pageTitle(m.bind_wechat_title()) }],
  }),
  component: BindWechatPage,
});
