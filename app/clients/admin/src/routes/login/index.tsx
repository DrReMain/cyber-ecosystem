import { createFileRoute, redirect } from "@tanstack/react-router";
import { z } from "zod";
import { pageTitle } from "#/config";
import { AREA_PATH, LOGIN_PATH } from "#/features/layout-dashboard/area";
import { sessionQuery } from "#/features/layout-dashboard/auth/session.fn";
import { LoginPage } from "#/features/layout-dashboard/login/login-page";
import { safeRedirect } from "#/features/login-default";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/login/")({
  head: () => ({
    meta: [
      { title: pageTitle(m.login_default_title()) },
      { name: "description", content: m.login_default_description() },
    ],
  }),
  validateSearch: z.object({
    redirect: z.string().optional(),
    expired: z.boolean().optional(),
  }),
  beforeLoad: async ({ context, search }) => {
    if (search.expired) return;
    const session = await context.queryClient.query(sessionQuery);
    if (session.status === "authed") {
      throw redirect({
        href: safeRedirect(search.redirect, { home: AREA_PATH, self: LOGIN_PATH }),
      });
    }
  },
  component: LoginPage,
});
