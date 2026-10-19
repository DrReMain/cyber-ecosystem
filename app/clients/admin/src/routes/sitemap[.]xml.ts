import { createFileRoute } from "@tanstack/react-router";
import { SEO_MODE, SITEMAP_ROUTES } from "#/config";
import { generateSitemap } from "#/domains/seo";
import { getSiteUrl } from "#/env";
import type { Locale } from "#/paraglide/runtime";
import { baseLocale, locales, localizeUrl } from "#/paraglide/runtime";

export const Route = createFileRoute("/sitemap.xml")({
  server: {
    handlers: {
      // A site that opts out of indexing has nothing to list - 404 keeps the
      // endpoint honest until SEO_MODE flips to "public".
      GET: async () =>
        SEO_MODE === "private"
          ? new Response("Not Found", { status: 404 })
          : new Response(
              generateSitemap({
                siteUrl: await getSiteUrl(),
                routes: SITEMAP_ROUTES,
                locales,
                baseLocale,
                // The domain stays locale-agnostic (string); the adapter
                // narrows to paraglide's Locale union here at the composition
                // root, where paraglide is a legitimate dependency.
                localize: (absoluteUrl, locale) =>
                  localizeUrl(absoluteUrl, { locale: locale as Locale }).href,
              }),
              {
                headers: {
                  "Content-Type": "application/xml; charset=utf-8",
                  "Cache-Control": "public, max-age=3600",
                },
              },
            ),
    },
  },
});
