export interface SitemapRoute {
  path: string;
  changefreq?: "always" | "hourly" | "daily" | "weekly" | "monthly" | "yearly" | "never";
  priority?: number;
  /** ISO 8601 date, e.g. "2026-07-20" */
  lastmod?: string;
}

export const sitemapRoutes: SitemapRoute[] = [
  { path: "/", changefreq: "weekly", priority: 1.0 },
  { path: "/theme", changefreq: "monthly", priority: 0.7 },
  { path: "/stores", changefreq: "monthly", priority: 0.7 },
];
