import { createFileRoute } from "@tanstack/react-router";
import { DISALLOW_PREFIXES, SEO_MODE } from "#/config";
import { generateRobotsTxt } from "#/domains/seo";
import { getSiteUrl } from "#/env";
import type { Locale } from "#/paraglide/runtime";
import { locales, localizeUrl } from "#/paraglide/runtime";

export const Route = createFileRoute("/robots.txt")({
  server: {
    handlers: {
      GET: async () =>
        new Response(
          generateRobotsTxt({
            mode: SEO_MODE,
            siteUrl: await getSiteUrl(),
            disallowPrefixes: DISALLOW_PREFIXES,
            locales,
            localize: (absoluteUrl, locale) =>
              localizeUrl(absoluteUrl, { locale: locale as Locale }).href,
          }),
          {
            headers: {
              "Content-Type": "text/plain; charset=utf-8",
              "Cache-Control": "public, max-age=86400",
            },
          },
        ),
    },
  },
});
