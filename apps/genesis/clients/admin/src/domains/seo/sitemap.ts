import { getSiteUrl } from "#/env"
import { baseLocale, locales, localizeUrl } from "#/paraglide/runtime"
import type { SitemapRoute } from "./routes"
import { sitemapRoutes } from "./routes"

function escapeXml(str: string): string {
  return str
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&apos;")
}

function buildAlternateRefs(routePath: string): string {
  return locales
    .map((locale) => {
      const localized = localizeUrl(`${getSiteUrl()}${routePath}`, { locale })
      return `    <xhtml:link rel="alternate" hreflang="${locale}" href="${escapeXml(localized.href)}" />`
    })
    .join("\n")
}

function buildUrlEntry(route: SitemapRoute): string {
  const defaultLocaleUrl = localizeUrl(`${getSiteUrl()}${route.path}`, {
    locale: baseLocale,
  })
  const lines = [
    "  <url>",
    `    <loc>${escapeXml(defaultLocaleUrl.href)}</loc>`,
    buildAlternateRefs(route.path),
  ]
  if (route.changefreq)
    lines.push(`    <changefreq>${route.changefreq}</changefreq>`)
  if (route.priority !== undefined)
    lines.push(`    <priority>${route.priority}</priority>`)
  lines.push("  </url>")
  return lines.join("\n")
}

export function generateSitemap(): string {
  const urls = sitemapRoutes.map(buildUrlEntry).join("\n")
  return [
    '<?xml version="1.0" encoding="UTF-8"?>',
    '<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"',
    '        xmlns:xhtml="http://www.w3.org/1999/xhtml">',
    urls,
    "</urlset>",
    "",
  ].join("\n")
}
