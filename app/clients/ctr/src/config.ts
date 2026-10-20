import type { SitemapRoute } from "#/domains/seo";

// ─── App constants ──────────────────────────────────────────────────────────

export const APP_NAME = "CYBER ESPORTS" as const;
export const HOME_PATH = "/" as const;

// ─── COMMON helper ────────────────────────────────────────────────────────────────────
// Router-level titles are plain strings; this keeps the "Page · App" convention in one place.
export function pageTitle(page: string): string {
  return `${page} · ${APP_NAME}`;
}

// ─── SEO ────────────────────────────────────────────────────────────────────
// The C-end site is public: indexable landing faces plus robots/noindex on
// every personal flow page.
//
// The site's indexing stance, all in one place. Business iteration changes
// this file only - the seo domain stays untouched.
//
// "private": the whole site opts out of indexing (/sitemap.xml answers 404).
// "public":  robots allows everything except DISALLOW_PREFIXES and references
//            the sitemap, which lists SITEMAP_ROUTES with hreflang alternates.

export const SEO_MODE: "public" | "private" = "public";

/** Indexable routes fed into the sitemap (public mode only). */
export const SITEMAP_ROUTES: readonly SitemapRoute[] = [
  { path: HOME_PATH, changefreq: "weekly", priority: 1.0 },
  // The public legal face (H2): crawled yearly, low priority.
  { path: "/agreement", changefreq: "yearly", priority: 0.3 },
];

/**
 * Path prefixes robots.txt disallows in every locale form (the generator
 * localizes each prefix). Personal flow pages never surface in search.
 */
export const DISALLOW_PREFIXES: readonly string[] = ["/login", "/profile", "/age-required"];
