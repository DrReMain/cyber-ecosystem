import { createFileRoute } from "@tanstack/react-router";
import { z } from "zod";
import { APP_NAME, HOME_PATH } from "#/config";
import { LocaleSelect } from "#/domains/i18n";
import { noindexHead } from "#/domains/seo";
import { ErrorPage } from "#/features/app";
import { DashboardLayout, DashboardNotFound } from "#/features/layout-dashboard";
import { ErrorLogoutButton } from "#/features/layout-dashboard/auth/error-logout-button";
import { sessionGuard } from "#/features/layout-dashboard/auth/guard";

export const Route = createFileRoute("/dashboard")({
  head: noindexHead,
  validateSearch: z.looseObject({}),
  beforeLoad: sessionGuard,
  component: DashboardLayout,
  notFoundComponent: DashboardNotFound,
  errorComponent: ({ error }) => (
    <ErrorPage
      appName={APP_NAME}
      error={error}
      extraBtn={<ErrorLogoutButton />}
      headerTrailing={<LocaleSelect />}
      homePath={HOME_PATH}
    />
  ),
});
