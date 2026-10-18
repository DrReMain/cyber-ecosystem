export interface SitemapRoute {
  path: string;
  changefreq?: "always" | "hourly" | "daily" | "weekly" | "monthly" | "yearly" | "never";
  priority?: number;
  /** ISO 8601 date, e.g. "2026-07-20" */
  lastmod?: string;
}

export interface SitemapConfig {
  /** Absolute site origin without a trailing slash. */
  siteUrl: string;
  routes: readonly SitemapRoute[];
  /** All supported locales; each becomes an hreflang alternate. */
  locales: readonly string[];
  /** The locale served without a URL prefix; also the hreflang x-default. */
  baseLocale: string;
  /** Maps an absolute URL to its localized counterpart. */
  localize: (absoluteUrl: string, locale: string) => string;
}

function escapeXml(str: string): string {
  return str
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&apos;");
}

function alternateRef(locale: string, href: string): string {
  return `    <xhtml:link rel="alternate" hreflang="${locale}" href="${escapeXml(href)}" />`;
}

function buildAlternateRefs(config: SitemapConfig, routePath: string): string {
  const absolute = `${config.siteUrl}${routePath}`;
  const lines = config.locales.map((locale) =>
    alternateRef(locale, config.localize(absolute, locale)),
  );
  // x-default tells crawlers which version to serve for unmatched languages.
  lines.push(alternateRef("x-default", config.localize(absolute, config.baseLocale)));
  return lines.join("\n");
}

function buildUrlEntry(config: SitemapConfig, route: SitemapRoute): string {
  const defaultLocaleUrl = config.localize(`${config.siteUrl}${route.path}`, config.baseLocale);
  const lines = ["  <url>", `    <loc>${escapeXml(defaultLocaleUrl)}</loc>`];
  if (route.lastmod) lines.push(`    <lastmod>${route.lastmod}</lastmod>`);
  if (route.changefreq) lines.push(`    <changefreq>${route.changefreq}</changefreq>`);
  if (route.priority !== undefined)
    lines.push(`    <priority>${route.priority.toFixed(1)}</priority>`);
  lines.push(buildAlternateRefs(config, route.path));
  lines.push("  </url>");
  return lines.join("\n");
}

export function generateSitemap(config: SitemapConfig): string {
  const urls = config.routes.map((route) => buildUrlEntry(config, route)).join("\n");
  return [
    '<?xml version="1.0" encoding="UTF-8"?>',
    '<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"',
    '        xmlns:xhtml="http://www.w3.org/1999/xhtml">',
    urls,
    "</urlset>",
    "",
  ].join("\n");
}
