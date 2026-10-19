import type { SitemapRoute } from "#/domains/seo";

// ─── App constants ──────────────────────────────────────────────────────────

export const APP_NAME = "CYBER ECOSYSTEM" as const;
export const HOME_PATH = "/" as const;

// ─── COMMON helper ────────────────────────────────────────────────────────────────────
// Router-level titles are plain strings; this keeps the "Page · App" convention in one place.
export function pageTitle(page: string): string {
  return `${page} · ${APP_NAME}`;
}

// ─── Areas ──────────────────────────────────────────────────────────────────

export interface AreaWiring {
  login: string;
  home: string;
}

export const AREAS = {
  dashboard: { login: "/login", home: "/dashboard" },
} as const satisfies Readonly<Record<string, AreaWiring>>;

// ─── SEO ────────────────────────────────────────────────────────────────────
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
];

/** Path prefixes robots.txt disallows (public mode only). */
export const DISALLOW_PREFIXES: readonly string[] = [
  ...new Set(Object.values(AREAS).flatMap((area) => [area.login, area.home])),
];
